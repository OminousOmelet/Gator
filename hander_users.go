package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OminousOmelet/Gator/internal/database"
	"github.com/google/uuid"
)

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

	// Potential unhandled error (fine??)
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

// Delete all users
func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.DeleteAll(ctx)
	if err != nil {
		return fmt.Errorf("unable to reset users: %v", err)
	}
	fmt.Println("all users have been deleted successfully!")
	return nil
}

// List all users
func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	userList, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch users: %v", err)
	}

	if len(userList) == 0 {
		fmt.Println("(no users)")
		return nil
	}
	for _, userName := range userList {
		currentFlag := ""
		if userName == s.cfg.CurrentUserName {
			currentFlag = " (current)"
		}
		fmt.Printf("* %s%s\n", userName, currentFlag)
	}
	return nil
}
