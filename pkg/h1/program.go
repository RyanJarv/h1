package h1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ryanjarv/h1/pkg/types"
)

const MaxRetries = 3

func (h1 *Hackerone) Programs(yield func(*Program) bool) {
	h1.ProgramsWithErrs(func(p *Program, err error) bool {
		if err != nil {
			log.Printf("error getting programs: %s", err)
			return true
		}
		return yield(p)
	})
}

func (h1 *Hackerone) ProgramsWithErrs(yield func(*Program, error) bool) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs")

	for uri != "" {
		var err error
		var resp []byte
		resp, uri, err = h1.send("GET", uri, nil)
		if err != nil {
			if !yield(nil, fmt.Errorf("programs: getting programs: %w", err)) {
				return
			}
		}

		data := struct {
			Data []types.ProgramDetail
		}{}
		if err := json.Unmarshal(resp, &data); err != nil {
			if !yield(nil, fmt.Errorf("programs: failed to unmarshal programs: %w", err)) {
				return
			}
		}

		for _, p := range data.Data {
			if !yield(&Program{
				Hackerone:         h1,
				Id:                p.Id,
				Type:              p.Type,
				Handle:            p.Attributes.Handle,
				ProgramAttributes: p.Attributes,
			}, nil) {
				return
			}
		}
	}
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

func (h1 *Program) GetId() string { return h1.Handle }
func (h1 *Program) GetDetail() (*types.ProgramDetail, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s", h1.Handle)

	resp, uri, err := h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("GetDetail: getting program: %w", err)
	} else if uri != "" {
		return nil, fmt.Errorf("GetDetail: unexpected pagination for single program: %s", uri)
	}

	program := types.ProgramDetail{}
	err = json.Unmarshal(resp, &program)
	if err != nil {
		return nil, fmt.Errorf("GetDetail: failed to unmarshal program: %w", err)
	}

	return &program, nil
}

func (h1 *Program) GetWeaknesses() (*types.Weaknesses, error) {
	uri := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s/weaknesses", h1.Handle)

	var resp []byte
	var err error

	resp, uri, err = h1.send("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("GetWeaknesses: getting weaknesses: %w", err)
	} else if uri != "" {
		return nil, fmt.Errorf("GetWeaknesses: unexpected pagination for weaknesses: %s", uri)
	}

	weaknesses := types.Weaknesses{}
	err = json.Unmarshal(resp, &weaknesses)
	if err != nil {
		return nil, fmt.Errorf("GetWeaknesses: failed to unmarshal weaknesses: %w", err)
	}

	return &weaknesses, nil
}

func (h1 *Hackerone) send(method string, uri string, body io.Reader) ([]byte, string, error) {
	var all []byte
	var err error
	if body != nil {
		if all, err = io.ReadAll(body); err != nil {
			return nil, "", fmt.Errorf("send: failed to read request body: %w", err)
		}
	}

	for retries := 0; retries < MaxRetries; retries++ {
		respBody, next, err := h1.sendOnce(method, uri, bytes.NewReader(all))
		// Check for specific error types
		var opErr *net.OpError
		var sysErr syscall.Errno
		if errors.As(err, &opErr) {
			if errors.As(opErr.Err, &sysErr) && errors.Is(sysErr, syscall.ECONNRESET) {
				// Handle the connection reset specifically with exponential backoff
				backoff := time.Duration(retries+1) * 100 * time.Millisecond
				log.Printf("connection reset (attempt %d/%d): %s, retrying in %v", retries+1, MaxRetries, err, backoff)
				time.Sleep(backoff)
				continue
			}
		}
		return respBody, next, err
	}

	return nil, "", fmt.Errorf("failed to send request after %d retries", MaxRetries)
}

func (h1 *Hackerone) sendOnce(method string, uri string, body io.Reader) ([]byte, string, error) {
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
	path := filepath.Join(os.Getenv("HOME"), ".config/h1_token")

	if tokenBytes, err := os.ReadFile(path); err == nil {
		log.Printf("Using H1 token from %s", path)
		return string(tokenBytes)
	}

	if token := os.Getenv("H1_TOKEN"); token != "" {
		log.Printf("Using H1 token from H1_TOKEN environment variable")
		return token
	}

	log.Printf("No H1 token found in %s or H1_TOKEN environment variable", path)
	return ""
}
