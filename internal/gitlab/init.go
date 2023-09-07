package gitlab

import (
	"fmt"

	"github.com/sunspinx/mergebot/internal/config"
	"github.com/xanzy/go-gitlab"
)

var gl *gitlab.Client

func Init() {
	gitlabc, err := gitlab.NewClient(config.GITLAB_TOKEN, gitlab.WithBaseURL(fmt.Sprintf("https://%s", config.GITLAB_HOST)))
	if err != nil {
		fmt.Printf("Could not create gitlab client:\n")
		panic(err)
	}
	gl = gitlabc
}
