package launcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
			bots[i] = main
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
				mods = append(mods, path.Join(c.LocalMods, file.Name()))
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
	bytes, err := ioutil.ReadFile(path.Join("node_modules", mod, "package.json"))
	if err != nil {
		return "", err
	}
	var pack PackageJson
	err = json.Unmarshal(bytes, &pack)
	if err != nil {
		return "", err
	}
	return path.Join("node_modules", mod, pack.Main), nil
}

func runYarn() error {
	cmd := exec.Command(path.Join("deps", "yarn", "bin", "yarn"))
	newPath := filepath.SplitList(os.Getenv("PATH"))
	cwd, _ := os.Getwd()
	newPath = append([]string{path.Join(cwd, "deps", "node", "bin")}, newPath...)
	cmd.Env = append(cmd.Env, "PATH="+strings.Join(newPath, string(filepath.ListSeparator)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func runNpm() error {
	cmd := exec.Command(path.Join("deps", "node", "bin", "node"), path.Join("deps", "node", "bin", "npm"), "--no-audit", "--silent", "install")
	newPath := filepath.SplitList(os.Getenv("PATH"))
	cwd, _ := os.Getwd()
	newPath = append([]string{path.Join(cwd, "deps", "node", "bin")}, newPath...)
	cmd.Env = append(cmd.Env, "PATH="+strings.Join(newPath, string(filepath.ListSeparator)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
