package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/donatj/pullaway"
	"github.com/gen2brain/beeep"
	"github.com/google/subcommands"
)

type listenCmd struct {
	ac *pullaway.AuthorizedClient
	l  pullaway.LeveledLogger

	format      string
	templateStr string
}

func (*listenCmd) Name() string     { return "listen" }
func (*listenCmd) Synopsis() string { return "listen for incoming messages" }
func (*listenCmd) Usage() string {
	return `listen:
	listen for incoming messages
`
}

func (st *listenCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&st.format, "format", "json", "Output format: json, text, template or notification")
	f.StringVar(&st.templateStr, "template", "", "Go template for formatting output (used with -format=template)")
}

func (st *listenCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if st.ac == nil {
		log.Println("No authorized client found. Please run 'init' first.")
		return subcommands.ExitFailure
	}

	// Initialize the display function based on the format
	displayFunc, err := st.initDisplayFunc()
	if err != nil {
		log.Printf("Error initializing display function: %v", err)
		return subcommands.ExitFailure
	}

	downloadAndDisplay := func() error {
		messages, _, err := st.ac.DownloadAndDeleteMessages()
		if err != nil {
			st.l.Error("error fetching messages", "error", err.Error())
			return nil
		}

		for _, m := range messages.Messages {
			if err := displayFunc(&m); err != nil {
				return err
			}
		}

		return nil
	}

	// ignore any initial errors, just start listening
	_ = downloadAndDisplay()

	// Start listening for new messages
	listener := st.ac.GetAuthorizedListener(st.l)

	err = listener.ListenWithReconnect(downloadAndDisplay)
	if err != nil {
		log.Printf("Error listening: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

// initDisplayFunc returns the appropriate display function based on the format
func (st *listenCmd) initDisplayFunc() (func(*pullaway.Messages) error, error) {
	switch st.format {
	case "json":
		return displayMessageJSON, nil
	case "text":
		return displayMessageText, nil
	case "notification":
		return displayMessageNotification, nil
	case "template":
		if st.templateStr == "" {
			return nil, fmt.Errorf("template string must be provided when format is 'template'")
		}
		// Compile the template once
		tmpl, err := template.New("output").Parse(st.templateStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing template: %w", err)
		}
		return displayMessageTemplate(tmpl), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", st.format)
	}
}

// displayMessageJSON outputs a single message in JSON format
func displayMessageJSON(m *pullaway.Messages) error {
	if err := json.NewEncoder(os.Stdout).Encode(m); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}
	return nil
}

// displayMessageText outputs a single message in a simple text format
func displayMessageText(m *pullaway.Messages) error {
	fmt.Printf("From %s: %s - %s", m.App, m.Title, m.Message)
	if m.URL != "" {
		fmt.Printf(" - URL: %s", m.URL)
	}
	fmt.Println()

	return nil
}

func displayMessageNotification(m *pullaway.Messages) error {
	beeep.Notify(fmt.Sprintf("%s: %s", m.App, m.Title), m.Message, "")

	return nil
}

// displayMessageTemplate returns a function that outputs a single message using the provided template
func displayMessageTemplate(tmpl *template.Template) func(*pullaway.Messages) error {
	return func(m *pullaway.Messages) error {
		if err := tmpl.Execute(os.Stdout, m); err != nil {
			return fmt.Errorf("error executing template: %w", err)
		}
		return nil
	}
}
