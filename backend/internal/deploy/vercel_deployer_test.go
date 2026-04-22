package deploy

import (
	"strings"
	"testing"
)

func TestParseVercelDeployBody(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantMissing []string
		wantID      string
		wantErrSub  string // 不为空则要求 error 包含该子串
	}{
		{
			name:        "2xx_all_good",
			status:      201,
			body:        `{"id":"dep_xyz"}`,
			wantID:      "dep_xyz",
			wantMissing: nil,
		},
		{
			name:        "2xx_with_missing_top_level",
			status:      200,
			body:        `{"id":"dep_xyz","missing":["sha1","sha2"]}`,
			wantID:      "dep_xyz",
			wantMissing: []string{"sha1", "sha2"},
		},
		{
			name:        "non_2xx_error_missing",
			status:      400,
			body:        `{"error":{"code":"missing_files","missing":["sha1","sha2","sha3"]}}`,
			wantMissing: []string{"sha1", "sha2", "sha3"},
		},
		{
			name:       "non_2xx_real_error_with_message",
			status:     401,
			body:       `{"error":{"code":"forbidden","message":"invalid token"}}`,
			wantErrSub: "invalid token",
		},
		{
			name:       "non_2xx_plain_text",
			status:     500,
			body:       "Internal Server Error",
			wantErrSub: "Internal Server Error",
		},
		{
			name:        "2xx_empty_body",
			status:      200,
			body:        ``,
			wantMissing: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVercelDeployBody(tt.status, []byte(tt.body))
			if tt.wantErrSub != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSub)
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Errorf("expected error to contain %q, got %v", tt.wantErrSub, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil resp")
			}
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if len(got.Missing) != len(tt.wantMissing) {
				t.Errorf("Missing = %v, want %v", got.Missing, tt.wantMissing)
			}
		})
	}
}
