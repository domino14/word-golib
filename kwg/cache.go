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

// LoadWordGraph loads either a KWG or KBWG based on the file extension.
// The letter distribution (and thus alphabet) is guessed from the lexicon
// name; use LoadWordGraphWithDistribution to override that guess explicitly.
func LoadWordGraph[T WordGraphConstraint](cfg *config.Config, filename string) (T, error) {
	return loadWordGraph[T](cfg, filename, "")
}

// LoadWordGraphWithDistribution loads either a KWG or KBWG based on the file
// extension, using distName as the letter distribution (and thus alphabet)
// instead of guessing it from the lexicon name. This allows callers to use
// lexicon names that word-golib has no built-in knowledge of (e.g. a
// user-defined lexicon), as long as the caller knows what letter
// distribution it should use. If distName is empty, this behaves the same
// as LoadWordGraph.
func LoadWordGraphWithDistribution[T WordGraphConstraint](cfg *config.Config, filename, distName string) (T, error) {
	return loadWordGraph[T](cfg, filename, distName)
}

func loadWordGraph[T WordGraphConstraint](cfg *config.Config, filename, distName string) (T, error) {
	log.Debug().Msgf("Loading %v ...", filename)
	file, filesize, err := cache.Open(filename)
	var result T
	if err != nil {
		return result, err
	}
	defer file.Close()

	// Determine if it's a KWG or KBWG file based on extension
	switch any(result).(type) {
	case *KBWG:
		kbwg, err := ScanKBWG(file, filesize)
		if err != nil {
			return result, err
		}
		result = any(kbwg).(T)
	case *KWG:
		kwg, err := ScanKWG(file, filesize)
		if err != nil {
			return result, err
		}
		result = any(kwg).(T)
	default:
		return result, errors.New("unsupported graph type for loading")
	}

	// Set lexicon name and alphabet
	lexfile := filepath.Base(filename)
	lexname, found := strings.CutSuffix(lexfile, filepath.Ext(lexfile))
	if !found {
		return result, errors.New("filename not in correct format")
	}

	// Note: we deliberately use the uncached NamedLetterDistribution here,
	// not GetDistribution. This function runs inside a locked cache.Load
	// call (from GetKWG/GetKWGWithDistribution/etc.), and GetDistribution
	// itself calls cache.Load, which would deadlock on the non-reentrant
	// cache mutex.
	var ld *tilemapping.LetterDistribution
	if distName != "" {
		ld, err = tilemapping.NamedLetterDistribution(cfg, distName)
	} else {
		ld, err = tilemapping.ProbableLetterDistribution(cfg, lexname)
	}
	if err != nil {
		return result, err
	}

	// We need to set these fields, but the interface doesn't have setters
	// We need to use type assertions to access the underlying types
	switch v := any(result).(type) {
	case *KWG:
		v.lexiconName = lexname
		v.alphabet = ld.TileMapping()
	case *KBWG:
		v.lexiconName = lexname
		v.alphabet = ld.TileMapping()
	}

	return result, nil
}

// LoadKWG loads a KWG from a file (for backward compatibility)
func LoadKWG(cfg *config.Config, filename string) (*KWG, error) {
	return LoadWordGraph[*KWG](cfg, filename)
}

// LoadKBWG loads a KBWG from a file
func LoadKBWG(cfg *config.Config, filename string) (*KBWG, error) {
	return LoadWordGraph[*KBWG](cfg, filename)
}

// CacheLoadFuncKWG is the function that loads a KWG into the global cache
func CacheLoadFuncKWG(cfg *config.Config, key string) (interface{}, error) {
	lexiconName := strings.TrimPrefix(key, CacheKeyPrefixKWG)
	return loadKWGFile(cfg, lexiconName, "")
}

// CacheLoadFuncKBWG is the function that loads a KBWG into the global cache
func CacheLoadFuncKBWG(cfg *config.Config, key string) (interface{}, error) {
	lexiconName := strings.TrimPrefix(key, CacheKeyPrefixKBWG)
	return loadKBWGFile(cfg, lexiconName, "")
}

