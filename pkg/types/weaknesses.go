package types

import "time"

type ProgramDetail struct {
	Id            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ProgramAttributes `json:"attributes"`
	Relationships Relationships     `json:"relationships"`
}

type ProgramAttributes struct {
	Handle                          string    `json:"handle,omitempty"`
	Name                            string    `json:"name,omitempty"`
	Currency                        string    `json:"currency,omitempty"`
	ProfilePicture                  string    `json:"profile_picture,omitempty"`
	SubmissionState                 string    `json:"submission_state,omitempty"`
	TriageActive                    bool      `json:"triage_active,omitempty"`
	State                           string    `json:"state,omitempty"`
	StartedAcceptingAt              time.Time `json:"started_accepting_at,omitempty"`
	NumberOfReportsForUser          int       `json:"number_of_reports_for_user,omitempty"`
	NumberOfValidReportsForUser     int       `json:"number_of_valid_reports_for_user,omitempty"`
	BountyEarnedForUser             float64   `json:"bounty_earned_for_user,omitempty"`
	LastInvitationAcceptedAtForUser time.Time `json:"last_invitation_accepted_at_for_user,omitempty"`
	Bookmarked                      bool      `json:"bookmarked,omitempty"`
	AllowsBountySplitting           bool      `json:"allows_bounty_splitting,omitempty"`
	OffersBounties                  bool      `json:"offers_bounties,omitempty"`
}

type Relationships struct {
	StructuredScopes StructuredScopes `json:"structured_scopes"`
}

type StructuredScopes struct {
	Data []ScopeData `json:"data"`
}

type ScopeData struct {
	Id         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes ScopeAttributes `json:"attributes"`
}

type ScopeAttributes struct {
	AssetType                  string    `json:"asset_type"`
	AssetIdentifier            string    `json:"asset_identifier"`
	EligibleForBounty          bool      `json:"eligible_for_bounty"`
	EligibleForSubmission      bool      `json:"eligible_for_submission"`
	Instruction                *string   `json:"instruction"`
	MaxSeverity                string    `json:"max_severity"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	ConfidentialityRequirement string    `json:"confidentiality_requirement,omitempty"`
	IntegrityRequirement       string    `json:"integrity_requirement,omitempty"`
	AvailabilityRequirement    string    `json:"availability_requirement,omitempty"`
}
