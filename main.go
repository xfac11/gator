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

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
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
