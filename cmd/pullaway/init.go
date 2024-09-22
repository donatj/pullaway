package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/donatj/pullaway"
	"github.com/google/subcommands"
)

type initCmd struct {
	pc     *pullaway.PushoverClient
	config *Config
}

func (*initCmd) Name() string { return "init" }
func (*initCmd) Synopsis() string {
	return "Sign into Pushover and Register the application as a Device on your account"
}
func (c *initCmd) Usage() string {
	return c.Name() + "\n\t" + c.Synopsis() + "\n"
}

func (st *initCmd) SetFlags(f *flag.FlagSet) {
}

func (st *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var username, password, twofa string

	var secret string
loginLoop:
	for {
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Email").Value(&username),
			huh.NewInput().Title("Password").EchoMode(huh.EchoModePassword).Value(&password),
			huh.NewInput().Title("Two Factor Auth").Description("leave blank if not enabled").Value(&twofa),
		))

		err := form.Run()
		if err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}

		lr, err := st.pc.Login(username, password, twofa)
		if err != nil {
			log.Println(err)
			log.Println("Please try again.")
			continue loginLoop
		}

		secret = lr.Secret
		err = st.config.SetKey(ConfigUserSecret, secret)
		if err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}

		break
	}

	hostname, _ := os.Hostname()
	hostname = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(hostname, "-")

	shortname := fmt.Sprintf("pullaway-%s", hostname)
	shortname = shortname[:min(25, len(shortname))]

	form := huh.NewInput().Title("Device ShortName").Value(&shortname).CharLimit(25)

	err := form.Run()
	if err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	rr, err := st.pc.Register(secret, shortname)
	if err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	err = st.config.SetKey(ConfigDeviceID, rr.ID)
	if err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	log.Printf("Device Registered: %s\n", shortname)

	return subcommands.ExitSuccess
}
