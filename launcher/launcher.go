package launcher

import (
	"github.com/screepers/screeps-launcher/v1/install"
	"log"
	"os"
	"os/exec"
	"runtime"
)

type Launcher struct {
	config    *Config
	needsInit bool
}

func (l *Launcher) Prepare() {
	c := NewConfig()
	_, err := c.GetConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	l.config = c
	l.needsInit = false
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		l.needsInit = true
	}
}

func (l *Launcher) Upgrade() error {
	os.Remove("yarn.lock")
	return l.Apply()
}

func (l *Launcher) Apply() error {
	var err error
	if _, err := os.Stat("deps/node/bin/node"); os.IsNotExist(err) {
		log.Print("Installing Node")
		err = install.InstallNode("Carbon")
		if err != nil {
			return err
		}
		if runtime.GOOS == "windows" {
			log.Print("Installing windows-build-tools (This may take a while)")
			err = install.InstallWindowsBuildTools()
			if err != nil {
				return err
			}
		}
	}
	if _, err := os.Stat("deps/yarn/bin/yarn"); os.IsNotExist(err) {
		log.Print("Installing Yarn")
		err = install.InstallYarn()
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
		initServer(l.config)
	}
	log.Print("Writing mods.json")
	err = writeMods(l.config)
	if err != nil {
		return err
	}
	return nil
}

func (l *Launcher) Start() error {
	err := l.Apply()
	if err != nil {
		return err
	}
	log.Print("Starting Server")
	runServer(l.config)
	return nil
}

func (l *Launcher) Cli() error {
	log.Print("Starting CLI")
	runCli(l.config)
	return nil
}

func cmdExists(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}
