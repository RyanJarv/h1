package h1

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ryanjarv/h1/pkg/types"
)

type MockClient struct {
	DoResponse []*http.Response
	DoErrors   []error
	Calls      []http.Request
	CallCount  int
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.Calls = append(m.Calls, *req)
	m.CallCount++

	// Return error if one is defined for this call
	if len(m.DoErrors) > 0 {
		err := m.DoErrors[0]
		m.DoErrors = m.DoErrors[1:]
		if err != nil {
			return nil, err
		}
	}

	// Return response if one is defined
	if len(m.DoResponse) > 0 {
		resp := m.DoResponse[0]
		m.DoResponse = m.DoResponse[1:]
		return resp, nil
	}

	return nil, nil
}

func TestProgram_GetDetail(t *testing.T) {
	type fields struct {
		Hackerone *Hackerone
		ProgramId string
		client    Client
	}
	tests := []struct {
		name              string
		fields            fields
		want              *types.ProgramDetail
		wantErr           bool
		wantTimesDoCalled int
	}{
		{
			name: "TestGetDetail",
			fields: fields{
				Hackerone: &Hackerone{
					token:    "token",
					username: "username",
					client: &MockClient{
						DoResponse: []*http.Response{{
							StatusCode: 200,
							Body: io.NopCloser(bytes.NewReader([]byte(`
{
  "id": "13",
  "type": "program",
  "attributes": {
    "handle": "security",
    "name": "HackerOne"
  },
  "relationships": {
    "structured_scopes": {
      "data": [
        {
          "id": "131858",
          "type": "structured-scope",
          "attributes": {
            "asset_type": "URL",
            "asset_identifier": "hackathon-photos-us-east-2.hackerone-user-content.com"
          }
        }
      ]
    }
  }
}`))),
						}},
					},
				},
			},
			wantErr: false,
			want: &types.ProgramDetail{
				Id:   "13",
				Type: "program",
				Attributes: types.ProgramAttributes{
					Handle: "security",
					Name:   "HackerOne",
				},
				Relationships: types.Relationships{
					StructuredScopes: types.StructuredScopes{
						Data: []types.ScopeData{
							{
								Id: "131858",
								Attributes: types.ScopeAttributes{
									AssetType:       "URL",
									AssetIdentifier: "hackathon-photos-us-east-2.hackerone-user-content.com",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := &Program{
				Hackerone: tt.fields.Hackerone,
				Id:        tt.fields.ProgramId,
			}
			got, err := h1.GetDetail()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetDetail() mismatch (-want +got):\n%s", diff)
			}

			called := len(h1.client.(*MockClient).Calls)
			if tt.wantTimesDoCalled != 0 && called != tt.wantTimesDoCalled {
				t.Errorf("GetDetail() called %d times, want %d", called, tt.wantTimesDoCalled)
			}
		})
	}
}

func TestProgram_Programs(t *testing.T) {
	type fields struct {
		Hackerone *Hackerone
		ProgramId string
		client    Client
	}
	tests := []struct {
		name              string
		fields            fields
		want              []Program
		wantErr           bool
		wantTimesDoCalled int
	}{
		{
			name: "TestListPrograms",
			fields: fields{
				Hackerone: &Hackerone{
					token:    "token",
					username: "username",
					client: &MockClient{
						DoResponse: []*http.Response{{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{ "data": [{ "id": "13" } ] }`))),
						}},
					},
				},
			},
			wantErr: false,
			want: []Program{
				{Id: "13"},
			},
		},
		{
			name: "pagination works",
			fields: fields{
				Hackerone: &Hackerone{
					token:    "token",
					username: "username",
					client: &MockClient{
						DoResponse: []*http.Response{
							{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(`{"data": [{"id": "1"}], "links": { "next": "test" }}`)))},
							{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(`{"data": [{"id": "2"}]}`)))},
						},
					},
				},
			},
			wantErr:           false,
			wantTimesDoCalled: 2,
			want: []Program{
				{Id: "1"},
				{Id: "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := tt.fields.Hackerone
			var got []Program
			var err error
			h1.Programs(func(p *Program, e error) bool {
				if e != nil {
					err = e
					return false
				}
				got = append(got, *p)
				return true
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Programs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(Program{}, "Hackerone")); diff != "" {
				t.Errorf("Programs() mismatch (-want +got):\n%s", diff)
			}

			called := len(h1.client.(*MockClient).Calls)
			if tt.wantTimesDoCalled != 0 && called != tt.wantTimesDoCalled {
				t.Errorf("Programs() called %d times, want %d", called, tt.wantTimesDoCalled)
			}
		})
	}
}

func TestHackerone_Retries(t *testing.T) {
	tests := []struct {
		name              string
		mockErrors        []error
		mockResponses     []*http.Response
		wantErr           bool
		wantCallCount     int
		wantErrContains   string
	}{
		{
			name: "success on first try",
			mockErrors: []error{nil},
			mockResponses: []*http.Response{{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"data": [{"id": "1"}]}`))),
			}},
			wantErr:       false,
			wantCallCount: 1,
		},
		{
			name: "success after one connection reset",
			mockErrors: []error{
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
				nil,
			},
			mockResponses: []*http.Response{
				{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"data": [{"id": "1"}]}`))),
				},
			},
			wantErr:       false,
			wantCallCount: 2,
		},
		{
			name: "success after two connection resets",
			mockErrors: []error{
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
				nil,
			},
			mockResponses: []*http.Response{
				{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"data": [{"id": "1"}]}`))),
				},
			},
			wantErr:       false,
			wantCallCount: 3,
		},
		{
			name: "fail after max retries with connection reset",
			mockErrors: []error{
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ECONNRESET)},
			},
			mockResponses:   []*http.Response{},
			wantErr:         true,
			wantCallCount:   3,
			wantErrContains: "failed to send request after 3 retries",
		},
		{
			name: "non-retryable error returns immediately",
			mockErrors: []error{
				&net.OpError{Op: "read", Net: "tcp", Err: syscall.Errno(syscall.ETIMEDOUT)},
			},
			mockResponses: []*http.Response{},
			wantErr:       true,
			wantCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				DoErrors:   tt.mockErrors,
				DoResponse: tt.mockResponses,
			}

			h1 := &Hackerone{
				token:    "test-token",
				username: "test-user",
				client:   mockClient,
			}

			var programs []Program
			var gotErr error
			h1.Programs(func(p *Program, err error) bool {
				if err != nil {
					gotErr = err
					return false
				}
				programs = append(programs, *p)
				return true
			})

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Programs() error = %v, wantErr %v", gotErr, tt.wantErr)
			}

			if tt.wantErrContains != "" && gotErr != nil {
				if !bytes.Contains([]byte(gotErr.Error()), []byte(tt.wantErrContains)) {
					t.Errorf("Programs() error = %v, want error containing %q", gotErr, tt.wantErrContains)
				}
			}

			if mockClient.CallCount != tt.wantCallCount {
				t.Errorf("Do() called %d times, want %d", mockClient.CallCount, tt.wantCallCount)
			}
		})
	}
}

func TestProgram__functional(t *testing.T) {
	user := os.Getenv("H1_USERNAME")
	if user == "" {
		t.Skip("no H1_USERNAME set")
	}

	h1 := NewHackerone(&NewHackeroneInput{Username: user})
	h1.Programs(func(p *Program, err error) bool {
		if err != nil {
			t.Fatalf("getting programs: %s", err)
		}

		// Get the program details
		_, err = p.GetDetail()
		if err != nil {
			t.Fatalf("getting detail: %s: %s", p.Id, err)
		}
		return true
	})
}
