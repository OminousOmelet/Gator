package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/OminousOmelet/Gator/internal/database"
	"github.com/google/uuid"
)

// Fetch an RSS feed from url, and unmarshal it into an RSSFeed struct
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error building http request: %v", err)
	}
	req.Header.Set("User-Agent", "gator")

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending http request: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching feed: status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	var rss RSSFeed
	xml.Unmarshal(data, &rss)
	unEscapeRSS(&rss)

	return &rss, nil
}

func unEscapeRSS(rss *RSSFeed) {
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i, item := range rss.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rss.Channel.Item[i] = item
	}
}

// Feed aggregator
func handlerAgg(s *state, cmd command) error {
	ctx := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	rss, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	fmt.Println(rss)
	return nil
}

// Add feed record to database
func handlerAddFeed(s *state, cmd command) error {
	if len(s.cfg.CurrentUserName) < 1 {
		return fmt.Errorf("(no current user)")
	}
	userName := s.cfg.CurrentUserName
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("Usage: %s <feed name> [feed url]", cmd.name)
	}
	feedName := cmd.arguments[0]
	feedUrl := cmd.arguments[1]
	ctx := context.Background()

	user, err := s.db.GetUser(ctx, userName)
	if err != nil {
		return fmt.Errorf("error getting user '%s': %v", userName, err)
	}

	newFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	}

	_, err = s.db.CreateFeed(ctx, newFeedParams)
	if err != nil {
		return fmt.Errorf("unable to create feed: %v", err)
	}

	// automatically follow the added feed
	feed, err := follow(s, feedUrl)
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
