package discord

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/richardbizik/mergebot/internal/config"
)

func onMessage(dg *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == dg.State.User.ID {
		return
	}
	if !isMessageOfInterest(m.Message) {
		return
	}

	channel, category, err := getMessageChannelAndCategory(dg, m.Message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(channel.Name, category.Name)

	if !isChannelOfInterest(channel, category) {
		return
	}
	sendMentionToReviewers(dg, category.Name, channel.GuildID, m)
}

func sendMentionToReviewers(dg *discordgo.Session, role string, guildId string, m *discordgo.MessageCreate) {
	guild, err := dg.State.Guild(guildId)
	if err != nil {
		fmt.Println(err)
		return
	}
	roles := guild.Roles
	//find role we are looking for - must be same as category in which there is merge-requests channel
	var roleId string
	for _, r := range roles {
		if r.Name == role {
			roleId = r.ID
			break
		}
	}
	if roleId == "" {
		fmt.Printf("Could not find role: %v\n", role)
		return
	}

	members := guild.Members
	var membersToRoll = make([]string, 0)
	var offlineMembers = make([]string, 0)
	for _, member := range members {
		if member.User.ID == m.Author.ID {
			continue
		}
		for _, r := range member.Roles {
			if r == roleId {
				presence, err := dg.State.Presence(guildId, member.User.ID)
				if err != nil {
					fmt.Println(err)
				}
				if presence != nil && presence.Status != discordgo.StatusOffline {
					membersToRoll = append(membersToRoll, member.User.ID)
				} else {
					offlineMembers = append(offlineMembers, member.User.ID)
				}
			}
		}
	}

	//check if we have enough members online
	var pickedMembers []string
	if len(membersToRoll) < config.REVIEWER_COUNT {
		membersToRoll = append(membersToRoll, offlineMembers...)
	}
	//pick two members and replace message with message with mentions
	if len(membersToRoll) < config.REVIEWER_COUNT {
		fmt.Printf("Not enough members online")
		return
	}
	for i := 0; i < config.REVIEWER_COUNT; i++ {
		var roll int
		if len(membersToRoll)-i == 0 {
			roll = 0
		} else {
			roll = rand.Intn(len(membersToRoll) - i)
		}
		pickedMembers = append(pickedMembers, membersToRoll[roll])
		membersToRoll = append(membersToRoll[:roll], membersToRoll[roll+1:]...)
	}

	newMessage := fmt.Sprintf("Review requested by <@!%s>: %s\nReviewers: ", m.Author.ID, m.Content)
	for _, pick := range pickedMembers {
		newMessage += fmt.Sprintf(" <@!%s>", pick)
	}
	dg.ChannelMessageDelete(m.ChannelID, m.ID)
	newMessage = replaceLinkWithoutEmbed(newMessage)
	_, err = dg.ChannelMessageSend(m.ChannelID, newMessage)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v\n", membersToRoll)
}

func replaceLinkWithoutEmbed(message string) string {
	var myExp = regexp.MustCompile(`(?P<link>https?:\/\/` + config.GITLAB_HOST + `\/.*?\/merge_requests\/\d+)`)

	match := myExp.FindStringSubmatch(message)
	result := make(map[string]string)
	if len(match) == 0 {
		return message
	}
	for i, name := range myExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	link := result["link"]
	return strings.ReplaceAll(message, link, fmt.Sprintf("<%s>", link))
}
