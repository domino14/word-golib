package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config structure just wraps a viper for ease in migrating to viper.
type Config struct {
	viper.Viper
}

var defaultConfig Config

func init() {
	v := viper.New()
	v.Set("DataPath", os.Getenv("DATA_PATH"))
	v.SetDefault("DefaultLexicon", "NWL20")
	v.SetDefault("DefaultLetterDistribution", "English")
	v.SetDefault("TTableFractionOfMem", 0.25)

	defaultConfig.Viper = *v
}

func DefaultConfig() *Config {
	return &defaultConfig
}
