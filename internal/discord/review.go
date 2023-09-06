package discord

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"
	"regexp"
	"strings"
	"time"

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
	reviewerCount := config.REVIEWER_COUNT
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
	fmt.Printf("rolling from: %v", membersToRoll)

	var pickedMembers []string
	// check if we have enough members online
	if len(membersToRoll) < reviewerCount {
		membersToRoll = append(membersToRoll, offlineMembers...)
	}
	// pick members and replace message with mentions
	if len(membersToRoll) < reviewerCount {
		fmt.Printf("Not enough members online")
		return
	}
	if len(m.Mentions) != 0 {
		for _, v := range m.Mentions {
			if v != nil {
				pickedMembers = append(pickedMembers, v.ID)
				membersToRoll = removeFromArray(membersToRoll, v.ID)
				reviewerCount--
			}
		}
	}
	for i := 0; i < reviewerCount; i++ {
		var roll int
		if len(membersToRoll)-1 == 0 {
			roll = 0
			fmt.Println("roll with 0")
		} else {
			roll = getRandomNumber(len(membersToRoll) - i)
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
	fmt.Printf("Picked: %v\n", pickedMembers)
}

func getRandomNumber(max int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		mrand.Seed(time.Now().Unix())
		return mrand.Intn(max)
	}
	return int(r.Int64())
}

func removeFromArray(array []string, s string) []string {
	for i, v := range array {
		if v == s {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
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
