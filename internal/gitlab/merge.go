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
		return fmt.Errorf("could not parse project and mergeId from message")
	}

	info, err := getProjectInfo(project, merge)
	if err != nil {
		return err
	}
	if info.merged {
		return nil
	}

	t := true
	_, resp, err := gl.MergeRequests.AcceptMergeRequest(info.projectId, info.mergeId, &gitlab.AcceptMergeRequestOptions{
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
	if !info.pipelineOk {
		return custerror.GitlabError{
			Code:    custerror.PipelineNotOk,
			Message: "Pipeline not in success state",
		}
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

type projectInfo struct {
	projectId  int
	mergeId    int
	merged     bool
	pipelineOk bool
}

func getProjectInfo(project string, merge string) (info projectInfo, err error) {
	opts := &gitlab.GetMergeRequestsOptions{}
	info.mergeId, err = strconv.Atoi(merge)
	if err != nil {
		fmt.Println(err)
		err = fmt.Errorf("wrong project id")
		return
	}
	p := project
	fmt.Println(p)
	mr, _, err := gl.MergeRequests.GetMergeRequest(p, info.mergeId, opts)
	if mr.State == "merged" {
		info.merged = true
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	info.projectId = mr.ProjectID
	if mr.Pipeline.Status != "success" {
		info.pipelineOk = true
	}
	return
}
