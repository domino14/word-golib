package tilemapping

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/domino14/word-golib/cache"
	"github.com/domino14/word-golib/config"
)

var CacheKeyPrefix = "letterdist:"

// CacheLoadFunc is the function that loads an object into the global cache.
func CacheLoadFunc(cfg *config.Config, key string) (interface{}, error) {
	dist := strings.TrimPrefix(key, CacheKeyPrefix)
	return NamedLetterDistribution(cfg, dist)
}

// NamedLetterDistribution loads a letter distribution by name.
func NamedLetterDistribution(cfg *config.Config, name string) (*LetterDistribution, error) {
	name = strings.ToLower(name)
	var dataPath = cfg.DataPath
	if dataPath == "" {
		return nil, errors.New("could not find data-path in the configuration")
	}

	filename := filepath.Join(dataPath, "letterdistributions", name)

	file, _, err := cache.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d, err := ScanLetterDistribution(file)
	if err != nil {
		return nil, err
	}
	d.Name = name
	return d, nil
}

// GetDistribution loads a named letter distribution from the cache or from a file.
// The Name field is set when the distribution is first loaded into the cache
// (via CacheLoadFunc -> NamedLetterDistribution), so callers receive a fully
// initialized object. We intentionally do not modify the returned object here
// to avoid data races when multiple goroutines access the same cached instance.
func GetDistribution(cfg *config.Config, name string) (*LetterDistribution, error) {
	key := CacheKeyPrefix + name
	obj, err := cache.Load(cfg, key, CacheLoadFunc)
	if err != nil {
		return nil, err
	}
	ret, ok := obj.(*LetterDistribution)
	if !ok {
		return nil, errors.New("could not read letter distribution from file")
	}
	return ret, nil
}
