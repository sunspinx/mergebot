package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/richardbizik/mergebot/internal/config"
)

var dg *discordgo.Session

func Init() {
	discord, err := discordgo.New("Bot " + config.DISCORD_TOKEN)
	if err != nil {
		panic(err)
	}
	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(onMessage)
	discord.AddHandler(onMessageReact)
	discord.AddHandler(onMessageReactRemove)

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsAll
	discord.StateEnabled = true

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	dg = discord
}

func Close() {
	// Cleanly close down the Discord session.
	dg.Close()
}
