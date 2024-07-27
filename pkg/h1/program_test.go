package h1

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ryanjarv/h1/pkg/types"
	"io"
	"net/http"
	"testing"
)

type MockClient struct {
	DoResponse []*http.Response
	Calls      []http.Request
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.Calls = append(m.Calls, *req)
	resp := m.DoResponse[0]
	m.DoResponse = m.DoResponse[1:]
	return resp, nil
}

func TestProgram_GetDetail(t *testing.T) {
	type fields struct {
		Hackerone Hackerone
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
				Hackerone: Hackerone{
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
								Id:   "131858",
								Type: "structured-scope",
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
				ProgramId: tt.fields.ProgramId,
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
		Hackerone Hackerone
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
				Hackerone: Hackerone{
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
				{ProgramId: "13"},
			},
		},
		{
			name: "pagination works",
			fields: fields{
				Hackerone: Hackerone{
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
				{ProgramId: "1"},
				{ProgramId: "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := &Program{
				Hackerone: tt.fields.Hackerone,
				ProgramId: tt.fields.ProgramId,
			}
			got, err := h1.Programs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Programs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(Hackerone{})); diff != "" {
				t.Errorf("Programs() mismatch (-want +got):\n%s", diff)
			}

			called := len(h1.client.(*MockClient).Calls)
			if tt.wantTimesDoCalled != 0 && called != tt.wantTimesDoCalled {
				t.Errorf("Programs() called %d times, want %d", called, tt.wantTimesDoCalled)
			}
		})
	}
}
