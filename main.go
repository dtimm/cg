package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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

	var err error

	client := openai.NewClient(opt.APIToken)

	if opt.Interactive {
		err = interactive(client, opt.OutFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}
	} else {
		resp, err := single(client, opt.Prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}

		fmt.Fprintf(os.Stdout, resp.Content)
		writeFile(opt.OutFile, resp.Content)
	}

	os.Exit(0)
}

func writeFile(outFile, text string) error {
	f, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("cannot open file: %s", err)
	}

	defer f.Close()
	fmt.Fprintln(f, text)

	return nil
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

func interactive(c *openai.Client, outFile string) error {
	messages := []openai.ChatCompletionMessage{}

	fmt.Print("user: ")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		var inputLines []string
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				break
			}
			inputLines = append(inputLines, line)
			t := scanner.Text()
			writeFile(outFile, fmt.Sprintf("user: %s", t))
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}

		t := strings.Join(inputLines, "\n")

		m := packageUserPrompt(t)
		messages = append(messages, m)
		resp, err := c.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			})
		if err != nil {
			return err
		}

		messages = append(messages, resp.Choices[0].Message)
		t = fmt.Sprintf("agent: %s\n", resp.Choices[0].Message.Content)
		fmt.Print(t)
		writeFile(outFile, t)
		fmt.Print("user: ")
	}
}
