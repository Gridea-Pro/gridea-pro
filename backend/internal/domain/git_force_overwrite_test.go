package domain

import "testing"

func TestGitForceOverwrite(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		cfg      map[string]any
		want     bool
	}{
		{"missing_platform", "github", nil, false},
		{"empty_cfg", "github", map[string]any{}, false},
		{"bool_true", "github", map[string]any{"gitForceOverwrite": true}, true},
		{"bool_false", "github", map[string]any{"gitForceOverwrite": false}, false},
		{"string_true", "github", map[string]any{"gitForceOverwrite": "true"}, true},
		{"string_1", "github", map[string]any{"gitForceOverwrite": "1"}, true},
		{"string_false", "github", map[string]any{"gitForceOverwrite": "false"}, false},
		{"unrelated_type", "github", map[string]any{"gitForceOverwrite": 42}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Setting{Platform: tt.platform}
			if tt.cfg != nil {
				s.PlatformConfigs = map[string]map[string]any{tt.platform: tt.cfg}
			}
			if got := s.GitForceOverwrite(); got != tt.want {
				t.Errorf("GitForceOverwrite() = %v, want %v", got, tt.want)
			}
		})
	}
}
