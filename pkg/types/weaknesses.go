package types

import "time"

type ProgramDetail struct {
	Id            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ProgramAttributes `json:"attributes"`
	Relationships Relationships     `json:"relationships"`
}

type ProgramAttributes struct {
	Handle                          string    `json:"handle"`
	Name                            string    `json:"name"`
	Currency                        string    `json:"currency"`
	ProfilePicture                  string    `json:"profile_picture"`
	SubmissionState                 string    `json:"submission_state"`
	TriageActive                    bool      `json:"triage_active"`
	State                           string    `json:"state"`
	StartedAcceptingAt              time.Time `json:"started_accepting_at"`
	NumberOfReportsForUser          int       `json:"number_of_reports_for_user"`
	NumberOfValidReportsForUser     int       `json:"number_of_valid_reports_for_user"`
	BountyEarnedForUser             float64   `json:"bounty_earned_for_user"`
	LastInvitationAcceptedAtForUser time.Time `json:"last_invitation_accepted_at_for_user"`
	Bookmarked                      bool      `json:"bookmarked"`
	AllowsBountySplitting           bool      `json:"allows_bounty_splitting"`
	OffersBounties                  bool      `json:"offers_bounties"`
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
