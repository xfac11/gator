package main

import (
	"fmt"
	"os"

	"github.com/xfac11/gator/internal/config"
)

type state struct {
	configHandler *config.Config
}

type command struct {
	name      string
	arguments []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("Argument slice empty. Excepts one argument")
	}
	err := s.configHandler.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set to:", cmd.arguments[0])
	return nil
}
func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Could not read config. Error: %w", err)
		os.Exit(1)
	}
	name := "filip"
	err = cfg.SetUser(name)
	if err != nil {
		fmt.Printf("Could not set user in config to %s. Error: /n", name)
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Println("Could not read config. Error: %w", err)
		os.Exit(1)
	}

	fmt.Println("Database url:", cfg.DatabaseUrl, "Username:", cfg.CurrentUserName)
}
