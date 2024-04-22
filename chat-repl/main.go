package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/ollama/ollama/api"
)

type Config struct {
	Model string
}

var (
	cfg = &Config{
		Model: "llama3",
	}
	messages = []api.Message{{
		Role:    "system",
		Content: "Keep responses brief. Responses should be formatted as markdown",
	}}
)

func init() {
	if model, exists := os.LookupEnv("GLOLLAMA_MODEL"); exists {
		cfg.Model = model
	}
}

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewScanner(os.Stdin)
	printPrompt()
	for reader.Scan() {
		messages = append(messages, api.Message{
			Role:    "user",
			Content: reader.Text(),
		})

		ctx := context.Background()
		req := &api.ChatRequest{
			Model:    "llama3",
			Messages: messages,
		}

		// since many responses will be returned, we append them here
		// so we can add them to the `messages` history later
		respAccumulator := ""

		respFunc := func(resp api.ChatResponse) error {
			respAccumulator = respAccumulator + resp.Message.Content
			out, err := glamour.Render(respAccumulator, "dark")
			if err != nil {
				return err
			}

			clearScreen()
			fmt.Print("\r" + out)
			return nil
		}

		err = client.Chat(ctx, req, respFunc)
		if err != nil {
			log.Fatal(err)
		}

		messages = append(messages, api.Message{
			Role:    "assistant",
			Content: respAccumulator,
		})

		printPrompt()
	}
}

func printPrompt() {
	fmt.Printf("%s > ", cfg.Model)
}

func clearScreen() {
	fmt.Fprint(os.Stderr, "\033[H\033[2J")
}
