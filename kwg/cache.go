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

// loadOpts holds the options settable via LoadOption.
type loadOpts struct {
	distName string
}

// LoadOption customizes how a KWG/KBWG is loaded. See WithDistribution.
type LoadOption func(*loadOpts)

// WithDistribution overrides the letter distribution used to resolve a
// KWG/KBWG's alphabet, instead of guessing it from the lexicon name via
// tilemapping.ProbableLetterDistribution. This is required for lexicon names
// word-golib has no built-in knowledge of (e.g. a user-defined lexicon), as
// long as the caller knows what letter distribution it should use.
func WithDistribution(distName string) LoadOption {
	return func(o *loadOpts) { o.distName = distName }
}

func resolveLoadOpts(opts []LoadOption) loadOpts {
	var o loadOpts
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// LoadWordGraph loads either a KWG or KBWG based on the file extension. By
// default, the letter distribution (and thus alphabet) is guessed from the
// lexicon name; pass WithDistribution to override that guess explicitly.
func LoadWordGraph[T WordGraphConstraint](cfg *config.Config, filename string, opts ...LoadOption) (T, error) {
	o := resolveLoadOpts(opts)
	return loadWordGraph[T](cfg, filename, o.distName)
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
	// call (from GetKWG/GetKBWG/etc.), and GetDistribution itself calls
	// cache.Load, which would deadlock on the non-reentrant cache mutex.
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
func LoadKWG(cfg *config.Config, filename string, opts ...LoadOption) (*KWG, error) {
	return LoadWordGraph[*KWG](cfg, filename, opts...)
}

// LoadKBWG loads a KBWG from a file
func LoadKBWG(cfg *config.Config, filename string, opts ...LoadOption) (*KBWG, error) {
	return LoadWordGraph[*KBWG](cfg, filename, opts...)
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
		return loadWordGraph[*KWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kwg"), distName)
	}
	return loadWordGraph[*KWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kwg"), distName)
}

func loadKBWGFile(cfg *config.Config, lexiconName, distName string) (interface{}, error) {
	dataPath := cfg.DataPath
	kwgPrefix := cfg.KWGPathPrefix

	if kwgPrefix == "" {
		return loadWordGraph[*KBWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", lexiconName+".kbwg"), distName)
	}
	return loadWordGraph[*KBWG](cfg, filepath.Join(dataPath, "lexica", "gaddag", kwgPrefix, lexiconName+".kbwg"), distName)
}

// GetGraph loads a named KWG or KBWG from the cache or from a file. By
// default, the letter distribution (and thus alphabet) is guessed from the
// lexicon name; pass WithDistribution to override that guess explicitly.
func GetGraph[T WordGraphConstraint](cfg *config.Config, name string, opts ...LoadOption) (T, error) {
	var result T
	switch any(result).(type) {
	case *KWG:
		k, err := GetKWG(cfg, name, opts...)
		if err != nil {
			return result, err
		}
		result = any(k).(T)
	case *KBWG:
		kb, err := GetKBWG(cfg, name, opts...)
		if err != nil {
			return result, err
		}
		result = any(kb).(T)
	default:
		return result, errors.New("unsupported graph type")
	}
	return result, nil
}

// GetKWG loads a named KWG from the cache or from a file. By default, the
// letter distribution (and thus alphabet) is guessed from the lexicon name;
// pass WithDistribution to override that guess explicitly, which is required
// for lexicon names word-golib has no built-in knowledge of.
func GetKWG(cfg *config.Config, name string, opts ...LoadOption) (*KWG, error) {
	o := resolveLoadOpts(opts)
	key := CacheKeyPrefixKWG + name
	if o.distName != "" {
		key += ":" + strings.ToLower(o.distName)
	}
	obj, err := cache.Load(cfg, key, func(cfg *config.Config, _ string) (interface{}, error) {
		return loadKWGFile(cfg, name, o.distName)
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

// GetKBWG loads a named KBWG from the cache or from a file. See GetKWG.
func GetKBWG(cfg *config.Config, name string, opts ...LoadOption) (*KBWG, error) {
	o := resolveLoadOpts(opts)
	key := CacheKeyPrefixKBWG + name
	if o.distName != "" {
		key += ":" + strings.ToLower(o.distName)
	}
	obj, err := cache.Load(cfg, key, func(cfg *config.Config, _ string) (interface{}, error) {
		return loadKBWGFile(cfg, name, o.distName)
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
