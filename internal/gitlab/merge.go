package gitlab

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/richardbizik/mergebot/internal/config"
	"github.com/richardbizik/mergebot/internal/custerror"
	"github.com/xanzy/go-gitlab"
)

func MergeMessage(content string) error {
	project, merge := getMergeRequestParts(content)
	if project == "" || merge == "" {
		return fmt.Errorf("Could not parse project and mergeId from message")
	}

	projectId, mergeId, merged, err := getProjectInfo(project, merge)
	if err != nil {
		return err
	}
	if merged {
		return nil
	}

	t := true
	_, resp, err := gl.MergeRequests.AcceptMergeRequest(projectId, mergeId, &gitlab.AcceptMergeRequestOptions{
		ShouldRemoveSourceBranch:  &t,
		MergeWhenPipelineSucceeds: &t,
	})
	if err != nil {
		switch resp.StatusCode {
		case 405:
			return custerror.GitlabError{
				Code:    custerror.CouldNotMerge,
				Message: err.Error(),
			}
		case 406:
			return custerror.GitlabError{
				Code:    custerror.ConflictCannotMerge,
				Message: err.Error(),
			}
		}
		return err
	}
	return nil
}

func getMergeRequestParts(message string) (projectId string, mergeRequestId string) {
	var myExp = regexp.MustCompile(`https?:\/\/` + config.GITLAB_HOST + `\/(?P<project>.+)\/-\/merge_requests/(?P<merge>\d+).*`)

	match := myExp.FindStringSubmatch(message)
	result := make(map[string]string)
	for i, name := range myExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["project"], result["merge"]
}

func getProjectInfo(project string, merge string) (projectId int, mergeId int, merged bool, err error) {
	opts := &gitlab.GetMergeRequestsOptions{}
	mergeId, err = strconv.Atoi(merge)
	if err != nil {
		fmt.Println(err)
		err = fmt.Errorf("Wrong project id")
		return
	}
	p := project
	fmt.Println(p)
	mr, _, err := gl.MergeRequests.GetMergeRequest(p, mergeId, opts)
	if mr.State == "merged" {
		merged = true
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	projectId = mr.ProjectID
	return
}
