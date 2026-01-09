package main

import (
	"Gator/internal/config"
	"errors"
	"fmt"
)

type state struct {
	data *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandNames map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("command 'arguments' list is empty")
	}
	err := s.data.SetUser(cmd.arguments[0])
	if err == nil {
		return err
	}
	fmt.Println("User name has been set")
	return nil
}

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
