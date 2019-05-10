package main

import (
	"log"
	"os"

	"github.com/screepers/screeps-launcher/v1/launcher"
	"github.com/screepers/screeps-launcher/v1/version"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error
	l := launcher.Launcher{}
	l.Prepare()
	cmd := "start"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	log.Printf("screeps-launcher %s (%s)", version.BuildVersion, version.BuildTime)
	switch cmd {
	case "init":
		err = l.Apply()
	case "apply":
		err = l.Apply()
	case "upgrade":
		err = l.Upgrade()
	case "cli":
		err = l.Cli()
	case "version":
	default:
		err = l.Start()
	}
	check(err)
}
