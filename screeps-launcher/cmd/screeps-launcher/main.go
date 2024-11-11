package main

import (
	"flag"
	"log"

	"github.com/screepers/screeps-launcher/v1/launcher"
	"github.com/screepers/screeps-launcher/v1/recovery"
	"github.com/screepers/screeps-launcher/v1/version"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()
	log.Printf("screeps-launcher %s (%s)", version.BuildVersion, version.BuildTime)
	if ver := version.CheckForUpdate(); ver != "" {
		log.Printf("A newer version is available")
	}
	cmd := "start"
	if arg := flag.Arg(0); arg != "" {
		cmd = arg
	}
	if cmd == "version" {
		return
	}
	var err error
	l := launcher.Launcher{}
	l.Prepare()
	r := recovery.New()

	switch cmd {
	case "init":
		err = l.Apply()
	case "apply":
		err = l.Apply()
	case "upgrade":
		err = l.Upgrade()
	case "cli":
		err = l.Cli()
	case "backup":
		err = r.Backup()
	case "restore":
		err = r.Restore()
	default:
		err = l.Start()
	}
	check(err)
}
