package kwg

import (
	"fmt"
	"testing"

	"github.com/matryer/is"

	"github.com/domino14/word-golib/config"
)

var DefaultConfig = config.DefaultConfig()

func TestLoadKWG(t *testing.T) {
	is := is.New(t)
	fmt.Println("DefaultConfig", DefaultConfig)
	kwg, err := Get(DefaultConfig, "NWL20")
	is.NoErr(err)
	is.Equal(len(kwg.nodes), 855967)
}
