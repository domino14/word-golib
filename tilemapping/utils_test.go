package tilemapping

import (
	"sort"
	"testing"

	"github.com/matryer/is"
)

func TestLeaveBlankExchange(t *testing.T) {
	// Test blank exchange/removal works (zeroIsPlaythrough=false means remove blanks)
	is := is.New(t)
	l, err := Leave([]MachineLetter{0, 3, 5, 6},
		[]MachineLetter{0}, false)

	is.NoErr(err)
	is.Equal(l, MachineWord{3, 5, 6})
}

func TestLeaveTilePlay(t *testing.T) {
	// Playing on board with designated blank (zeroIsPlaythrough=true)
	is := is.New(t)
	l, err := Leave([]MachineLetter{0, 3, 5, 6, 7, 9},
		[]MachineLetter{5, 0x86, 6}, true)

	is.NoErr(err)
	is.Equal(l, MachineWord{3, 7, 9})
}

func TestLeaveTilePlayTooManyBlanks(t *testing.T) {
	// Playing with too many designated blanks (zeroIsPlaythrough=true)
	is := is.New(t)
	_, err := Leave([]MachineLetter{0, 3, 5, 6, 7, 9},
		[]MachineLetter{5, 0x86, 0x88}, true)

	is.Equal(err.Error(), "tile 0 not in rack")
}

func TestLeaveTilePlayRepeat(t *testing.T) {
	// Playing with designated blank and duplicate tiles (zeroIsPlaythrough=true)
	is := is.New(t)
	l, err := Leave([]MachineLetter{0, 3, 5, 5, 6, 7, 9},
		[]MachineLetter{5, 0x86, 6}, true)

	is.NoErr(err)
	is.Equal(l, MachineWord{3, 5, 7, 9})
}

func TestLeaveTilePlayWithThroughLetters(t *testing.T) {
	// Playing with play-through markers (tile 0) (zeroIsPlaythrough=true)
	is := is.New(t)
	l, err := Leave([]MachineLetter{0, 3, 5, 6, 7, 9, 15},
		[]MachineLetter{5, 0, 0, 15, 6, 0, 7, 0}, true)

	is.NoErr(err)
	is.Equal(l, MachineWord{0, 3, 9})
}

func TestSortMW(t *testing.T) {
	is := is.New(t)
	mw := []MachineLetter{3, 7, 1, 0, 4, 9}
	SortMW(mw)

	is.Equal(mw, []MachineLetter{0, 1, 3, 4, 7, 9})

}

func BenchmarkSortMW(b *testing.B) {
	// 3.779 ns
	for i := 0; i < b.N; i++ {
		mw := []MachineLetter{3, 7, 1, 0, 4, 9}
		SortMW(mw)
	}
}

func BenchmarkSortMWBuiltin(b *testing.B) {
	// 116 ns
	for i := 0; i < b.N; i++ {
		mw := []MachineLetter{3, 7, 1, 0, 4, 9}
		sort.Slice(mw, func(i, j int) bool {
			return mw[i] < mw[j]
		})
	}
}
