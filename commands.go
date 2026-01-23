package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/OminousOmelet/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Pfffffffffffffff
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if len(s.cfg.CurrentUserName) < 1 {
			return fmt.Errorf("(no current user)")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error getting user '%s': %v", user.Name, err)
		}
		err = handler(s, cmd, user)
		return err
	}
}

// Called by main and then calls the appropriate function from the function map
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

// Formatting for unmarshalled feeds
func unEscapeRSS(rss *RSSFeed) {
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i, item := range rss.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rss.Channel.Item[i] = item
	}
}

// 'follow' handler helper function (also used by 'addfeed')
func follow(s *state, feedUrl string, user database.User) (database.Feed, error) {
	feed, err := s.db.GetFeed(context.Background(), feedUrl)
	if err != nil {
		return database.Feed{}, fmt.Errorf("error fetching feed: %v", err)
	}

	newFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), newFollowParams)
	if err != nil {
		return database.Feed{}, fmt.Errorf("error creating feed follow: %v", err)
	}
	return feed, nil
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("next-feed fetch failed: %v", err)
	}

	s.db.MarkFeedFetched(ctx, feed.ID)
	fetched, err := fetchFeed(ctx, feed.Url)

	if err != nil {
		return fmt.Errorf("fetch failed: %v", err)
	}
	for _, item := range fetched.Channel.Item {
		postParams := database.CreatePostsParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       ToNullString(item.Title),
			Url:         item.Link,
			Description: ToNullString(item.Description),
			PlubishedAt: item.PubDate,
			FeedID:      feed.ID,
		}
		_, err := s.db.CreatePosts(ctx, postParams)
		if err != nil {
			if err.Error()[:17] == "pq: duplicate key" {
				fmt.Println("(ignoring duplicate link)...")
			} else {
				fmt.Printf("error creating post: %v", err)
			}
			continue
		}

		// better console message for no-title posts
		var title string
		if item.Title == "" {
			title = "(no title)"
		} else {
			title = "'" + item.Title + "'"
		}
		fmt.Printf("%s posted...\n", title)
	}

	fmt.Printf("-- batch completed at %v--\n", time.Now())
	return nil
}

// String-to-sqlNULLstring converter
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
