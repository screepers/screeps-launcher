package main

import (
	"github.com/screepers/screeps-launcher/v1/launcher"
	"log"
	"os"
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
	if len(os.Args) >= 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "init":
		err = l.Apply()
	case "apply":
		err = l.Apply()
	case "upgrade":
		err = l.Upgrade()
	case "cli":
		err = l.Cli()
	default:
		err = l.Start()
	}
	check(err)
}
