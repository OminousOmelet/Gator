package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/OminousOmelet/Gator/internal/database"
	"github.com/google/uuid"
)

// Conitnuously aggregates feeds in a loop until task-killed
func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Usage: %s <time interval> (e.g. 20s, 5m, 1h, ...)", cmd.name)
	}

	time_between_reqs := cmd.arguments[0]
	interval, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("interval parsing failed: %v", err)
	}

	fmt.Printf("Collecting feeds every %s...\n\n", time_between_reqs)
	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

// Add feed record to database
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("Usage: %s <feed name> [feed url]", cmd.name)
	}
	feedName := cmd.arguments[0]
	feedUrl := cmd.arguments[1]

	newFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	}

	_, err := s.db.CreateFeed(context.Background(), newFeedParams)
	if err != nil {
		return fmt.Errorf("unable to create feed: %v", err)
	}

	// automatically follow the added feed
	feed, err := follow(s, feedUrl, user)
	if err != nil {
		return fmt.Errorf("follow failed: %v", err)
	}

	fmt.Println("feed record created successfully!")
	fmt.Printf("user info:\n- ID: %v\n- name: %s\n- url: %s\n- user ID: %v\n- created at: %s\n- updated at: %s\n",
		feed.ID, feed.Name, feed.Url,
		user.ID, feed.CreatedAt, feed.UpdatedAt)
	return nil
}

// List all feeds
func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feedList, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch feeds: %v", err)
	}

	if len(feedList) == 0 {
		fmt.Println("(no feeds)")
		return nil
	}
	for _, feed := range feedList {
		fmt.Printf("%s\n- url: %s\n- user: %s\n\n", feed.Name, feed.Url, feed.User)
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arguments) > 0 {
		validInt, err := strconv.ParseInt(cmd.arguments[0], 10, 0)
		if err != nil {
			return fmt.Errorf("Usage: %s <max number of posts to show> (default is 2)", cmd.name)
		}
		limit = int(validInt)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return fmt.Errorf("failed to get posts: %v", err)
	}

	if len(posts) < 1 {
		return fmt.Errorf("(no posts to show)")
	}
	if len(posts) < limit {
		limit = len(posts)
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("--%s--\n", posts[i].Title.String)
		fmt.Println(posts[i].Description.String)
		fmt.Printf("- publshed on %s\n\n", posts[i].PlubishedAt)
	}
	return nil
}
