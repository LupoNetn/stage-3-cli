/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/luponetn/insighta-cli/cmd"
)

func main() {
	//load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	cmd.Execute()
}
