# Matrix Bot

This package contains an extremely simple Matrix Bot framework.

Usage is fairly simple:

```go
package main

import (
	"log"
	"os"

	bot "gitlab.com/silkeh/matrix-bot"
)

func main() {
	// Configure the bot so that it only listens for commands prefixed by `!bot` (and highlights). 
	config := &bot.ClientConfig{CommandPrefixes: []string{"!bot "}}

	// Create a client based on environment variables
	client, err := bot.NewClient(os.Getenv("MATRIX_URL"), os.Getenv("MATRIX_UID"), os.Getenv("MATRIX_TOKEN"), config)
	if err != nil {
		log.Fatal(err)
	}

	// Set a command for the bot.
	client.SetCommand("test", &bot.Command{
		Summary:        "Test command",
		Description:    "This shows the description",
		MessageHandler: nil,
		Subcommands: map[string]*bot.Command{
			"sub": &bot.Command{
				Summary:        "Test subcommand",
				Description:    "This shows the description of the subcommand",
				MessageHandler: nil,
			},
		},
	})

	// Run the bot
	log.Fatal(client.Run())
}
```
