package coach

import "errors"

var (
	errUnknownTool          = errors.New("unknown tool")
	errMalformedToolArgs    = errors.New("malformed tool arguments")
	errEmptyAssistantAnswer = errors.New("empty assistant answer")
)
