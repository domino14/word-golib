package kwg

import (
	"errors"
	"fmt"

	"github.com/domino14/word-golib/tilemapping"
)

// zero value works. not threadsafe.
type KWGAnagrammer struct {
	ans         tilemapping.MachineWord
	freq        []uint8
	blanks      uint8
	queryLength int
}

func (da *KWGAnagrammer) commonInit(kwg WordGraph) {
	alph := kwg.GetAlphabet()
	numLetters := alph.NumLetters()
	if cap(da.freq) < int(numLetters) {
		da.freq = make([]uint8, numLetters)
	} else {
		da.freq = da.freq[:numLetters]
		for i := range da.freq {
			da.freq[i] = 0
		}
	}
	da.blanks = 0
	da.ans = da.ans[:0]
}

func (da *KWGAnagrammer) InitForString(kwg WordGraph, tiles string) error {
	da.commonInit(kwg)
	da.queryLength = 0
	alph := kwg.GetAlphabet()

	mls, err := tilemapping.ToMachineLetters(tiles, alph)
	if err != nil {
		return err
	}
	return da.InitForMachineWord(kwg, mls)
}

func (da *KWGAnagrammer) InitForMachineWord(kwg WordGraph, machineTiles tilemapping.MachineWord) error {
	da.commonInit(kwg)
	da.queryLength = len(machineTiles)
	alph := kwg.GetAlphabet()
	numLetters := alph.NumLetters()
	for _, v := range machineTiles {
		if v == 0 {
			da.blanks++
		} else if uint8(v) < numLetters {
			da.freq[v]++
		} else {
			return fmt.Errorf("invalid byte %v", v)
		}
	}
	return nil
}

// f must not modify the given slice. if f returns error, abort iteration.
func (ka *KWGAnagrammer) iterate(kwg WordGraph, nodeIdx uint32, minLen int, minExact int, f func(tilemapping.MachineWord) error) error {
	for ; ; nodeIdx++ {
		j := kwg.Tile(nodeIdx)
		if ka.freq[j] > 0 {
			ka.freq[j]--
			ka.ans = append(ka.ans, tilemapping.MachineLetter(j))
			if minLen <= 1 && minExact <= 1 && kwg.Accepts(nodeIdx) {
				if err := f(ka.ans); err != nil {
					return err
				}
			}
			if arcIndex := kwg.ArcIndex(nodeIdx); arcIndex != 0 {
				if err := ka.iterate(kwg, arcIndex, minLen-1, minExact-1, f); err != nil {
					return err
				}
			}
			ka.ans = ka.ans[:len(ka.ans)-1]
			ka.freq[j]++
		} else if ka.blanks > 0 {
			ka.blanks--
			ka.ans = append(ka.ans, tilemapping.MachineLetter(j))
			if minLen <= 1 && minExact <= 0 && kwg.Accepts(nodeIdx) {
				if err := f(ka.ans); err != nil {
					return err
				}
			}
			if arcIndex := kwg.ArcIndex(nodeIdx); arcIndex != 0 {
				if err := ka.iterate(kwg, arcIndex, minLen-1, minExact, f); err != nil {
					return err
				}
			}
			ka.ans = ka.ans[:len(ka.ans)-1]
			ka.blanks++
		}
		if kwg.IsEnd(nodeIdx) {
			return nil
		}
	}
}

func (da *KWGAnagrammer) Anagram(dawg WordGraph, f func(tilemapping.MachineWord) error) error {
	return da.iterate(dawg, dawg.ArcIndex(0), da.queryLength, 0, f)
}

func (da *KWGAnagrammer) Subanagram(dawg WordGraph, f func(tilemapping.MachineWord) error) error {
	return da.iterate(dawg, dawg.ArcIndex(0), 1, 0, f)
}

func (da *KWGAnagrammer) Superanagram(dawg WordGraph, f func(tilemapping.MachineWord) error) error {
	minExact := da.queryLength - int(da.blanks)
	blanks := da.blanks
	da.blanks = 255
	err := da.iterate(dawg, dawg.ArcIndex(0), da.queryLength, minExact, f)
	da.blanks = blanks
	return err
}

var errHasAnagram = errors.New("has anagram")
var errHasBlanks = errors.New("has blanks")

func foundAnagram(tilemapping.MachineWord) error {
	return errHasAnagram
}

// checks if a word with no blanks has any valid anagrams.
func (da *KWGAnagrammer) IsValidJumble(dawg WordGraph, word tilemapping.MachineWord) (bool, error) {
	if err := da.InitForMachineWord(dawg, word); err != nil {
		return false, err
	} else if da.blanks > 0 {
		return false, errHasBlanks
	}
	err := da.Anagram(dawg, foundAnagram)
	if err == nil {
		return false, nil
	} else if err == errHasAnagram {
		return true, nil
	} else {
		return false, err
	}
}
