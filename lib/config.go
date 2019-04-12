package lib

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mudkipme/lilycove/lib/purge"
	"github.com/mudkipme/lilycove/lib/queue"
)

var appConfig *AppConfig
var muc sync.Mutex

// HTTPConfig defines the configuration of the purger service
type HTTPConfig struct {
	Port int `toml:"port"`
}

// AppConfig defines the app configruation
type AppConfig struct {
	HTTP  *HTTPConfig   `toml:"http"`
	Queue *queue.Config `toml:"queue"`
	Purge *purge.Config `toml:"purge"`
}

// Config reads the app configruation from `config.toml`
func Config() *AppConfig {
	muc.Lock()
	defer muc.Unlock()
	if appConfig != nil {
		return appConfig
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bindir := filepath.Dir(os.Args[0])
	if !filepath.IsAbs(bindir) {
		bindir = filepath.Join(pwd, bindir)
	}

	var configFile string
	if _, err := os.Stat(filepath.Join(bindir, "config.toml")); !os.IsNotExist(err) {
		configFile = filepath.Join(bindir, "config.toml")
	} else {
		configFile = filepath.Join(pwd, "config.toml")
	}

	_, err = toml.DecodeFile(configFile, &appConfig)
	if err != nil {
		panic(err)
	}
	return appConfig
}
