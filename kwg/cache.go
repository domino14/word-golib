package kwg

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/domino14/word-golib/cache"
	"github.com/domino14/word-golib/config"
	"github.com/domino14/word-golib/tilemapping"
)

const (
	CacheKeyPrefix = "kwg:"
)

// CacheLoadFunc is the function that loads an object into the global cache.
func CacheLoadFunc(cfg *config.Config, key string) (interface{}, error) {
	lexiconName := strings.TrimPrefix(key, CacheKeyPrefix)
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix
	if kwgPrefix == "" {
		return LoadKWG(cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kwg"))
	}

	return LoadKWG(cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kwg"))

}

func LoadKWG(cfg *config.Config, filename string) (*KWG, error) {
	log.Debug().Msgf("Loading %v ...", filename)
	file, filesize, err := cache.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// KWG is a simple map of nodes. There is no alphabet information in it,
	// so we must derive it from the filename, for now.
	lexfile := filepath.Base(filename)
	lexname, found := strings.CutSuffix(lexfile, ".kwg")
	if !found {
		return nil, errors.New("filename not in correct format")
	}

	kwg, err := ScanKWG(file, filesize)
	if err != nil {
		return nil, err
	}
	kwg.lexiconName = lexname

	ld, err := tilemapping.ProbableLetterDistribution(cfg, lexname)
	if err != nil {
		return nil, err
	}
	// we don't care about the distribution right now, just the tilemapping.
	kwg.alphabet = ld.TileMapping()

	return kwg, nil

}

// Get loads a named KWG from the cache or from a file
func Get(cfg *config.Config, name string) (*KWG, error) {

	key := CacheKeyPrefix + name
	obj, err := cache.Load(cfg, key, CacheLoadFunc)
	if err != nil {
		return nil, err
	}
	ret, ok := obj.(*KWG)
	if !ok {
		return nil, errors.New("could not read kwg from file")
	}
	return ret, nil
}
