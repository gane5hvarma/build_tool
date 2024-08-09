package main

import (
	"fmt"
	"os"

	"github.com/gane5hvarma/build_tool/cmd"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("failed to read env file")
		os.Exit(1)
	}
	command := cmd.New()
	if err := command.Execute(); err != nil {
		fmt.Println("failed to execute command", err)
	}
}
