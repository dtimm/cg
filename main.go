package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sashabaranov/go-openai"
)

type options struct {
	APIToken    string `short:"t" long:"api-token" env:"OPENAI_API_TOKEN" description:"OpenAI API token"`
	Interactive bool   `short:"i" long:"interactive" description:"start in interactive mode"`
	Prompt      string `short:"p" long:"prompt" description:"prompt to pass to OpenAI"`
	// OutFile     string `short:"o" long:"out-file" description:"file to write chat to (append only)"`
}

func main() {
	var opt options
	flags.Parse(&opt)

	client := openai.NewClient(opt.APIToken)

	messages := []openai.ChatCompletionMessage{}

	if opt.Interactive {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Printf("user: ")
			m := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: scanner.Text(),
			}
			messages = append(messages, m)
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:    openai.GPT3Dot5Turbo,
					Messages: messages,
				})
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				break
			}
			messages = append(messages, resp.Choices[0].Message)
			fmt.Printf("agent: %s\n\n", resp.Choices[0].Message.Content)
		}
	} else {
		resp, err := single(client, opt.Prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}

		fmt.Println(resp)
	}

	fmt.Printf("\n\n\n%#v\n\n", messages)
	os.Exit(0)
}

func single(c *openai.Client, p string) (string, error) {
	resp, err := c.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleUser,
				Content: p,
			}},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error in CreateChatCompletion: %s", err)
	}

	return resp.Choices[0].Message.Content, nil
}
