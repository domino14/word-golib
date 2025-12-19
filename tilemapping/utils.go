package tilemapping

import (
	"fmt"
)

func SortMW(l MachineWord) {
	// sort in place. This should be fast enough for small arrays.
	ll := len(l)
	for i := 1; i < ll; i++ {
		for j := i; j > 0 && l[j-1] > l[j]; j-- {
			l[j-1], l[j] = l[j], l[j-1]
		}
	}
}

// Leave calculates the tiles remaining in the rack after removing the specified tiles.
//
// Parameters:
//   - rack: The current rack tiles
//   - tilesToRemove: The tiles to remove from the rack
//   - zeroIsPlaythrough: If true, tile 0 in tilesToRemove represents a play-through marker
//     (a tile already on the board being played through) and won't be removed from the rack.
//     If false, tile 0 is treated as a regular blank tile and will be removed from the rack.
//
// Returns the remaining tiles (leave) after removal, or an error if the rack doesn't contain
// the tiles to be removed.
func Leave(rack MachineWord, tilesToRemove MachineWord, zeroIsPlaythrough bool) (MachineWord, error) {

	rackletters := map[MachineLetter]int{}
	for _, l := range rack {
		rackletters[l]++
	}
	leave := make([]MachineLetter, 0)

	for _, t := range tilesToRemove {
		if t == 0 && zeroIsPlaythrough {
			// play-through marker - don't remove from rack
			continue
		}
		if t.IsBlanked() {
			if !zeroIsPlaythrough {
				return nil, fmt.Errorf("cannot remove a designated blank from rack (use tile 0 for undesignated blanks)")
			}
			// it's a designated blank on the board, count as undesignated blank in rack
			t = 0
		}
		if rackletters[t] != 0 {
			rackletters[t]--
		} else {
			return nil, fmt.Errorf("tile %v not in rack", t)
		}
	}

	for k, v := range rackletters {
		if v > 0 {
			for i := 0; i < v; i++ {
				leave = append(leave, k)
			}
		}
	}
	SortMW(leave)
	return leave, nil

}
