package discord

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/richardbizik/mergebot/internal/config"
)

//Checks if message contains gitlab merge request link
func isMessageOfInterest(m *discordgo.Message) bool {
	var myExp = regexp.MustCompile(`(?P<merge_request>https?:\/\/` + config.GITLAB_HOST + `\/.*?\/merge_requests\/\d+)[\s|#]?`)

	match := myExp.FindStringSubmatch(m.Content)
	return len(match) != 0
}

//Checks if channel is for merge requests and its category
func isChannelOfInterest(channel *discordgo.Channel, category *discordgo.Channel) bool {
	return channel.Name == config.MERGE_REQUEST_CHANNEL &&
		constainsString(config.CATEGORIES, category.Name)
}

func getMessageChannelAndCategory(dg *discordgo.Session, m *discordgo.Message) (channel *discordgo.Channel, category *discordgo.Channel, err error) {
	channel, err = dg.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println(err)
	}
	category, err = dg.State.Channel(channel.ParentID)
	if err != nil {
		fmt.Println(err)
	}
	return channel, category, err
}

func constainsString(arr []string, s string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}
