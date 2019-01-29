package main

import (
	"encoding/json"
	"fmt"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type ConfigEnv struct {
	Shared  map[string]string `yaml:"shared"`
	Backend map[string]string `yaml:"backend"`
	Engine  map[string]string `yaml:"engine"`
	Storage map[string]string `yaml:"storage"`
}
type Config struct {
	SteamKey			string						`yaml:"steamKey"`
	Env           *ConfigEnv        `yaml:"env"`
	Processors    int               `yaml:"processors"`
	Version       string            `yaml:"version"`
	Mods          []string          `yaml:"mods"`
	Bots          map[string]string `yaml:"bots"`
	ExtraPackages map[string]string `yaml:"extraPackages"`
	LocalMods     string            `yaml:"localMods"`
}

type PackageJson struct {
	Main         string            `json:"main"`
	Dependencies map[string]string `json:"dependencies"`
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

func (c *Config) GetConfig() *Config {
	configFile, err := ioutil.ReadFile("config.yml")
	if err == nil {
		err = yaml.Unmarshal(configFile, c)
		check(err)
	}
	c.Env.Shared["MODFILE"] = "mods.json"
	for key, val := range c.Env.Shared {
		c.Env.Backend[key] = val
		c.Env.Engine[key] = val
		c.Env.Storage[key] = val
	}
	if c.SteamKey != "" {
		c.Env.Backend["SteamKey"] = c.SteamKey
	}
	return c
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	instCmd := "yarn"
	if !cmdExists("npm") {
		log.Fatalln("npm not found! Install nodejs.")
	}
	if !cmdExists("yarn") {
		instCmd = "npm"
		log.Println("yarn not found, while not required, it is recommended")
	}
	c := NewConfig()
	c.GetConfig()
	needsInit := false
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		needsInit = true
	}
	log.Print("Writing package.json")
	writePackage(c)
	log.Printf("Running %s\n", instCmd)
	if instCmd == "yarn" {
		runYarn()
	} else {
		runNpm()
	}
	if needsInit {
		log.Print("Initializing server")
		initServer(c)
	}
	log.Print("Writing mods.json")
	writeMods(c)
	log.Print("Starting Server")
	runServer(c)
	select {} // Wait Forever
}

func cmdExists(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}

func initServer(c *Config) {
	copy.Copy("node_modules/@screeps/launcher/init_dist/", "./")
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
		Mods []string          `json:"mods"`
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

func runNpm() {
	cmd := exec.Command("npm install")
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
	os.Mkdir("logs", 0777)
	n := path.Join("logs", fmt.Sprintf("%s.log", name))
	f, err := os.OpenFile(n, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening log %s: %v", n, err)
	}
	// logger := log.New(f, fmt.Sprintf("[%s]", name), log.lstdFlags)
	for {
		log.Printf("[%s] exec: %s", name, module)
		lenv := os.Environ()
		for key, val := range env {
			lenv = append(lenv, fmt.Sprintf("%s=%s", key, val))
		}
		cmd := exec.Command(path.Join("node_modules", ".bin", module))
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Env = lenv
		err := cmd.Run()
		log.Printf("[%s] Exited with error: %v", name, err)
		if err != nil {
			time.Sleep(1 * time.Second) // Wait before trying to restart
		}
	}
}
