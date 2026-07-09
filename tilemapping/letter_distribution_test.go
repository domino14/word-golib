package tilemapping

import (
	"fmt"
	"testing"

	"github.com/domino14/word-golib/config"
	"github.com/matryer/is"
)

func TestLetterDistributionScores(t *testing.T) {
	is := is.New(t)
	ld, err := EnglishLetterDistribution(config.DefaultConfig)
	is.NoErr(err)

	is.Equal(ld.Score(0), 0)
	is.Equal(ld.Score(0x81), 0)
	is.Equal(ld.Score(25), 4)
	is.Equal(ld.Score(26), 10)
	is.Equal(ld.Score(8), 4)
	is.Equal(ld.Score(1), 1)
}

func TestLetterDistributionWordScore(t *testing.T) {
	is := is.New(t)
	ld, err := EnglishLetterDistribution(config.DefaultConfig)
	is.NoErr(err)

	word := "CoOKIE"
	mls, err := ToMachineLetters(word, ld.TileMapping())
	fmt.Println("mls", mls)
	is.NoErr(err)
	is.Equal(ld.WordScore(mls), 11)
}

func TestProbableLetterDistributionName(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		lexname string
		want    string
	}{
		{"SLV26", "slovene"},
		{"slv26", "slovene"},
		{"NWL23", "english"},
		{"CSW24", "english"},
		{"OSPS50", "polish"},
	}
	for _, c := range cases {
		got, err := ProbableLetterDistributionName(c.lexname)
		is.NoErr(err)
		is.Equal(got, c.want)
	}

	_, err := ProbableLetterDistributionName("WOW24")
	is.True(err != nil)
}
