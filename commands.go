package main

import (
	"fmt"

	"github.com/OminousOmelet/Gator/internal/config"
)

type state struct {
	//db  *database.Queries
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandNames map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.name)
	}
	err := s.cfg.SetUser(cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("unable to login: %v", err)
	}
	fmt.Println("user switched successfully!")
	return nil
}

// func handlerRegister(s *state, cmd command) error {
// 	if len(cmd.arguments) != 1 {
// 		return fmt.Errorf("Usage: %s <name>", cmd.name)
// 	}
// 	userName := cmd.arguments[0]

// 	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
// 		ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: userName})
// 	if err != nil {
// 		return fmt.Errorf("create user failed: %v", err)
// 	}

// 	s.cfg.SetUser(userName)
// 	fmt.Println("user successfully created!")
// 	log.Printf("new user: %v", user)
// 	return nil
// }

func (c *commands) run(s *state, cmd command) error {
	err := c.commandNames[cmd.name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandNames[name] = f
}
