package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OminousOmelet/Gator/internal/database"
	"github.com/google/uuid"
)

func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <feed link>", cmd.name)
	}

	feedUrl := cmd.arguments[0]
	feed, err := follow(s, feedUrl)
	if err != nil {
		return fmt.Errorf("follow failed: %v", err)
	}

	fmt.Println("feed follow created successfully!")
	fmt.Printf("- feed name: %s\n- followed by: %s", feed.Name, s.cfg.CurrentUserName)
	return nil
}

// creates a follow record, and returns it so callers can display relevant info
func follow(s *state, feedUrl string) (database.Feed, error) {
	if len(s.cfg.CurrentUserName) < 1 {
		return database.Feed{}, fmt.Errorf("(no current user)")
	}

	ctx := context.Background()
	feed, err := s.db.GetFeed(ctx, feedUrl)
	if err != nil {
		return database.Feed{}, fmt.Errorf("error fetching feed: %v", err)
	}
	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return database.Feed{}, fmt.Errorf("error fetching current user: %v", err)
	}

	newFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(ctx, newFollowParams)
	if err != nil {
		return database.Feed{}, fmt.Errorf("error creating feed follow: %v", err)
	}
	return feed, nil
}

// List all follows for current user
func handlerFollowing(s *state, cmd command) error {
	if len(s.cfg.CurrentUserName) < 1 {
		return fmt.Errorf("(no current user)")
	}
	userName := s.cfg.CurrentUserName
	ctx := context.Background()
	follows, err := s.db.GetFeedFollowsForUser(ctx, userName)
	if err != nil {
		return fmt.Errorf("error fetching follows: %v", err)
	}

	fmt.Printf("User '%s' is following:\n", userName)
	for _, follow := range follows {
		fmt.Printf("- %s\n", follow.FeedName)
	}
	return nil
}
