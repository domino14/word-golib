package kwg

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/domino14/word-golib/tilemapping"
)

type Lexicon struct {
	WordGraph *KWG
}

func (l Lexicon) Name() string {
	return l.WordGraph.LexiconName()
}

func (l Lexicon) HasWord(word tilemapping.MachineWord) bool {
	return FindMachineWord(l.WordGraph, word)
}

// DaPools is a map of sync.Pools for KWGAnagrammer instances, keyed by type.
var DaPools = sync.Map{}

func getDaPool() *sync.Pool {
	var zero *KWGAnagrammer
	pool, ok := DaPools.Load(fmt.Sprintf("%T", zero))
	if !ok {
		pool, _ = DaPools.LoadOrStore(fmt.Sprintf("%T", zero), &sync.Pool{
			New: func() interface{} {
				return &KWGAnagrammer{}
			},
		})
	}
	return pool.(*sync.Pool)
}

func (l Lexicon) HasAnagram(word tilemapping.MachineWord) bool {
	log.Debug().Str("word", word.UserVisible(l.WordGraph.GetAlphabet())).Msg("has-anagram?")

	pool := getDaPool()
	da := pool.Get().(*KWGAnagrammer)
	defer pool.Put(da)

	v, err := da.IsValidJumble(l.WordGraph, word)
	if err != nil {
		log.Err(err).Str("word", word.UserVisible(l.WordGraph.GetAlphabet())).Msg("has-anagram?-error")
		return false
	}

	return v
}
