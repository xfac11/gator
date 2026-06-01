package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/xfac11/gator/internal/config"
	"github.com/xfac11/gator/internal/database"
	"github.com/xfac11/gator/internal/feed"
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("Excepts one argument (url) but %d was given", len(cmd.arguments))
	}
	url := cmd.arguments[0]

	user, err := s.db.GetUser(context.Background(), s.configHandler.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Could not retrieve current user from the database")
	}

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

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("Excepts two arguments (name, url) but %d was given", len(cmd.arguments))
	}
	name := cmd.arguments[0]
	url := cmd.arguments[1]

	currentUserName := s.configHandler.CurrentUserName
	currentUser, err := s.db.GetUser(context.Background(), currentUserName)
	if err != nil {
		return fmt.Errorf("Could not retrieve the current user. Error: %w", err)
	}
	userId := currentUser.ID

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    userId,
	}
	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("Could not create a feed in the database. Error: %w", err)
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
	url := "https://www.wagslane.dev/index.xml"

	rssFeed, err := feed.FetchFeed(context.Background(), url)

	if err != nil {
		return fmt.Errorf("Could not fecth feed. Error: %w", err)
	}

	fmt.Println(rssFeed.Channel.Title)
	fmt.Println(rssFeed.Channel.Link)
	fmt.Println(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Item {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.PubDate)
		fmt.Println(item.Description)
	}
	return nil
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
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)
	if len(os.Args) < 2 {
		fmt.Println("No command given")
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
