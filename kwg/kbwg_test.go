package kwg

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanKBWG(t *testing.T) {
	// Create a simple KWG with a few nodes
	nodes := []uint32{
		0x01000000, // Node 0: Tile 1, no arc
		0x02800000, // Node 1: Tile 2, accepts, no arc
		0x03400000, // Node 2: Tile 3, is end, no arc
		0x04C00000, // Node 3: Tile 4, accepts and is end, no arc
		0x05000001, // Node 4: Tile 5, arc to node 1
	}

	// Convert nodes to bytes
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, nodes)
	assert.NoError(t, err)

	// Scan the bytes as a KWG
	kwgBuf := bytes.NewReader(buf.Bytes())
	kwg, err := ScanKWG(kwgBuf, len(nodes)*4)
	assert.NoError(t, err)
	assert.NotNil(t, kwg)
	assert.Equal(t, len(nodes), len(kwg.nodes))

	// Scan the bytes as a KBWG
	kbwgBuf := bytes.NewReader(buf.Bytes())
	kbwg, err := ScanKBWG(kbwgBuf, len(nodes)*4)
	assert.NoError(t, err)
	assert.NotNil(t, kbwg)
	assert.Equal(t, len(nodes), len(kbwg.nodes))

	// Test that KWG and KBWG methods work correctly
	assert.Equal(t, uint8(1), kwg.Tile(0))
	assert.Equal(t, uint8(1)&0x3f, kbwg.Tile(0)) // KBWG uses 6 bits for tile

	assert.Equal(t, uint32(0), kwg.ArcIndex(0))
	assert.Equal(t, uint32(0), kbwg.ArcIndex(0))

	assert.Equal(t, uint32(1), kwg.ArcIndex(4))
	assert.Equal(t, uint32(1), kbwg.ArcIndex(4))

	assert.False(t, kwg.IsEnd(0))
	assert.False(t, kbwg.IsEnd(0))

	assert.True(t, kwg.IsEnd(2))
	assert.True(t, kbwg.IsEnd(2))

	assert.False(t, kwg.Accepts(0))
	assert.False(t, kbwg.Accepts(0))

	assert.True(t, kwg.Accepts(1))
	assert.True(t, kbwg.Accepts(1))
}

func TestWordGraphInterface(t *testing.T) {
	// Create a simple KWG with a few nodes
	nodes := []uint32{
		0x01000000, // Node 0: Tile 1, no arc
		0x02800000, // Node 1: Tile 2, accepts, no arc
		0x03400000, // Node 2: Tile 3, is end, no arc
		0x04C00000, // Node 3: Tile 4, accepts and is end, no arc
		0x05000001, // Node 4: Tile 5, arc to node 1
	}

	// Convert nodes to bytes
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, nodes)
	assert.NoError(t, err)

	// Scan the bytes as a KWG and KBWG
	kwgBuf := bytes.NewReader(buf.Bytes())
	kwg, err := ScanKWG(kwgBuf, len(nodes)*4)
	assert.NoError(t, err)

	kbwgBuf := bytes.NewReader(buf.Bytes())
	kbwg, err := ScanKBWG(kbwgBuf, len(nodes)*4)
	assert.NoError(t, err)

	// Test that both KWG and KBWG implement the WordGraph interface
	var wg1, wg2 WordGraph
	wg1 = kwg
	wg2 = kbwg

	// Test interface methods
	assert.Equal(t, uint32(0), wg1.ArcIndex(0))
	assert.Equal(t, uint32(0), wg2.ArcIndex(0))

	assert.Equal(t, uint32(1), wg1.ArcIndex(4))
	assert.Equal(t, uint32(1), wg2.ArcIndex(4))

	assert.False(t, wg1.IsEnd(0))
	assert.False(t, wg2.IsEnd(0))

	assert.True(t, wg1.IsEnd(2))
	assert.True(t, wg2.IsEnd(2))

	assert.False(t, wg1.Accepts(0))
	assert.False(t, wg2.Accepts(0))

	assert.True(t, wg1.Accepts(1))
	assert.True(t, wg2.Accepts(1))

	// Test that the Tile method returns different values for KWG and KBWG
	assert.Equal(t, uint8(1), wg1.Tile(0))
	assert.Equal(t, uint8(1)&0x3f, wg2.Tile(0))
}
