package discord

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/richardbizik/mergebot/internal/config"
	"github.com/richardbizik/mergebot/internal/custerror"
	"github.com/richardbizik/mergebot/internal/gitlab"
)

const (
	approve    string = "✅"
	disapprove string = "❌"
)

func onMessageReactRemove(dg *discordgo.Session, r *discordgo.MessageReactionRemove) {
	onReaction(dg, r.MessageReaction)
}

func onMessageReact(dg *discordgo.Session, r *discordgo.MessageReactionAdd) {
	onReaction(dg, r.MessageReaction)
}

func onReaction(dg *discordgo.Session, r *discordgo.MessageReaction) {
	m, err := dg.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !isMessageOfInterest(m) {
		return
	}
	channel, category, err := getMessageChannelAndCategory(dg, m)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !isChannelOfInterest(channel, category) {
		return
	}
	approveCount := 0
	for _, mr := range m.Reactions {
		if mr.Emoji.Name == approve {
			approveCount += mr.Count
		} else if mr.Emoji.Name == disapprove {
			approveCount -= mr.Count
		}
	}
	if approveCount >= config.APPROVE_COUNT {
		err := gitlab.MergeMessage(m.Content)
		var newMessage string
		oldMessage := getCleanedMessage(m.Content)
		//check if error is due to conflict/failed pipeline etc.. and update status
		if err != nil {
			if errors.Is(err, custerror.GitlabError{}) {
				ce := err.(custerror.GitlabError)
				switch ce.Code {
				case custerror.CouldNotMerge:
					newMessage = fmt.Sprintf("%s❌ Status: %s", oldMessage, "Cannot merge check for pipeline failures")
				case custerror.ConflictCannotMerge:
					newMessage = fmt.Sprintf("%s❌ Status: %s", oldMessage, "Cannot merge check for conflicts")
				}
			} else {
				newMessage = fmt.Sprintf("%s❌ Status: %s", oldMessage, err)
			}
		} else {
			newMessage = fmt.Sprintf("%s✅ Status: %s", oldMessage, "Merged")
		}
		_, err = dg.ChannelMessageEdit(m.ChannelID, m.ID, newMessage)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getCleanedMessage(message string) string {
	var s string
	scanner := bufio.NewScanner(strings.NewReader(message))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "Status:") {
			s += fmt.Sprintf("%s\n", line)
		}
	}
	return s
}
