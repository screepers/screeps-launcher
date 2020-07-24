package launcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/otiai10/copy"
	"github.com/screepers/screeps-launcher/v1/install"
	// "path/filepath"
	// "strings"
)

// Launcher manages server
type Launcher struct {
	config    *Config
	needsInit bool
}

// Prepare loads config
func (l *Launcher) Prepare() {
	c := NewConfig()
	_, err := c.GetConfig("")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	l.config = c
	l.needsInit = false
	checkFiles := []string{"package.json", "assets", "db.json", ".screepsrc", "example-mods"}
	for _, file := range checkFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			l.needsInit = true
		}
	}
}

// Upgrade upgrades screeps
func (l *Launcher) Upgrade() error {
	os.Remove("yarn.lock")
	return l.Apply()
}

// Apply applies config without starting
func (l *Launcher) Apply() error {
	var err error
	if _, err := os.Stat(install.NodePath); os.IsNotExist(err) {
		log.Print("Installing Node")
		err = install.Node(l.config.NodeVersion)
		if err != nil {
			return err
		}
		// This requires an admin prompt, need to figure out howto prompt the user.
		// if runtime.GOOS == "windows" {
		// 	log.Print("Installing windows-build-tools (This may take a while)")
		// 	err = install.WindowsBuildTools()
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	} else {
		cmd := exec.Command(install.NodePath, "--version")
		ret, err := cmd.Output()
		if err != nil {
			return err
		}
		curVer := strings.TrimRight(string(ret), "\r\n")
		if ver, err := install.GetNodeVersion(l.config.NodeVersion); ver != curVer {
			if err != nil {
				log.Printf("Could not get node version: %s", err.Error())
			}
			if ver != "" {
				log.Printf("Node version doesn't match\n Current Version: %s\n Wanted Version: %s\nUpdating...", curVer, ver)
				os.RemoveAll("deps")
				os.Remove("yarn.lock")
				err = install.Node(l.config.NodeVersion)
				if err != nil {
					return err
				}
			}
		}
	}
	if _, err := os.Stat(install.YarnPath); os.IsNotExist(err) {
		log.Print("Installing Yarn")
		err = install.Yarn()
		if err != nil {
			return err
		}
	}
	log.Print("Writing package.json")
	err = writePackage(l.config)
	if err != nil {
		return err
	}
	log.Print("Running yarn")
	err = runYarn()
	if err != nil {
		return err
	}
	if l.needsInit {
		log.Print("Initializing server")
		copy.Copy(filepath.Join("node_modules", "@screeps", "launcher", "init_dist"), ".")
		os.RemoveAll(filepath.Join("node_modules", ".hooks"))
	}

	os.MkdirAll(l.config.LocalMods, 0755)
	cliModPath := filepath.Join(l.config.LocalMods, "screeps-launcher-cli.js")
	err = ioutil.WriteFile(cliModPath, []byte(cliMod), 0644)
	if err != nil {
		log.Printf("WARNING: failed to write %s, CLI may not function correctly. Error was: %v", cliModPath, err)
	}
	log.Print("Writing mods.json")
	err = writeMods(l.config)
	if err != nil {
		return err
	}
	return nil
}

// Start starts the server
func (l *Launcher) Start() error {
	err := l.Apply()
	if err != nil {
		return err
	}
	log.Print("Starting Server")
	s := NewServer(l.config)
	if err := s.Start(); err != nil {
		log.Printf("Error while starting: %v", err)
	} else {
		log.Print("Started")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM)
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGHUP:
				log.Printf("SIGHUP Received")
				log.Printf("Applying config")
				_, err := l.config.GetConfig("")
				if err != nil {
					log.Fatalf("Error loading config: %v", err)
				}
				select {
				case <-time.After(2 * time.Second):
					l.Apply()
				case <-c:
					l.Upgrade()
				}
				log.Print("Stopping")
				if err := s.Stop(); err != nil {
					log.Printf("Error while stopping: %v", err)
				} else {
					log.Print("Stopped")
				}
				time.Sleep(1 * time.Second)
				log.Print("Starting")
				if err := s.Start(); err != nil {
					log.Printf("Error while starting: %v", err)
				} else {
					log.Print("Started")
				}
			case syscall.SIGTERM:
				fallthrough
			case os.Interrupt:
				fallthrough
			case os.Kill:
				log.Printf("Stopping")
				if err := s.Stop(); err != nil {
					log.Printf("Error while stopping: %v", err)
				} else {
					log.Print("Stopped")
				}
				return nil
			}
		}
	}
}

// Cli Opens a CLI
func (l *Launcher) Cli() error {
	log.Print("Starting CLI")
	runCli(l.config)
	return nil
}

func cmdExists(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}

func runYarn() error {
	cmd := exec.Command(install.NodePath, install.YarnPath)
	newPath := filepath.SplitList(os.Getenv("PATH"))
	cwd, _ := os.Getwd()
	newPath = append([]string{filepath.Join(cwd, filepath.Dir(install.NodePath))}, newPath...)
	lenv := os.Environ()
	lenv = append(lenv, fmt.Sprintf("PATH=%s", strings.Join(newPath, string(filepath.ListSeparator))))
	cmd.Env = lenv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
