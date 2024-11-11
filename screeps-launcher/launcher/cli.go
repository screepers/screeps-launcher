package launcher

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/screepers/screeps-launcher/v1/cli"
)

var (
	hostFlag = flag.String("host", "", "CLI Host")
	portFlag = flag.Int("port", 0, "CLI Port")
	userFlag = flag.String("username", "", "CLI Username")
	passFlag = flag.String("password", "", "CLI Password")
)

func runCli(config *Config) error {
	host := config.Cli.Host
	port := config.Cli.Port
	user := config.Cli.Username
	pass := config.Cli.Password
	if v, ok := config.Env.Backend["CLI_HOST"]; ok {
		host = v
	}
	if v, ok := config.Env.Backend["CLI_PORT"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			port = int16(i)
		}
	}
	if *hostFlag != "" {
		host = *hostFlag
	}
	if *portFlag != 0 {
		port = int16(*portFlag)
	}
	if *userFlag != "" {
		user = *userFlag
	}
	if *passFlag != "" {
		pass = *passFlag
	}
	c := cli.NewScreepsCLI(host, port, user, pass)

	if err := c.Start(); err != nil {
		log.Fatalf("Error! %v", err)
	}
	fmt.Println(c.WelcomeText)
	p := prompt.New(
		func(cmd string) {
			if cmd == "exit" || cmd == "quit" {
				os.Exit(0)
			}
			ret := c.Command(cmd)
			if len(ret) > 0 {
				fmt.Println(ret)
			}
		},
		completer,
		prompt.OptionTitle("Screeps CLI"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionCompletionWordSeparator("."),
	)
	p.Run()
	return nil
}

func completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}
	parts := strings.Split(d.TextBeforeCursor(), ".")
	w := parts[len(parts)-1]
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
	} else {
		parts = []string{}
	}
	prefix := strings.Join(parts, ".")
	// w := d.GetWordBeforeCursor()
	log.Print(prefix, w)

	completions := map[string][]prompt.Suggest{
		"": []prompt.Suggest{
			{Text: "help()", Description: "Print Help"},
			{Text: "storage", Description: "A global database storage object"},
			{Text: "map", Description: "Map editing functions."},
			{Text: "bots", Description: "Manage NPC bot players and their AI scripts."},
			{Text: "system", Description: "System utility functions."},
		},
		"storage": []prompt.Suggest{
			{Text: "db", Description: "An object containing all database collections in it. Use it to fetch or modify game objects. The database is based on LokiJS project, so you can learn more about available functionality in its documentation"},
			{Text: "env", Description: "A simple key-value storage with an interface based on Redis syntax."},
			{Text: "pubsub", Description: "A Pub/Sub mechanism allowing to publish events across all processes."},
		},
		"map": []prompt.Suggest{
			{Text: "generateRoom(roomName, [opts])", Description: "Generate a new room at the specified location."},
			{Text: "openRoom(roomName, [timestamp])", Description: "Make a room available for use. Specify a timestamp in the future if you want it to be opened later automatically."},
			{Text: "closeRoom(roomName)", Description: "Make a room not available."},
			{Text: "removeRoom(roomName)", Description: "Delete the room and all room objects from the database."},
			{Text: "updateRoomImageAssets(roomName)", Description: "Update images in assets folder for the specified room."},
			{Text: "updateTerrainData()", Description: "Update cached world terrain data."},
		},
		"bots": []prompt.Suggest{
			{Text: "spawn(botAiName, roomName, [opts])", Description: "Create a new NPC player with bot AI scripts, and spawn it to the specified room."},
			{Text: "reload(botAiName)", Description: "Reload scripts for the specified bot AI."},
			{Text: "removeUser(username)", Description: "Delete the specified bot player and all its game objects."},
		},
		"strongholds": []prompt.Suggest{
			{Text: "spawn(roomName, [opts])", Description: "Create a new NPC Stronghold, and spawn it to the specified room."},
			{Text: "expand(roomName)", Description: "Force an NPC Stronghold to spawn a new lesser Invader Core in a nearby room."},
		},
		"system": []prompt.Suggest{
			{Text: "resetAllData()", Description: "Wipe all world data and reset the database to the default state."},
			{Text: "sendServerMessage(message)", Description: "Send a text server message to all currently connected players."},
			{Text: "pauseSimulation()", Description: "Stop main simulation loop execution."},
			{Text: "resumeSimulation()", Description: "Resume main simulation loop execution."},
			{Text: "runCronjob(jobName)", Description: "Run a cron job immediately."},
			{Text: "getTickDuration()", Description: "Show current minimal tick duration (in milliseconds)."},
			{Text: "setTickDuration(minimalDuration)", Description: "Set current minimal tick duration (in milliseconds)."},
		},
	}
	if part, ok := completions[prefix]; ok {
		return prompt.FilterHasPrefix(part, w, true)
	}
	return []prompt.Suggest{}
}
