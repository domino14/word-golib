package kwg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"

	"github.com/domino14/word-golib/config"
	"github.com/domino14/word-golib/tilemapping"
)

// These tests exercise the generic LoadWordGraph[T]/LoadWordGraphWithDistribution[T]
// and GetGraph[T]/GetGraphWithDistribution[T] functions for both of their valid
// instantiations (*KWG and *KBWG), plus the cache-key behavior of the
// WithDistribution variants.
//
// The GlobalObjectCache backing GetKWG/GetKBWG/GetKWGWithDistribution/etc. is a
// process-wide singleton keyed only by lexicon (and, for the WithDistribution
// variants, distribution) name -- NOT by *config.Config or DataPath. So every
// lexicon and distribution name used below is a "zzz"-prefixed fake, guaranteed
// not to collide with any real lexicon/distribution name (or with whatever the
// rest of this package's tests -- which use a real DATA_PATH -- might warm the
// cache with, e.g. "english").

const (
	fakeDistOne        = "zzztestdistone"
	fakeDistOneContent = "?,2,0,0\nA,9,1,1\nB,2,3,0\n"
	fakeDistTwo        = "zzztestdisttwo"
	fakeDistTwoContent = "?,2,0,0\nX,4,8,0\nY,2,4,1\nZ,1,10,0\n"
)

// fakeGraphNodes is a valid (multiple-of-4-bytes) node array. Its content is
// irrelevant: these tests only exercise metadata resolution (lexicon name +
// alphabet assignment), never graph traversal.
var fakeGraphNodes = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	dataPath := t.TempDir()
	is.New(t).NoErr(os.MkdirAll(filepath.Join(dataPath, "letterdistributions"), 0755))
	is.New(t).NoErr(os.MkdirAll(filepath.Join(dataPath, "lexica", "gaddag"), 0755))
	return &config.Config{DataPath: dataPath}
}

func writeFakeDist(t *testing.T, cfg *config.Config, name, content string) {
	t.Helper()
	is.New(t).NoErr(os.WriteFile(filepath.Join(cfg.DataPath, "letterdistributions", name), []byte(content), 0644))
}

// writeFakeLexFile writes a fixture at lexica/gaddag/<lexName><ext>, e.g.
// writeFakeLexFile(t, cfg, "ZZZLEX01", ".kwg").
func writeFakeLexFile(t *testing.T, cfg *config.Config, lexName, ext string) string {
	t.Helper()
	path := filepath.Join(cfg.DataPath, "lexica", "gaddag", lexName+ext)
	is.New(t).NoErr(os.WriteFile(path, fakeGraphNodes, 0644))
	return path
}

func TestLoadWordGraph_KWG_UnrecognizedLexiconName_Errors(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	path := writeFakeLexFile(t, cfg, "ZZZLEX01", ".kwg")

	_, err := LoadWordGraph[*KWG](cfg, path)
	is.True(err != nil)
}

func TestLoadWordGraphWithDistribution_KWG_OverridesUnrecognizedName(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistOne, fakeDistOneContent)
	path := writeFakeLexFile(t, cfg, "ZZZLEX02", ".kwg")

	g, err := LoadWordGraph[*KWG](cfg, path, WithDistribution(fakeDistOne))
	is.NoErr(err)
	is.Equal(g.LexiconName(), "ZZZLEX02")
	is.True(g.GetAlphabet() != nil)

	wantLD, err := tilemapping.NamedLetterDistribution(cfg, fakeDistOne)
	is.NoErr(err)
	// fakeDistOneContent defines 3 distinct tiles (?, A, B).
	is.Equal(g.GetAlphabet().NumLetters(), wantLD.TileMapping().NumLetters())
	is.Equal(g.GetAlphabet().NumLetters(), uint8(3))
}

func TestLoadWordGraphWithDistribution_KWG_EmptyDistNameBehavesLikePlainLoad(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	path := writeFakeLexFile(t, cfg, "ZZZLEX03", ".kwg")

	_, errWithEmptyOverride := LoadWordGraph[*KWG](cfg, path, WithDistribution(""))
	_, errPlain := LoadWordGraph[*KWG](cfg, path)

	is.True(errWithEmptyOverride != nil)
	is.True(errPlain != nil)
	is.Equal(errWithEmptyOverride.Error(), errPlain.Error())
}

func TestLoadWordGraph_KBWG_UnrecognizedLexiconName_Errors(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	path := writeFakeLexFile(t, cfg, "ZZZLEX04", ".kbwg")

	_, err := LoadWordGraph[*KBWG](cfg, path)
	is.True(err != nil)
}

func TestLoadWordGraphWithDistribution_KBWG_OverridesUnrecognizedName(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistTwo, fakeDistTwoContent)
	path := writeFakeLexFile(t, cfg, "ZZZLEX05", ".kbwg")

	g, err := LoadWordGraph[*KBWG](cfg, path, WithDistribution(fakeDistTwo))
	is.NoErr(err)
	is.Equal(g.LexiconName(), "ZZZLEX05")

	wantLD, err := tilemapping.NamedLetterDistribution(cfg, fakeDistTwo)
	is.NoErr(err)
	// fakeDistTwoContent defines 4 distinct tiles (?, X, Y, Z).
	is.Equal(g.GetAlphabet().NumLetters(), wantLD.TileMapping().NumLetters())
	is.Equal(g.GetAlphabet().NumLetters(), uint8(4))
}

