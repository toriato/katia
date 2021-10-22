package main

import (
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/toriato/katia"
)

const token = "yea i accidentally leak my token and i feel dumb af :)"

func main() {
	bot, err := katia.New(token)
	if err != nil {
		panic(err)
	}

	bot.Session.AddHandler(func(_ *discordgo.Session, e *discordgo.Ready) {
		bot.Logger.
			WithField("user", bot.Session.State.User).
			Infof("Logged in as %s", bot.Session.State.User)
	})

	if err := bot.Session.Open(); err != nil {
		bot.Logger.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	bot.Logger.Info("gracefully shutdown")

	if err := bot.Session.Close(); err != nil {
		bot.Logger.Fatal(err)
	}
}
