package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
	"log"
	"fmt"
	"os/exec"
	"os"
	"strings"
	"path"
	"time"
)

type ConfigEnv struct {
	Shared map[string]string `yaml:"shared"`
	Backend map[string]string `yaml:"backend"`
	Engine map[string]string `yaml:"engine"`
	Storage map[string]string `yaml:"storage"`
}
type Config struct {
	Env *ConfigEnv `yaml:"env"`
	Processors int `yaml:"processors"`
	Version string `yaml:"version"`
	Mods []string `yaml:"mods"`
	Bots map[string]string `yaml:"bots"`
	ExtraPackages map[string]string `yaml:"extraPackages"`
	LocalMods string `yaml:"localMods"`
}

type PackageJson struct {
  Main string `json:"main"`
  Dependencies map[string]string `json:"dependencies"`
}

func NewConfig() *Config {
	ce := ConfigEnv{}
	ce.Shared = map[string]string{
		"MODFILE": "mods.json",
	}
	ce.Backend = map[string]string{
		"GAME_HOST": "0.0.0.0",
		"GAME_PORT": "21025",
		"CLI_HOST": "127.0.0.1",
		"CLI_PORT": "21026",
		"ASSET_DIR": "assets",
	}
	ce.Engine = map[string]string{
		"DRIVER_MODULE": "@screeps/driver",
	}
	ce.Storage = map[string]string{}
	c := Config{}
	c.Processors = 2
	c.Env = &ce
	c.Mods = make([]string,0)
	c.Bots = make(map[string]string)
	c.ExtraPackages = make(map[string]string)
	return &c
}

func (c *Config) GetConfig() *Config {
	configFile, err := ioutil.ReadFile("config.yml")
	check(err)
	err = yaml.Unmarshal(configFile, c)
	check(err)
	c.Env.Shared["MODFILE"] = "mods.json"
	for key, val := range c.Env.Shared {
		c.Env.Backend[key] = val
		c.Env.Engine[key] = val
	}
	return c
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := NewConfig()
	c.GetConfig()
  fmt.Print("Writing package.json...")
	writePackage(c)
  fmt.Println("Done")
  fmt.Println("Running yarn...")
	runYarn()
	fmt.Println("Done")
  fmt.Print("Writing mods.json...")
  writeMods(c)
	fmt.Println("Done")
	runServer(c)
	select {} // Wait Forever
}

func writePackage(c *Config) {
  deps := make(map[string]string)

  if c.Version == "latest" {
    c.Version = "*"
  }

  deps["screeps"] = c.Version

  for _, mod := range c.Mods {
    
    deps[mod] = "*"
  }

  for pkg, ver := range c.ExtraPackages {
    deps[pkg] = ver
  }

  for _, bot := range c.Bots {
    if strings.HasPrefix(bot, ".") { // Ignore local bots
      continue
    }
    deps[bot] = "*"
  } 
  var pack PackageJson
  pack.Dependencies = deps
  bytes, err := json.MarshalIndent(pack, "", "  ")
  check(err)
  err = ioutil.WriteFile("package.json", bytes, 0644)
  check(err)
}

func writeMods(c *Config) {
  bots := make(map[string]string)
  mods := make([]string, len(c.Mods))
  for i, mod := range c.Mods {
    main := getPackageMain(mod)
    mods[i] = main
  }
  for i, bot := range c.Bots {
    if strings.HasPrefix(bot, ".") {
      bots[i] = bot
    } else {
      main := getPackageMain(bot)
      bots[i] = main
    }
  }
  var out struct {
    Mods []string `json:"mods"`
    Bots map[string]string `json:"bots"`
  }

  if c.LocalMods != "" {
    files, err := ioutil.ReadDir(c.LocalMods)
    check(err)
    for _, file := range files {
      if strings.HasSuffix(file.Name(), ".js") {
        mods = append(mods, path.Join(c.LocalMods, file.Name()))
      }
    }
  }

  out.Mods = mods
  out.Bots = bots
  bytes, err := json.MarshalIndent(out, "", "  ")
  check(err)
  err = ioutil.WriteFile("mods.json", bytes, 0644)
  check(err)  
}

func getPackageMain(mod string) string {
  bytes, err := ioutil.ReadFile(path.Join("node_modules", mod, "package.json"))
  check(err)
  var pack PackageJson
  err = json.Unmarshal(bytes, &pack)
  check(err)
  return path.Join("node_modules", mod, pack.Main)
}

func runYarn() {
  cmd := exec.Command("yarn")
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  err := cmd.Run()
  check(err)
}

func runServer(c *Config) {
  needsStorage := true
  for _, mod := range c.Mods {
    if mod == "screepsmod-mongo" {
      needsStorage = false
    }
  }
  if needsStorage {
    go runModule("storage", "screeps-storage", c.Env.Storage)
    time.Sleep(5 * time.Second) // Give storage time to launch
  }
  go runModule("runner", "screeps-engine-runner", c.Env.Engine)
  for i := 0; i < c.Processors; i++ {
    go runModule(fmt.Sprintf("processor_%d", i), "screeps-engine-processor", c.Env.Engine)
  }
  go runModule("main", "screeps-engine-main", c.Env.Engine)
  go runModule("backend", "screeps-backend", c.Env.Backend)
}

func runModule(name string, module string, env map[string]string) {
	for {
    log.Printf("[%s] exec: npx %s", name, module)
    lenv := os.Environ()
    for key, val := range env {
			lenv = append(lenv, fmt.Sprintf("%s=%s", key, val))
    }
    cmd := exec.Command("npx", module)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Env = lenv
    err := cmd.Run()
    log.Printf("[%s] Exited with error: %v", name, err)
    check(err)
  }
}