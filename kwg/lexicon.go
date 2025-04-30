package kwg

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/domino14/word-golib/tilemapping"
)

type Lexicon[T WordGraphConstraint] struct {
	WordGraph T
}

func (l Lexicon[T]) Name() string {
	return l.WordGraph.LexiconName()
}

func (l Lexicon[T]) HasWord(word tilemapping.MachineWord) bool {
	return FindMachineWord(l.WordGraph, word)
}

// DaPools is a map of sync.Pools for KWGAnagrammer instances, keyed by type.
var DaPools = sync.Map{}

func getDaPool[T WordGraphConstraint]() *sync.Pool {
	var zero T
	pool, ok := DaPools.Load(fmt.Sprintf("%T", zero))
	if !ok {
		pool, _ = DaPools.LoadOrStore(fmt.Sprintf("%T", zero), &sync.Pool{
			New: func() interface{} {
				return &KWGAnagrammer[T]{}
			},
		})
	}
	return pool.(*sync.Pool)
}

func (l Lexicon[T]) HasAnagram(word tilemapping.MachineWord) bool {
	log.Debug().Str("word", word.UserVisible(l.WordGraph.GetAlphabet())).Msg("has-anagram?")

	pool := getDaPool[T]()
	da := pool.Get().(*KWGAnagrammer[T])
	defer pool.Put(da)

	v, err := da.IsValidJumble(l.WordGraph, word)
	if err != nil {
		log.Err(err).Str("word", word.UserVisible(l.WordGraph.GetAlphabet())).Msg("has-anagram?-error")
		return false
	}

	return v
}
