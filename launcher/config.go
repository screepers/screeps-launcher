package launcher

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ConfigEnv ENV Section of Config
type ConfigEnv struct {
	Shared  map[string]string `yaml:"shared" json:"shared"`
	Backend map[string]string `yaml:"backend" json:"backend"`
	Engine  map[string]string `yaml:"engine" json:"engine"`
	Storage map[string]string `yaml:"storage" json:"storage"`
}

// ConfigBackup Backup section of config
type ConfigBackup struct {
	Dirs  []string `yaml:"dirs" json:"dirs"`
	Files []string `yaml:"files" json:"files"`
}

// ConfigCli CLI section of config
type ConfigCli struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Host     string `yaml:"host" json:"host"`
	Port     int16  `yaml:"port" json:"port"`
}

// Config server config structure
type Config struct {
	SteamKey       string            `yaml:"steamKey" json:"steamKey"`
	SteamKeyFile   string            `yaml:"steamKeyFile" json:"steamKeyFile"`
	Cli            *ConfigCli        `yaml:"cli" json:"cli"`
	Env            *ConfigEnv        `yaml:"env" json:"env"`
	Processors     int               `yaml:"processors" json:"processors"`
	RunnerThreads  int               `yaml:"runnerThreads" json:"runnerThreads"`
	Version        string            `yaml:"version" json:"version"`
	NodeVersion    string            `yaml:"nodeVersion" json:"nodeVersion"`
	Mods           []string          `yaml:"mods" json:"mods"`
	Bots           map[string]string `yaml:"bots" json:"bots"`
	ExtraPackages  map[string]string `yaml:"extraPackages" json:"extraPackages"`
	PinnedPackages map[string]string `yaml:"pinnedPackages" json:"pinnedPackages"`
	LocalMods      string            `yaml:"localMods" json:"localMods"`
	Backup         *ConfigBackup     `yaml:"backup" json:"backup"`
	Modules        map[string]bool   `yaml:"modules" json:"modules"`
}

// NewConfig Create a new Config
func NewConfig() *Config {
	cores := runtime.NumCPU()
	runners := math.Max(1, float64(cores)-1)
	return &Config{
		Processors:    cores,
		RunnerThreads: int(runners),
		Version:       "latest",
		NodeVersion:   "Erbium",
		Cli: &ConfigCli{
			Username: "",
			Password: "",
			Host:     "127.0.0.1",
			Port:     21026,
		},
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
		LocalMods:      "mods",
		Mods:           make([]string, 0),
		Bots:           make(map[string]string),
		ExtraPackages:  make(map[string]string),
		PinnedPackages: map[string]string{
				"ssri": "8.0.1",
				"cacache": "16.1.3",
				"passport-steam": "1.0.17",
				"minipass-fetch": "3.0.3",
		},
		Backup: &ConfigBackup{
			Dirs:  make([]string, 0),
			Files: make([]string, 0),
		},
		Modules: map[string]bool{
			"backend":   true,
			"main":      true,
			"processor": true,
			"runner":    true,
			"storage":   true,
		},
	}
}

// GetConfig loads a config from config.yml
func (c *Config) GetConfig(dir string) (*Config, error) {
	files := []string{"config.yml", "config.yaml"}
	for _, file := range files {
		if dir != "" {
			file = filepath.Join(dir, file)
		}
		configFile, err := ioutil.ReadFile(file)
		if err == nil {
			err = yaml.Unmarshal(configFile, c)
			if err != nil {
				return nil, err
			}
			log.Printf("Loaded config from %s", file)
		}
	}
	c.Env.Shared["MODFILE"] = "mods.json"
	for key, val := range c.Env.Shared {
		c.Env.Backend[key] = val
		c.Env.Engine[key] = val
		c.Env.Storage[key] = val
	}
	if c.RunnerThreads > 0 {
		c.Env.Engine["RUNNER_THREADS"] = strconv.Itoa(c.RunnerThreads)
	}
	if _, err := os.Stat("STEAM_KEY"); c.SteamKeyFile == "" && !os.IsNotExist(err) {
		c.SteamKeyFile = "STEAM_KEY"
	}
	if c.SteamKeyFile != "" {
		bytes, err := ioutil.ReadFile(c.SteamKeyFile)
		if err != nil {
			return c, errors.Wrap(err, "Failed to load steamKeyFile")
		}
		c.SteamKey = strings.TrimSpace(string(bytes))
	}
	if c.SteamKey != "" {
		c.Env.Backend["STEAM_KEY"] = c.SteamKey
	}
	if c.Backup.Dirs == nil {
		c.Backup.Dirs = make([]string, 0)
	}
	if c.Backup.Files == nil {
		c.Backup.Files = make([]string, 0)
	}
	return c, nil
}
