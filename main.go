package main

import (
	"log"
	"os"

	"database/sql"

	"github.com/OminousOmelet/gator/internal/config"
	"github.com/OminousOmelet/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	// Removes timestamp from error messages
	log.SetFlags(0)

	cfg, err := config.Read()
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	gatorState := state{database.New(db), cfg}

	cmds := commands{commandNames: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("agg", handlerAgg)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	_, exists := cmds.commandNames[os.Args[1]]
	if !exists {
		log.Fatal("invalid command")
	}

	err = cmds.run(&gatorState, command{name: os.Args[1], arguments: os.Args[2:]})
	if err != nil {
		log.Fatal(err)
	}
}
