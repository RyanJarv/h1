package h1

import (
	"net/http"
	"strings"
)

// NewHackeroneInput is the input parameters for NewHackerone.
//
// Username must be provided.
// Token is optional, if not provided it will be read from ~/.config/h1_token.
type NewHackeroneInput struct {
	Username string `json:"username"`

	Token string `json:"token"`
}

func NewHackerone(input *NewHackeroneInput) *Hackerone {
	if input.Token == "" {
		input.Token = GetH1Token()
	}

	return &Hackerone{
		username: input.Username,
		token:    strings.Trim(input.Token, " \t\n"),
		client:   http.DefaultClient,
	}
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type Hackerone struct {
	token    string
	username string
	client   Client
}
