package launcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PackageJson struct {
	Name         string            `json:"name"`
	Main         string            `json:"main"`
	Dependencies map[string]string `json:"dependencies"`
	Private      bool              `json:"private"`
}

func writePackage(c *Config) error {
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
		if !strings.Contains(bot, string(filepath.Separator)) { // Ignore local bots
			deps[bot] = "*"
		}
	}
	var pack PackageJson
	pack.Dependencies = deps
	pack.Private = true
	pack.Name = "screeps-private-server"
	bytes, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("package.json", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeMods(c *Config) error {
	bots := make(map[string]string)
	mods := make([]string, len(c.Mods))
	for i, mod := range c.Mods {
		main, err := getPackageMain(mod)
		if err != nil {
			return err
		}
		mods[i] = main
	}
	for i, bot := range c.Bots {
		if strings.HasPrefix(bot, ".") {
			bots[i] = bot
		} else {
			main, err := getPackageMain(bot)
			if err != nil {
				return err
			}
			bots[i] = filepath.Dir(main)
		}
	}
	var out struct {
		Mods []string          `json:"mods"`
		Bots map[string]string `json:"bots"`
	}

	if c.LocalMods != "" {
		os.MkdirAll(c.LocalMods, 0777)
		files, err := ioutil.ReadDir(c.LocalMods)
		if err != nil {
			return err
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".js") {
				mods = append(mods, filepath.Join(c.LocalMods, file.Name()))
			}
		}
	}

	out.Mods = mods
	out.Bots = bots
	bytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("Writing %d mods and %d bots", len(out.Mods), len(out.Bots))
	err = ioutil.WriteFile("mods.json", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getPackageMain(mod string) (string, error) {
	bytes, err := ioutil.ReadFile(filepath.Join("node_modules", mod, "package.json"))
	if err != nil {
		return "", err
	}
	var pack PackageJson
	err = json.Unmarshal(bytes, &pack)
	if err != nil {
		return "", err
	}
	return filepath.Join("node_modules", mod, pack.Main), nil
}
