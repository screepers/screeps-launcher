package launcher

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigEnv struct {
	Shared  map[string]string `yaml:"shared"`
	Backend map[string]string `yaml:"backend"`
	Engine  map[string]string `yaml:"engine"`
	Storage map[string]string `yaml:"storage"`
}

type Config struct {
	SteamKey      string            `yaml:"steamKey"`
	Env           *ConfigEnv        `yaml:"env"`
	Processors    int               `yaml:"processors"`
	Version       string            `yaml:"version"`
	Mods          []string          `yaml:"mods"`
	Bots          map[string]string `yaml:"bots"`
	ExtraPackages map[string]string `yaml:"extraPackages"`
	LocalMods     string            `yaml:"localMods"`
}

func NewConfig() *Config {
	return &Config{
		Processors: 2,
		Version:    "latest",
		Env: &ConfigEnv{
			Shared: map[string]string{
				"MODFILE":      "mods.json",
				"STORAGE_HOST": "127.0.0.1",
				"STORAGE_PORT": "21027",
			},
			Backend: map[string]string{
				"GAME_HOST": "0.0.0.0",
				"GAME_PORT": "21025",
				"CLI_HOST":  "127.0.0.1",
				"CLI_PORT":  "21026",
				"ASSET_DIR": "assets",
			},
			Engine: map[string]string{
				"DRIVER_MODULE": "@screeps/driver",
			},
			Storage: map[string]string{
				"DB_PATH": "db.json",
			},
		},
		Mods:          make([]string, 0),
		Bots:          make(map[string]string),
		ExtraPackages: make(map[string]string),
	}
}

func (c *Config) GetConfig() (*Config, error) {
	configFile, err := ioutil.ReadFile("config.yml")
	if err == nil {
		err = yaml.Unmarshal(configFile, c)
		if err != nil {
			return nil, err
		}
	}
	configFile, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		err = yaml.Unmarshal(configFile, c)
		if err != nil {
			return nil, err
		}
	}
	c.Env.Shared["MODFILE"] = "mods.json"
	for key, val := range c.Env.Shared {
		c.Env.Backend[key] = val
		c.Env.Engine[key] = val
		c.Env.Storage[key] = val
	}
	if c.SteamKey != "" {
		c.Env.Backend["STEAM_KEY"] = c.SteamKey
	}
	return c, nil
}
