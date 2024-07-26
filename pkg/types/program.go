package types

import "time"

type Weaknesses struct {
	Data []struct {
		Id         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name        string    `json:"name"`
			Description string    `json:"description"`
			CreatedAt   time.Time `json:"created_at"`
			ExternalId  string    `json:"external_id"`
		} `json:"attributes"`
	} `json:"data"`
	Links struct {
	} `json:"links"`
}
