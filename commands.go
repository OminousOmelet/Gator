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

func agg(s *state, cmd command) error {
	ctx := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	rss, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	fmt.Println(rss)
	return nil
}

// add feed record to database
func handlerAddFeed(s *state, cmd command) error {
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

	fmt.Println("feed record created successfully!")
	fmt.Printf("user info:\n- ID: %v\n- name: %s\n- url: %s\n- user ID: %v\n- created at: %s\n- updated at: %s\n", newFeedParams.ID, newFeedParams.Name, newFeedParams.Url, newFeedParams.ID, newFeedParams.CreatedAt, newFeedParams.UpdatedAt)
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
