package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"github.com/screepers/screeps-launcher/v1/install"
)

func initServer(c *Config) {
	copy.Copy(filepath.Join("node_modules", "@screeps", "launcher", "init_dist"), ".")
	os.RemoveAll(filepath.Join("node_modules", ".hooks"))
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
	if runtime.GOOS == "windows" {
		module = module + ".cmd"
	}
	os.Mkdir("logs", 0777)
	n := filepath.Join("logs", fmt.Sprintf("%s.log", name))
	f, err := os.OpenFile(n, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening log %s: %v", n, err)
	}
	// logger := log.New(f, fmt.Sprintf("[%s]", name), log.lstdFlags)
	for {
		log.Printf("[%s] exec: %s", name, module)
		newPath := filepath.SplitList(os.Getenv("PATH"))
		cwd, _ := os.Getwd()
		newPath = append([]string{filepath.Join(cwd, filepath.Dir(install.NodePath))}, newPath...)
		env["PATH"] = strings.Join(newPath, string(filepath.ListSeparator))
		lenv := os.Environ()
		for key, val := range env {
			lenv = append(lenv, fmt.Sprintf("%s=%s", key, val))
		}
		fmt.Fprintf(f, "==== %s Starting ====\n", name)
		cmd := exec.Command(filepath.Join("node_modules", ".bin", module))
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Env = lenv
		err := cmd.Run()
		log.Printf("[%s] Exited with error: %v", name, err)
		fmt.Fprintf(f, "==== %s Exited ====\n", name)
		if err != nil {
			time.Sleep(1 * time.Second) // Wait before trying to restart
		}
	}
}
