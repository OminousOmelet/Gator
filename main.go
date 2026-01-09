package main

import (
	"Gator/internal/config"
	"fmt"
	"os"
)

func main() {
	cfg, err := config.Read()
	cfgState := state{cfg}

	cmds := commands{commandNames: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("too few arguments passed, exiting program")
		os.Exit(1)
	}

	_, exists := cmds.commandNames[os.Args[1]]
	if !exists {
		fmt.Println("invalid command")
		os.Exit(1)
	}

	if os.Args[1] == "login" && len(os.Args) < 3 {
		fmt.Println("must provide a username")
		os.Exit(1)
	}

	err = cmds.run(&cfgState, command{name: os.Args[1], arguments: os.Args[2:]})
	if err != nil {
		fmt.Println(err)
	}

}
