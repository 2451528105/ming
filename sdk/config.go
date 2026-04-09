package sdk

import (
	"ming/sdk/xlog"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func ReadConfig(filename string, v any, listen bool) error {
	vp := viper.New()
	if err := load(vp, filename, v); err != nil {
		return err
	}

	if listen {
		vp.WatchConfig()
		vp.OnConfigChange(func(e fsnotify.Event) {
			if err := load(vp, filename, v); err != nil {
				xlog.Error().Err(err).Msgf("error reloading config:%v", e.Name)
			} else {
				xlog.Info().Msgf("config file changed: %v", v)
			}
		})
	}
	return nil
}
func tagFromFilename(filename string) string {

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	case ".ini":
		return "ini"
	case ".hcl":
		return "hcl"
	default:
		return "mapstructure"
	}
}

func load(vp *viper.Viper, filename string, v any) error {

	vp.SetConfigFile(filename)
	if err := vp.ReadInConfig(); err != nil {
		return err
	}

	tag := tagFromFilename(filename)
	return vp.Unmarshal(v, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = tag
	})
}
