package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/xfac11/gator/internal/config"
	"github.com/xfac11/gator/internal/database"
	rssfeed "github.com/xfac11/gator/internal/feed"
)

type state struct {
	db            *database.Queries
	configHandler *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	plan map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	cmdFunc, ok := c.plan[cmd.name]
	if !ok {
		return fmt.Errorf("No command registered with that name")
	}
	return cmdFunc(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.plan[name] = f
}
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUserName := s.configHandler.CurrentUserName
		currentUser, err := s.db.GetUser(context.Background(), currentUserName)
		if err != nil {
			return fmt.Errorf("Could not retrieve user. Error: %w", err)
		}
		return handler(s, cmd, currentUser)
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	var err error
	if len(cmd.arguments) == 1 {
		limit, err = strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return fmt.Errorf("Non valid 'limit' parameter. Error: %w", err)
		}
	}
	postsArg := database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), postsArg)
	if err != nil {
		return fmt.Errorf("Could not retrieve posts for current user. Error: %w", err)
	}
	fmt.Println("Showing posts, limit:", limit)
	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.PublishedAt)
		fmt.Println(post.Url)
		fmt.Println(post.Description)
		fmt.Println()
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Excepts one argument (url) but %d was given", len(cmd.arguments))
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("Could not find that feed with the given url")
	}

	deleteParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), deleteParams)
	if err != nil {
		return fmt.Errorf("No feed with that user combination exists")
	}

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Could not retrieve user following. Error: %w", err)
	}

	for _, feed := range following {
		fmt.Println(feed)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Excepts one argument (url) but %d was given", len(cmd.arguments))
	}
	url := cmd.arguments[0]

	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Could not retrieve the feed given that url: %s", url)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("Could not create feedfollow. Error: %w", err)
	}

	fmt.Println(feedFollow.FeedName, feedFollow.UserName)
	return nil
}
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Could not retrieve all feeds from the database. Error: %w", err)
	}

	fmt.Println("|", "Name", "|", "URL", "|", "Username", "|")
	for _, feed := range feeds {
		fmt.Println("|", feed.Name, "|", feed.Url, "|", feed.UserName, "|")
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("Excepts two arguments (name, url) but %d was given", len(cmd.arguments))
	}
	name := cmd.arguments[0]
	url := cmd.arguments[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("Could not create a feed in the database. Error: %w", err)
	}

	arg := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), arg)
	if err != nil {
		return fmt.Errorf("Could not create feed follow. Error: %w", err)
	}
	fmt.Println("{")
	fmt.Println(" id:", feed.ID)
	fmt.Println(" created_at:", feed.CreatedAt)
	fmt.Println(" updated_at:", feed.UpdatedAt)
	fmt.Println(" name:", feed.Name)
	fmt.Println(" url:", feed.Url)
	fmt.Println(" user_id:", feed.UserID)
	fmt.Println("}")

	return nil
}
func handlerAgg(s *state, cmd command) error {

	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Excepts one argument (time between requests) but %d was given.", len(cmd.arguments))
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("Non valid time given")
	}
	fmt.Println("Collecting feeds every", cmd.arguments[0])

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return fmt.Errorf("Could not scrape feed. Error: %w", err)
		}
	}
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Could not retrieve users. Error: %w", err)
	}
	for _, user := range users {
		text := "* "
		text += user.Name
		if user.Name == s.configHandler.CurrentUserName {
			text += " (current)"
		}
		fmt.Println(text)
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset and remove all users. error: %w", err)
	}
	fmt.Println("Successfully removed all users")
	return nil
}
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("error: Excepts one username as argument")
	}
	_, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("No user is registered with that name")
	}
	err = s.configHandler.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set to:", cmd.arguments[0])
	return nil
}
func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("error: Excepts one name as argument")
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
	}
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("User with that name already exists, Error: %w", err)
	}
	err = s.configHandler.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Could not set config, error: %w", err)
	}

	fmt.Println("User was created with name", user.Name)
	fmt.Println("id:", user.ID)
	fmt.Println("created_at:", user.CreatedAt)
	fmt.Println("updated_at:", user.UpdatedAt)
	fmt.Println("name:", user.Name)

	return nil
}

func parseRSSTimeString(date string) (time.Time, error) {
	layouts := []string{time.RFC1123Z, time.RFC1123}

	var err error
	for _, layout := range layouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse date/time string. Error: %w", err)
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Could not retrieve next feed from db. Error: %w", err)
	}

	markArgs := database.MarkFeedFetchedParams{
		ID: feed.ID,
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}
	err = s.db.MarkFeedFetched(context.Background(), markArgs)
	if err != nil {
		return fmt.Errorf("Could not mark feed as fetched. Error: %w", err)
	}

	rssFeed, err := rssfeed.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("Could not fetch feed with the url: %s. Error: %w", feed.Url, err)
	}
	for _, item := range rssFeed.Channel.Item {
		pubAt, err := parseRSSTimeString(item.PubDate)
		if err != nil {
			return fmt.Errorf("Could not parse date. Error: %w", err)
		}
		postParam := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: pubAt,
			FeedID:      feed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), postParam)
		if err != nil {
			if pqerr, ok := err.(*pq.Error); ok {
				if pqerr.Code.Name() == "unique_violation" {
					continue
				} else {
					fmt.Println("Error creating post. Error: %w", err)
				}
			}
		}
	}
	return nil
}
func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Could not read config. Error: %w", err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	newState := state{
		configHandler: &cfg,
		db:            dbQueries,
	}

	commands := commands{
		plan: map[string]func(*state, command) error{},
	}

	commands.register("users", handlerUsers)
	commands.register("reset", handlerReset)
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))
	if len(os.Args) < 2 {
		fmt.Println("No command given")
		for key := range commands.plan {
			fmt.Println(key)
		}
		os.Exit(1)
	}

	name := os.Args[1]
	arguments := os.Args[2:]

	cmd := command{
		name:      name,
		arguments: arguments,
	}
	err = commands.run(&newState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
