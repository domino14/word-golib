package kwg

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

var DefaultConfig = map[string]any{
	"data-path":                   os.Getenv("DATA_PATH"),
	"default-lexicon":             "NWL20",
	"default-letter-distribution": "English",
}

func TestLoadKWG(t *testing.T) {
	is := is.New(t)
	kwg, err := Get(DefaultConfig, "NWL20")
	is.NoErr(err)
	is.Equal(len(kwg.nodes), 855967)
}
