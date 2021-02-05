package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/spf13/viper"
)

var log *logger.Logger = logger.NewLogger("file-uploader-config")
var configFile *string = new(string)
var lastConf *Config
var defaultConfig = &Config{
	ApiPort:  12000,
	LogLevel: "info",
	Source:   defaultSourceConfig,
	Target:   defaultTargetConfig,
}

type Config struct {
	ApiPort  int     `json:"apiPort"`
	LogLevel string  `json:"logLevel"`
	Source   *Source `json:"source"`
	Target   *Target `json:"target"`
}

func (c *Config) Validate() error {
	if c.ApiPort <= 0 {
		return fmt.Errorf("api port must be greater than 0")
	}
	if err := c.Source.Validate(); err != nil {
		return err
	}
	if err := c.Target.Validate(); err != nil {
		return err
	}
	return nil
}
func (c *Config) hash() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	h := sha256.New()
	_, _ = h.Write(b)
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}
func (c *Config) copy() *Config {
	b, _ := json.Marshal(c)
	n := &Config{}
	_ = json.Unmarshal(b, n)
	return n
}
func load() (*Config, error) {
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, err
}

func Load(cfgCh chan *Config) (*Config, error) {
	viper.AddConfigPath("./")
	cfg, err := load()
	if err != nil {
		return nil, err
	}
	lastConf = cfg.copy()
	cfg.Print()
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		cfg, err := load()
		if err != nil {
			log.Errorf("error loading new configuration file: %s", err.Error())
			return
		}
		if cfg.hash() != lastConf.hash() {
			log.Info("config file changed, reloading...")
			cfg.Print()
			lastConf = cfg.copy()
			cfgCh <- cfg
		}

	})
	return cfg, err
}

func (c *Config) Print() {
	log.Infof("ApiPort-> %d", c.ApiPort)
	log.Infof("LogLevel-> %s", c.LogLevel)
	c.Source.Print()
	c.Target.Print()
}
