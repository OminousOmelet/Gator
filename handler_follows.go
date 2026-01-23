package main

import (
	"context"
	"fmt"

	"github.com/OminousOmelet/Gator/internal/database"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <feed link>", cmd.name)
	}

	feedUrl := cmd.arguments[0]
	feed, err := follow(s, feedUrl, user)
	if err != nil {
		return fmt.Errorf("follow failed: %v", err)
	}

	fmt.Println("feed follow created successfully!")
	fmt.Printf("- feed name: %s\n- followed by: %s", feed.Name, s.cfg.CurrentUserName)
	return nil
}

// List all follows for current user
func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error fetching follows: %v", err)
	}

	fmt.Printf("User '%s' is following:\n", user.Name)
	for _, follow := range follows {
		fmt.Printf("- %s\n", follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <feed link>", cmd.name)
	}

	ctx := context.Background()
	feedUrl := cmd.arguments[0]
	feed, err := s.db.GetFeed(ctx, feedUrl)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	err = s.db.DeleteFollow(ctx, database.DeleteFollowParams{UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("error removing follow: %v", err)
	}

	fmt.Printf("successfully unfollowed feed named '%s'\n", feed.Name)
	return nil
}
