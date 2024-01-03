package kwg

import (
	"fmt"
	"os"
	"testing"

	"github.com/matryer/is"
)

var DefaultConfig = map[string]any{
	"DataPath":                  os.Getenv("DATA_PATH"),
	"DefaultLexicon":            "NWL20",
	"DefaultLetterDistribution": "English",
}

func TestLoadKWG(t *testing.T) {
	is := is.New(t)
	fmt.Println("DefaultConfig", DefaultConfig)
	kwg, err := Get(DefaultConfig, "NWL20")
	is.NoErr(err)
	is.Equal(len(kwg.nodes), 855967)
}
