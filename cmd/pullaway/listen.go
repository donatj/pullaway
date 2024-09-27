package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/donatj/pullaway"
	"github.com/google/subcommands"
)

type listenCmd struct {
	ac *pullaway.AuthorizedClient
}

func (*listenCmd) Name() string     { return "listen" }
func (*listenCmd) Synopsis() string { return "listen for incoming messages" }
func (*listenCmd) Usage() string {
	return `listen:
	listen for incoming messages
`
}

func (st *listenCmd) SetFlags(f *flag.FlagSet) {
}

func (st *listenCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if st.ac == nil {
		log.Println("No authorized client found. Please run `init` first.")
		return subcommands.ExitFailure
	}

	downloadAndDisplay := func() error {
		messages, _, err := st.ac.DownloadAndDeleteMessages()
		if err != nil {
			return err
		}

		displayMessages(messages)

		return nil
	}

	downloadAndDisplay()

	err := pullaway.ListenWithReconnect(st.ac.DeviceID, st.ac.UserSecret, downloadAndDisplay)
	if err != nil {
		log.Printf("Error listening: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func displayMessages(messages *pullaway.DownloadResponse) error {
	for _, m := range messages.Messages {
		// fmt.Printf("ID: %d\n", m.ID)
		// fmt.Printf("IDStr: %s\n", m.IDStr)
		// fmt.Printf("Message: %s\n", m.Message)
		// fmt.Printf("App: %s\n", m.App)
		// fmt.Printf("Aid: %d\n", m.Aid)
		// fmt.Printf("AidStr: %s\n", m.AidStr)
		json.NewEncoder(os.Stdout).Encode(m)
	}

	return nil
}
