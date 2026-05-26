package main

import (
	"fmt"
	"os"

	"github.com/xfac11/gator/internal/config"
)

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
