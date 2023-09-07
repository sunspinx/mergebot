package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sunspinx/mergebot/internal/config"
	"github.com/sunspinx/mergebot/internal/discord"
	"github.com/sunspinx/mergebot/internal/gitlab"
)

func main() {
	config.Init()
	gitlab.Init()
	discord.Init()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}
