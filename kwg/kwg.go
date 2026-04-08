package kwg

import (
	"encoding/binary"
	"io"

	"github.com/domino14/word-golib/tilemapping"
	"github.com/rs/zerolog/log"
)

// A KWG is a Kurnia Word Graph. More information is available here:
// https://github.com/andy-k/wolges/blob/main/details.txt
// Thanks to Andy Kurnia.
type KWG struct {
	// Nodes is just a slice of 32-bit elements, the node array.
	nodes       []uint32
	alphabet    *tilemapping.TileMapping
	lexiconName string
	wordCounts  []int32
}

func ScanKWG(data io.Reader, filesize int) (*KWG, error) {
	nodes := make([]uint32, filesize/4)
	err := binary.Read(data, binary.LittleEndian, nodes)
	if err != nil {
		return nil, err
	}

	log.Debug().Int("num-nodes", len(nodes)).Msg("loaded-kwg")
	return &KWG{nodes: nodes}, nil
}

func (k *KWG) GetRootNodeIndex() uint32 {
	return k.ArcIndex(1) // (1) for a GADDAG, (0) for a DAWG
}

func (k KWG) GetAlphabet() *tilemapping.TileMapping {
	return k.alphabet
}

func (k *KWG) LexiconName() string {
	return k.lexiconName
}

func (k *KWG) NextNodeIdx(nodeIdx uint32, letter tilemapping.MachineLetter) uint32 {
	for i := nodeIdx; ; i++ {
		if k.Tile(i) == uint8(letter) {
			return k.ArcIndex(i)
		}
		if k.IsEnd(i) {
			return 0
		}
	}
}

func (k *KWG) InLetterSet(letter tilemapping.MachineLetter, nodeIdx uint32) bool {
	letter = letter.Unblank()
	for i := nodeIdx; ; i++ {
		if k.Tile(i) == uint8(letter) {
			return k.Accepts(i)
		}
		if k.IsEnd(i) {
			return false
		}
	}
}

func (k *KWG) GetLetterSet(nodeIdx uint32) tilemapping.LetterSet {
	var ls tilemapping.LetterSet
	for i := nodeIdx; ; i++ {
		t := k.Tile(i)
		if k.Accepts(i) {
			ls |= (1 << t)
		}
		if k.IsEnd(i) {
			break
		}
	}
	return ls
}

func (k *KWG) IsEnd(nodeIdx uint32) bool {
	return k.nodes[nodeIdx]&0x400000 != 0
}

func (k *KWG) Accepts(nodeIdx uint32) bool {
	return k.nodes[nodeIdx]&0x800000 != 0
}

func (k *KWG) ArcIndex(nodeIdx uint32) uint32 {
	return k.nodes[nodeIdx] & 0x3fffff
}

func (k *KWG) Tile(nodeIdx uint32) uint8 {
	return uint8(k.nodes[nodeIdx] >> 24)
}

// KWG node bit-field layout. Each node is a uint32:
//
//	bits  0..21  arc index (the link to the child node row)
//	bit   22     end-of-arc-list flag (1 = last node in this arc list)
//	bit   23     accepts flag (1 = path through this node spells a word)
//	bits 24..31  tile letter (1..MAX_ALPHABET; 0 is unused)
//
// These constants are exported so a caller that has obtained the
// underlying nodes slice via [KWG.Nodes] can decode each node
// directly without having to call Tile/IsEnd/Accepts/ArcIndex four
// separate times — useful in tight move-generator inner loops where
// the same node value is used to read several bit fields back to
// back.
//
// IMPORTANT: these constants describe the KWG layout only. [KBWG]
// (the big word graph for huge alphabets) uses a different bit
// layout — see KBWG's accessor methods. Don't use these constants
// on a KBWG node value.
const (
	KWGNodeArcMask    uint32 = 0x3fffff
	KWGNodeIsEndBit   uint32 = 0x400000
	KWGNodeAcceptsBit uint32 = 0x800000
	KWGNodeTileShift  uint32 = 24
)

