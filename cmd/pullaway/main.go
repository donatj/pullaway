package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/donatj/pullaway"
	"github.com/google/subcommands"
)

func main() {
	pc := &pullaway.PushoverClient{}

	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	subcommands.Register(&initCmd{pc, cfg}, "initial setup")

	secret, err := cfg.GetKey(ConfigUserSecret)
	if err != nil {
		log.Fatalf("Error getting secret from config: %v", err)
	}

	deviceID, err := cfg.GetKey(ConfigDeviceID)
	if err != nil {
		log.Fatalf("Error getting device ID from config: %v", err)
	}

	var ac *pullaway.AuthorizedClient
	if secret != "" && deviceID != "" {
		ac = pullaway.NewAuthorizedClient(secret, deviceID)
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	subcommands.Register(&listenCmd{
		ac: ac,
		l:  l,
	}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
