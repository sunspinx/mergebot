package custerror

import "fmt"

type GEType string

const (
	CouldNotMerge       GEType = "COULD_NOT_MERGE"
	ConflictCannotMerge GEType = "CONFLICTS_CANNOT_MERGE"
)

type GitlabError struct {
	Code    GEType
	Message string
}

func (e GitlabError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e GitlabError) Is(target error) bool {
	_, ok := target.(GitlabError)
	return ok
}