// Nodes returns the underlying node array of this KWG. The returned
// slice aliases internal state — callers must treat it as read-only.
//
// This is the fast path for tight inner loops that need to read
// multiple bit fields out of the same node (Tile + IsEnd + Accepts +
// ArcIndex). Without this accessor each of those four reads goes
// through a separate method call and bounds-checked slice access
// into the unexported k.nodes field; with the slice in hand the
// caller can hoist the load and extract the bit fields locally:
//
//	nodes := k.Nodes()
//	for i := startIdx; ; i++ {
//	    node := nodes[i]
//	    tile := uint8(node >> kwg.KWGNodeTileShift)
//	    if tile == target {
//	        arc := node & kwg.KWGNodeArcMask
//	        accepts := node&kwg.KWGNodeAcceptsBit != 0
//	        // ...
//	        break
//	    }
//	    if node&kwg.KWGNodeIsEndBit != 0 {
//	        break
//	    }
//	}
//
// In a move-generator profile, KWG.Tile and KWG.IsEnd combined for
// roughly 18% of total CPU when called four-method-style; switching
// to the hoisted-slice pattern collapses those entries and lets the
// compiler keep the inner loop in registers.
func (k *KWG) Nodes() []uint32 {
	return k.nodes
}

// I have no idea what is going on in these functions. See wolges kwg.rs
func (k *KWG) countWordsAt(p uint32) int {
	if p >= uint32(len(k.wordCounts)) {
		return 0
	}
	if k.wordCounts[p] == -1 {
		panic("unexpected -1")
	}
	if k.wordCounts[p] == 0 {
		k.wordCounts[p] = -1

		a := 0
		if k.Accepts(p) {
			a = 1
		}
		b := 0
		if k.ArcIndex(p) != 0 {
			b = k.countWordsAt(k.ArcIndex(p))
		}
		c := 0
		if !k.IsEnd(p) {
			c = k.countWordsAt(p + 1)
		}
		k.wordCounts[p] = int32(a + b + c)
	}
	return int(k.wordCounts[p])
}

// WordCountAt returns the number of words in the subtree rooted at nodeIdx.
// The caller must ensure CountWords has been called before using this method.
func (k *KWG) WordCountAt(nodeIdx uint32) int32 {
	return k.wordCounts[nodeIdx]
}

func (k *KWG) CountWords() {
	k.wordCounts = make([]int32, len(k.nodes))
	for p := len(k.wordCounts) - 1; p >= 0; p-- {
		k.countWordsAt(uint32(p))
	}
}

func (k *KWG) GetWordIndexOf(nodeIdx uint32, letters tilemapping.MachineWord) int32 {
	idx := int32(0)
	lidx := 0

	for nodeIdx != 0 {
		idx += k.wordCounts[nodeIdx]
		for k.Tile(nodeIdx) != uint8(letters[lidx]) {
			if k.IsEnd(nodeIdx) {
				return -1
			}
			nodeIdx++
		}
		idx -= k.wordCounts[nodeIdx]
		lidx++
		if lidx > len(letters)-1 {
			if k.Accepts(nodeIdx) {
				return int32(idx)
			}
			return -1
		}
		if k.Accepts(nodeIdx) {
			idx += 1
		}
		nodeIdx = k.ArcIndex(nodeIdx)
	}
	return -1
}

// KBWG is a "Big Word Graph" that uses 24 instead of 22 bits for the pointer.
// All accessor functions (Tile, ArcIndex, IsEnd, Accepts) must be overridden
// LSB to MSB:
// typedef struct { uint8_t c : 6; bool e : 1, d : 1; uint32_t p : 24; } Kwg2NodeE;

type KBWG struct {
	KWG
}

// Override the Tile method for KBWG
func (k *KBWG) Tile(nodeIdx uint32) uint8 {
	return uint8(k.nodes[nodeIdx]) & 0x3f
}

// Override the ArcIndex method for KBWG
func (k *KBWG) ArcIndex(nodeIdx uint32) uint32 {
	return k.nodes[nodeIdx] >> 8
}

func (k *KBWG) IsEnd(nodeIdx uint32) bool {
	return k.nodes[nodeIdx]&0x40 != 0
}

func (k *KBWG) Accepts(nodeIdx uint32) bool {
	return k.nodes[nodeIdx]&0x80 != 0
}

// ScanKBWG scans a KBWG from a reader
func ScanKBWG(data io.Reader, filesize int) (*KBWG, error) {
	kwg, err := ScanKWG(data, filesize)
	if err != nil {
		return nil, err
	}
	return &KBWG{KWG: *kwg}, nil
}