func loadKWGFile(cfg *config.Config, lexiconName, distName string) (interface{}, error) {
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix

	if kwgPrefix == "" {
		return LoadWordGraphWithDistribution[*KWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kwg"), distName)
	}
	return LoadWordGraphWithDistribution[*KWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kwg"), distName)
}

func loadKBWGFile(cfg *config.Config, lexiconName, distName string) (interface{}, error) {
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix

	if kwgPrefix == "" {
		return LoadWordGraphWithDistribution[*KBWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kbwg"), distName)
	}
	return LoadWordGraphWithDistribution[*KBWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kbwg"), distName)
}

func GetGraph[T WordGraphConstraint](cfg *config.Config, name string) (T, error) {
	var result T
	switch any(result).(type) {
	case *KWG:
		k, err := GetKWG(cfg, name)
		if err != nil {
			return result, err
		}
		result = any(k).(T)
	case *KBWG:
		kb, err := GetKBWG(cfg, name)
		if err != nil {
			return result, err
		}
		result = any(kb).(T)
	default:
		return result, errors.New("unsupported graph type")
	}
	return result, nil
}

// GetGraphWithDistribution loads a named KWG or KBWG from the cache or from a
// file, using distName as the letter distribution instead of guessing it
// from the lexicon name. See LoadWordGraphWithDistribution.
func GetGraphWithDistribution[T WordGraphConstraint](cfg *config.Config, name, distName string) (T, error) {
	var result T
	switch any(result).(type) {
	case *KWG:
		k, err := GetKWGWithDistribution(cfg, name, distName)
		if err != nil {
			return result, err
		}
		result = any(k).(T)
	case *KBWG:
		kb, err := GetKBWGWithDistribution(cfg, name, distName)
		if err != nil {
			return result, err
		}
		result = any(kb).(T)
	default:
		return result, errors.New("unsupported graph type")
	}
	return result, nil
}

// GetKWG loads a named KWG from the cache or from a file (for backward compatibility)
func GetKWG(cfg *config.Config, name string) (*KWG, error) {
	key := CacheKeyPrefixKWG + name
	obj, err := cache.Load(cfg, key, CacheLoadFuncKWG)
	if err != nil {
		return nil, err
	}

	kwg, ok := obj.(*KWG)
	if !ok {
		return nil, errors.New("could not convert cached object to KWG")
	}
	return kwg, nil
}

// GetKWGWithDistribution loads a named KWG from the cache or from a file,
// using distName as the letter distribution instead of guessing it from the
// lexicon name. This is useful for lexicon names that word-golib has no
// built-in knowledge of. If distName is empty, this behaves the same as
// GetKWG.
func GetKWGWithDistribution(cfg *config.Config, name, distName string) (*KWG, error) {
	key := CacheKeyPrefixKWG + name
	if distName != "" {
		key += ":" + strings.ToLower(distName)
	}
	obj, err := cache.Load(cfg, key, func(cfg *config.Config, _ string) (interface{}, error) {
		return loadKWGFile(cfg, name, distName)
	})
	if err != nil {
		return nil, err
	}

	kwg, ok := obj.(*KWG)
	if !ok {
		return nil, errors.New("could not convert cached object to KWG")
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
		return nil, errors.New("could not convert cached object to KBWG")
	}
	return kbwg, nil
}

// GetKBWGWithDistribution loads a named KBWG from the cache or from a file,
// using distName as the letter distribution instead of guessing it from the
// lexicon name. If distName is empty, this behaves the same as GetKBWG.
func GetKBWGWithDistribution(cfg *config.Config, name, distName string) (*KBWG, error) {
	key := CacheKeyPrefixKBWG + name
	if distName != "" {
		key += ":" + strings.ToLower(distName)
	}
	obj, err := cache.Load(cfg, key, func(cfg *config.Config, _ string) (interface{}, error) {
		return loadKBWGFile(cfg, name, distName)
	})
	if err != nil {
		return nil, err
	}

	kbwg, ok := obj.(*KBWG)
	if !ok {
		return nil, errors.New("could not convert cached object to KBWG")
	}
	return kbwg, nil
}
