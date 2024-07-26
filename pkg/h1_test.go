package lib

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
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
	t.SkipNow()

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
							Body: io.NopCloser(bytes.NewReader([]byte(`{}`))),
						}},
					},
				},
			},
			wantErr: false,
			want:    &types.ProgramDetail{},
		},
		{
			name: "pagination works",
			fields: fields{
				Hackerone: Hackerone{
					token:    "token",
					username: "username",
					client: &MockClient{
						DoResponse: []*http.Response{
							{Body: io.NopCloser(bytes.NewReader([]byte(`{"links": { "next": "test" }}`)))},
							{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))},
						},
					},
				},
			},
			wantErr:           false,
			wantTimesDoCalled: 2,
			want:              &types.ProgramDetail{},
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
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("GetDetail() mismatch (-want +got):\n%s", diff)
			}

			called := len(h1.client.(*MockClient).Calls)
			if tt.wantTimesDoCalled != 0 && called != tt.wantTimesDoCalled {
				t.Errorf("GetDetail() called %d times, want %d", called, tt.wantTimesDoCalled)
			}
		})
	}
}
