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
	OutFile     string `short:"o" long:"out-file" description:"file to write chat to (append only)"`
}

func main() {
	var opt options
	flags.Parse(&opt)

	client := openai.NewClient(opt.APIToken)

	if opt.OutFile != "" {
		f, err := os.Stat(opt.OutFile)
	}

	printIt := func(s string) {
		fmt.Println(s)

	}
	var messages []openai.ChatCompletionMessage
	var err error

	if opt.Interactive {
		messages, err = interactive(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}
	} else {
		resp, err := single(client, opt.Prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}

		messages = []openai.ChatCompletionMessage{resp}
	}

	fmt.Printf("\n\n\n%#v\n\n", messages)
	os.Exit(0)
}

func packageUserPrompt(p string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: p,
	}
}

func single(c *openai.Client, p string) (openai.ChatCompletionMessage, error) {
	m := packageUserPrompt(p)
	resp, err := c.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{m},
		},
	)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("error in CreateChatCompletion: %s", err)
	}

	return resp.Choices[0].Message, nil
}

func interactive(c *openai.Client) ([]openai.ChatCompletionMessage, error) {
	messages := []openai.ChatCompletionMessage{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("user: ")
		m := packageUserPrompt(scanner.Text())
		messages = append(messages, m)
		resp, err := c.CreateChatCompletion(
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
	return messages, nil
}
