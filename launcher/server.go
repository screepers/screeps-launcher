package launcher

import (
	"fmt"
	"github.com/otiai10/copy"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func initServer(c *Config) {
	copy.Copy("node_modules/@screeps/launcher/init_dist/", "./")
	os.RemoveAll("node_modules/.hooks")
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
	select {}
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
		cmd := exec.Command(path.Join("deps", "node", "bin", "node"), path.Join("node_modules", ".bin", module))
		newPath := filepath.SplitList(os.Getenv("PATH"))
		cwd, _ := os.Getwd()
		newPath = append([]string{path.Join(cwd, "deps", "node", "bin")}, newPath...)
		cmd.Env = append(cmd.Env, "PATH="+strings.Join(newPath, string(filepath.ListSeparator)))
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