func TestGetKWG_UnrecognizedLexiconName_ErrorsWithoutOverride(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeLexFile(t, cfg, "ZZZLEX06", ".kwg")

	_, err := GetKWG(cfg, "ZZZLEX06")
	is.True(err != nil)
}

func TestGetKWGWithDistribution_UnrecognizedLexiconName_Succeeds(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistOne, fakeDistOneContent)
	writeFakeLexFile(t, cfg, "ZZZLEX07", ".kwg")

	g, err := GetKWG(cfg, "ZZZLEX07", WithDistribution(fakeDistOne))
	is.NoErr(err)
	is.Equal(g.LexiconName(), "ZZZLEX07")
}

func TestGetKWGWithDistribution_DifferentDistributionsCachedSeparately(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistOne, fakeDistOneContent)
	writeFakeDist(t, cfg, fakeDistTwo, fakeDistTwoContent)
	writeFakeLexFile(t, cfg, "ZZZLEX08", ".kwg")

	g1, err := GetKWG(cfg, "ZZZLEX08", WithDistribution(fakeDistOne))
	is.NoErr(err)
	g2, err := GetKWG(cfg, "ZZZLEX08", WithDistribution(fakeDistTwo))
	is.NoErr(err)

	// Same lexicon, different distribution overrides: must not share a cache
	// slot, or one distribution's alphabet would silently clobber the other's.
	is.True(g1.GetAlphabet() != g2.GetAlphabet())

	// Re-requesting the same (name, distribution) pair should hit the cache
	// and return the exact same object, not reload from disk.
	g1Again, err := GetKWG(cfg, "ZZZLEX08", WithDistribution(fakeDistOne))
	is.NoErr(err)
	is.True(g1 == g1Again)
}

func TestGetKBWG_UnrecognizedLexiconName_ErrorsWithoutOverride(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeLexFile(t, cfg, "ZZZLEX09", ".kbwg")

	_, err := GetKBWG(cfg, "ZZZLEX09")
	is.True(err != nil)
}

func TestGetKBWGWithDistribution_UnrecognizedLexiconName_Succeeds(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistTwo, fakeDistTwoContent)
	writeFakeLexFile(t, cfg, "ZZZLEX10", ".kbwg")

	g, err := GetKBWG(cfg, "ZZZLEX10", WithDistribution(fakeDistTwo))
	is.NoErr(err)
	is.Equal(g.LexiconName(), "ZZZLEX10")
	is.Equal(g.GetAlphabet().NumLetters(), uint8(4))
}

func TestGetGraph_KWG_UnrecognizedLexiconName_Errors(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeLexFile(t, cfg, "ZZZLEX11", ".kwg")

	_, err := GetGraph[*KWG](cfg, "ZZZLEX11")
	is.True(err != nil)
}

func TestGetGraph_KBWG_UnrecognizedLexiconName_Errors(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeLexFile(t, cfg, "ZZZLEX12", ".kbwg")

	_, err := GetGraph[*KBWG](cfg, "ZZZLEX12")
	is.True(err != nil)
}

func TestGetGraphWithDistribution_KWG_MatchesGetKWGWithDistribution(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistOne, fakeDistOneContent)
	writeFakeLexFile(t, cfg, "ZZZLEX13", ".kwg")

	viaGeneric, err := GetGraph[*KWG](cfg, "ZZZLEX13", WithDistribution(fakeDistOne))
	is.NoErr(err)
	viaSpecific, err := GetKWG(cfg, "ZZZLEX13", WithDistribution(fakeDistOne))
	is.NoErr(err)

	is.True(viaGeneric == viaSpecific)
}

func TestGetGraphWithDistribution_KBWG_MatchesGetKBWGWithDistribution(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeDist(t, cfg, fakeDistTwo, fakeDistTwoContent)
	writeFakeLexFile(t, cfg, "ZZZLEX14", ".kbwg")

	viaGeneric, err := GetGraph[*KBWG](cfg, "ZZZLEX14", WithDistribution(fakeDistTwo))
	is.NoErr(err)
	viaSpecific, err := GetKBWG(cfg, "ZZZLEX14", WithDistribution(fakeDistTwo))
	is.NoErr(err)

	is.True(viaGeneric == viaSpecific)
}

func TestGetGraphWithDistribution_UnrecognizedLexiconName_ErrorsWithoutOverride(t *testing.T) {
	is := is.New(t)
	cfg := newTestConfig(t)
	writeFakeLexFile(t, cfg, "ZZZLEX15", ".kwg")

	_, err := GetGraph[*KWG](cfg, "ZZZLEX15", WithDistribution(""))
	is.True(err != nil)
}
