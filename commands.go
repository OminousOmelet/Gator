package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OminousOmelet/Gator/internal/config"
	"github.com/OminousOmelet/Gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
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
	userName := cmd.arguments[0]
	ctx := context.Background()

	user, err := s.db.GetUser(ctx, userName)
	if err != nil {
		return fmt.Errorf("user %s does not exist", userName)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("unable to set current user: %v", err)
	}
	fmt.Println("user switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.name)
	}
	userName := cmd.arguments[0]
	ctx := context.Background()

	existingUser, err := s.db.GetUser(ctx, userName)
	if err == nil && existingUser.Name == userName {
		return fmt.Errorf("user %s already exists", userName)
	}

	newUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}

	_, err = s.db.CreateUser(ctx, newUserParams)
	if err != nil {
		return fmt.Errorf("unable to create user: %v", err)
	}

	s.cfg.SetUser(userName)
	fmt.Println("user created successfully!")
	fmt.Printf("user info:\n- ID: %v\n- name: %s\n- created at: %s\n- updated at: %s\n", newUserParams.ID, newUserParams.Name, newUserParams.CreatedAt, newUserParams.UpdatedAt)

	return nil
}

func reset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.DeleteAll(ctx)
	if err != nil {
		return fmt.Errorf("unable to reset users: %v", err)
	}
	fmt.Println("all users have been deleted successfully!")
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
