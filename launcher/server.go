package launcher

import (
	"context"
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

// Status is the status of the server
type Status int

//
const (
	StatusStopped Status = iota
	StatusUpgrading
	StatusStarting
	StatusRunning
	StatusStopping
)

// Server is a screeps server instance
type Server struct {
	ctx    context.Context
	status Status
	config *Config
	cancel context.CancelFunc
}

// NewServer returns a new Server instance
func NewServer(c *Config) *Server {
	return &Server{
		config: c,
		status: StatusStopped,
	}
}

// Status returns the status of the server
func (s Server) Status() Status {
	return s.status
}

// Start the server
func (s *Server) Start() error {
	if s.status != StatusStopped {
		return fmt.Errorf("Server not stopped. Status: %v", s.status)
	}
	s.status = StatusStarting
	s.ctx, s.cancel = context.WithCancel(context.Background())
	needsStorage := true

	for _, mod := range s.config.Mods {
		if mod == "screepsmod-mongo" {
			needsStorage = false
		}
	}
	if key := os.Getenv("STEAM_KEY"); key == "" && s.config.Env.Backend["STEAM_KEY"] == "" {
		log.Print("STEAM_KEY is not set, either set an environment variable or steamKey in the config.yml")
		log.Print("Steam key can be obtained from https://steamcommunity.com/dev/apikey")
		os.Exit(1)
	}
	os.Mkdir("logs", 0777)
	if needsStorage {
		go s.runModule(s.ctx, "storage", "screeps-storage", s.config.Env.Storage)
		time.Sleep(3 * time.Second) // Give storage time to launch
	}
	go s.runModule(s.ctx, "runner", "screeps-engine-runner", s.config.Env.Engine)
	for i := 0; i < s.config.Processors; i++ {
		go s.runModule(s.ctx, fmt.Sprintf("processor_%d", i), "screeps-engine-processor", s.config.Env.Engine)
	}
	go s.runModule(s.ctx, "main", "screeps-engine-main", s.config.Env.Engine)
	go s.runModule(s.ctx, "backend", "screeps-backend", s.config.Env.Backend)
	s.status = StatusRunning
	return nil
}

// Stop the server
func (s *Server) Stop() error {
	if s.status != StatusRunning {
		return fmt.Errorf("Server not running. Status: %v", s.status)
	}
	s.status = StatusStopping
	s.cancel()
	s.status = StatusStopped
	return nil
}

func (s *Server) runModule(ctx context.Context, name string, module string, env map[string]string) {
	if runtime.GOOS == "windows" {
		module = module + ".cmd"
	}
	n := filepath.Join("logs", fmt.Sprintf("%s.log", name))
	f, err := os.OpenFile(n, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening log %s: %v", n, err)
	}
	defer f.Close()
	for {
		log.Printf("[%s] exec: %s", name, module)
		newPath := filepath.SplitList(os.Getenv("PATH"))
		cwd, _ := os.Getwd()
		newPath = append([]string{filepath.Join(cwd, filepath.Dir(install.NodePath))}, newPath...)
		lenv := os.Environ()
		for key, val := range env {
			lenv = append(lenv, fmt.Sprintf("%s=%s", key, val))
		}
		lenv = append(lenv, fmt.Sprintf("PATH=%s", strings.Join(newPath, string(filepath.ListSeparator))))
		fmt.Fprintf(f, "==== %s Starting ====\n", name)
		cmd := exec.CommandContext(ctx, filepath.Join("node_modules", ".bin", module))
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Env = lenv
		err := cmd.Run()
		log.Printf("[%s] Exited with error: %v", name, err)
		fmt.Fprintf(f, "==== %s Exited ====\n", name)
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func initServer(c *Config) {
	copy.Copy(filepath.Join("node_modules", "@screeps", "launcher", "init_dist"), ".")
	os.RemoveAll(filepath.Join("node_modules", ".hooks"))
}

func runServer(c *Config) {
	s := NewServer(c)
	s.Start()
	select {}
}
