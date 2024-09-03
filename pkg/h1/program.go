package h1

import (
	"encoding/json"
	"fmt"
	"github.com/ryanjarv/h1/pkg/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (h1 *Hackerone) Programs() ([]Program, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs")

	var programs []Program

	for uri != "" {
		var err error
		var resp []byte
		resp, uri, err = h1.send("GET", uri, nil)
		if err != nil {
			return nil, fmt.Errorf("Program.GetDetails: getting program: %w", err)
		}

		data := struct {
			Data []types.ProgramDetail
		}{}
		if err := json.Unmarshal(resp, &data); err != nil {
			return nil, fmt.Errorf("Program.GetDetails: failed to unmarshal program: %w", err)
		}

		for _, p := range data.Data {
			programs = append(programs, Program{
				Hackerone:         h1,
				Id:                p.Id,
				Type:              p.Type,
				Handle:            p.Attributes.Handle,
				ProgramAttributes: p.Attributes,
			})
		}
	}

	return programs, nil
}

func (h1 *Hackerone) Program(handle string) *Program {
	return &Program{
		Hackerone: h1,
		Handle:    handle,
	}
}

type Program struct {
	*Hackerone `json:"-"`

	// Handle is required for fetching program details.
	Handle string `json:"handle"`

	Id   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`

	types.ProgramAttributes
}

func (h1 *Program) GetDetail() (*types.ProgramDetail, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s", h1.Handle)

	resp, uri, err := h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Program.GetDetails: getting program: %w", err)
	} else if uri != "" {
		return nil, fmt.Errorf("failed to get all pages: %s", uri)
	}

	program := types.ProgramDetail{}
	err = json.Unmarshal(resp, &program)
	if err != nil {
		return nil, fmt.Errorf("Program.GetDetails: failed to unmarshal program: %w", err)
	}

	return &program, nil
}

func (h1 *Program) GetWeaknesses() (*types.Weaknesses, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s/weaknesses", h1.Handle)

	var resp []byte
	var err error

	resp, uri, err = h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", uri, err)
	} else if uri != "" {
		return nil, fmt.Errorf("failed to get all pages")
	}

	weaknesses := types.Weaknesses{}
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
	if token, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".config/h1_token")); err == nil {
		return string(token)
	}

	if token := os.Getenv("H1_TOKEN"); token != "" {
		return string(token)
	}
	return ""
}
