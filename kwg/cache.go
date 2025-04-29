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
	CacheKeyPrefixKWG  = "kwg:"
	CacheKeyPrefixKBWG = "kbwg:"
)

// LoadWordGraph loads either a KWG or KBWG based on the file extension
func LoadWordGraph(cfg *config.Config, filename string) (WordGraph, error) {
	log.Debug().Msgf("Loading %v ...", filename)
	file, filesize, err := cache.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Determine if it's a KWG or KBWG file based on extension
	var wg WordGraph
	if strings.HasSuffix(filename, ".kbwg") {
		kbwg, err := ScanKBWG(file, filesize)
		if err != nil {
			return nil, err
		}
		wg = kbwg
	} else {
		kwg, err := ScanKWG(file, filesize)
		if err != nil {
			return nil, err
		}
		wg = kwg
	}

	// Set lexicon name and alphabet
	lexfile := filepath.Base(filename)
	lexname, found := strings.CutSuffix(lexfile, filepath.Ext(lexfile))
	if !found {
		return nil, errors.New("filename not in correct format")
	}

	// We need to set these fields, but the interface doesn't have setters
	// We need to use type assertions to access the underlying types
	switch v := wg.(type) {
	case *KWG:
		v.lexiconName = lexname
		ld, err := tilemapping.ProbableLetterDistribution(cfg, lexname)
		if err != nil {
			return nil, err
		}
		v.alphabet = ld.TileMapping()
	case *KBWG:
		v.lexiconName = lexname
		ld, err := tilemapping.ProbableLetterDistribution(cfg, lexname)
		if err != nil {
			return nil, err
		}
		v.alphabet = ld.TileMapping()
	}

	return wg, nil
}

// LoadKWG loads a KWG from a file (for backward compatibility)
func LoadKWG(cfg *config.Config, filename string) (*KWG, error) {
	wg, err := LoadWordGraph(cfg, filename)
	if err != nil {
		return nil, err
	}

	kwg, ok := wg.(*KWG)
	if !ok {
		return nil, errors.New("could not convert WordGraph to KWG")
	}
	return kwg, nil
}

// LoadKBWG loads a KBWG from a file
func LoadKBWG(cfg *config.Config, filename string) (*KBWG, error) {
	wg, err := LoadWordGraph(cfg, filename)
	if err != nil {
		return nil, err
	}

	kbwg, ok := wg.(*KBWG)
	if !ok {
		return nil, errors.New("could not convert WordGraph to KBWG")
	}
	return kbwg, nil
}

// CacheLoadFuncKWG is the function that loads a KWG into the global cache
func CacheLoadFuncKWG(cfg *config.Config, key string) (interface{}, error) {
	lexiconName := strings.TrimPrefix(key, CacheKeyPrefixKWG)
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix

	if kwgPrefix == "" {
		return LoadWordGraph(cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kwg"))
	}

	return LoadWordGraph(cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kwg"))
}

// CacheLoadFuncKBWG is the function that loads a KBWG into the global cache
func CacheLoadFuncKBWG(cfg *config.Config, key string) (interface{}, error) {
	lexiconName := strings.TrimPrefix(key, CacheKeyPrefixKBWG)
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix

	if kwgPrefix == "" {
		return LoadWordGraph(cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kbwg"))
	}

	return LoadWordGraph(cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kbwg"))
}

// Get loads a named KWG from the cache or from a file (for backward compatibility)
func Get(cfg *config.Config, name string) (*KWG, error) {
	key := CacheKeyPrefixKWG + name
	obj, err := cache.Load(cfg, key, CacheLoadFuncKWG)
	if err != nil {
		return nil, err
	}

	kwg, ok := obj.(*KWG)
	if !ok {
		return nil, errors.New("could not convert WordGraph to KWG")
	}
	return kwg, nil
}

// GetKBWG loads a named KBWG from the cache or from a file
func GetKBWG(cfg *config.Config, name string) (*KBWG, error) {
	key := CacheKeyPrefixKBWG + name
	obj, err := cache.Load(cfg, key, CacheLoadFuncKBWG)
	if err != nil {
		return nil, err
	}

	kbwg, ok := obj.(*KBWG)
	if !ok {
		return nil, errors.New("could not convert WordGraph to KBWG")
	}
	return kbwg, nil
}

// GetWordGraph loads either a KWG or KBWG based on the isKBWG flag
func GetWordGraph(cfg *config.Config, name string, isKBWG bool) (WordGraph, error) {
	if isKBWG {
		return GetKBWG(cfg, name)
	}
	return Get(cfg, name)
}
