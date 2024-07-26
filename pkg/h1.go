package lib

import (
	"encoding/json"
	"fmt"
	h1Types "github.com/ryanjarv/h1/pkg/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// NewHackeroneInput is the input parameters for NewHackerone.
//
// Username must be provided.
// Token is optional, if not provided it will be read from ~/.config/h1_token.
type NewHackeroneInput struct {
	Username string

	Token string
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

func (h1 *Hackerone) Program(id string) *Program {
	return &Program{
		Hackerone: *h1,
		ProgramId: id,
	}
}

type Program struct {
	Hackerone
	ProgramId string `json:"id"`
}

func (h1 *Program) GetDetail() (*h1Types.ProgramDetail, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s", h1.ProgramId)

	resp, uri, err := h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Program.GetDetails: getting program: %w", err)
	} else if uri != "" {
		return nil, fmt.Errorf("failed to get all pages: %s", uri)
	}

	program := h1Types.ProgramDetail{}
	err = json.Unmarshal(resp, &program)
	if err != nil {
		return nil, fmt.Errorf("Program.GetDetails: failed to unmarshal program: %w", err)
	}

	return &program, nil
}

func (h1 *Program) GetWeaknesses() (*h1Types.Weaknesses, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s/weaknesses", h1.ProgramId)

	var resp []byte
	var err error

	resp, uri, err = h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", uri, err)
	} else if uri != "" {
		return nil, fmt.Errorf("failed to get all pages")
	}

	weaknesses := h1Types.Weaknesses{}
	err = json.Unmarshal(resp, &weaknesses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal program: %w", err)
	}

	return &weaknesses, nil
}

func (h1 *Hackerone) send(method string, uri string, body io.Reader) ([]byte, string, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create send: %w", err)
	}

	req.Header = map[string][]string{
		"Accept": {"application/json"},
	}

	req.SetBasicAuth(h1.username, h1.token)
	resp, err := h1.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("api call failed: %s returned %s", uri, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	next, err := h1.nextPage(respBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get next page: %w", err)
	}
	return respBody, next, nil
}

func (h1 *Hackerone) nextPage(body []byte) (string, error) {
	type Resp struct {
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	resp := Resp{}

	err := json.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return resp.Links.Next, nil
}

func GetH1Token() string {
	if token, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".config/h1_token")); err != nil {
		return ""
	} else {
		return string(token)
	}
}
