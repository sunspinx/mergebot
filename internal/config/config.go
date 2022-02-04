package config

import (
	"os"
	"strconv"
	"strings"
)

var (
	DISCORD_TOKEN         string
	GITLAB_TOKEN          string
	GITLAB_HOST           string
	APPROVE_COUNT         int
	REVIEWER_COUNT        int
	MERGE_REQUEST_CHANNEL string
	CATEGORIES            []string
)

func Init() {
	GITLAB_TOKEN = os.Getenv("GITLAB_TOKEN")
	DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
	GITLAB_HOST = os.Getenv("GITLAB_HOST")

	ac := os.Getenv("APPROVE_COUNT")
	approveCount, err := strconv.Atoi(ac)
	if err != nil {
		panic("Could not parse APPROVE_COUNT")
	}
	APPROVE_COUNT = approveCount

	rc := os.Getenv("REVIEWER_COUNT")
	reviewerCount, err := strconv.Atoi(rc)
	if err != nil {
		panic("Could not parse REVIEWER_COUNT")
	}
	REVIEWER_COUNT = reviewerCount
	MERGE_REQUEST_CHANNEL = os.Getenv("MERGE_REQUEST_CHANNEL")
	cat := os.Getenv("CATEGORIES")
	CATEGORIES = strings.Split(cat, ",")
}
