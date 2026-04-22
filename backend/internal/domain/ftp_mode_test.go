package domain

import "testing"

func TestFtpMode(t *testing.T) {
	tests := []struct {
		name string
		cfg  map[string]any
		want string
	}{
		{"missing", nil, ""},
		{"explicit", map[string]any{"ftpMode": "ftps-explicit"}, "ftps-explicit"},
		{"implicit", map[string]any{"ftpMode": "ftps-implicit"}, "ftps-implicit"},
		{"plain", map[string]any{"ftpMode": "ftp"}, "ftp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Setting{Platform: "sftp"}
			if tt.cfg != nil {
				s.PlatformConfigs = map[string]map[string]any{"sftp": tt.cfg}
			}
			if got := s.FtpMode(); got != tt.want {
				t.Errorf("FtpMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllowInsecureTLS(t *testing.T) {
	tests := []struct {
		name string
		cfg  map[string]any
		want bool
	}{
		{"missing", nil, false},
		{"bool_true", map[string]any{"allowInsecureTLS": true}, true},
		{"bool_false", map[string]any{"allowInsecureTLS": false}, false},
		{"string_true", map[string]any{"allowInsecureTLS": "true"}, true},
		{"string_1", map[string]any{"allowInsecureTLS": "1"}, true},
		{"string_false", map[string]any{"allowInsecureTLS": "false"}, false},
		{"empty", map[string]any{"allowInsecureTLS": ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Setting{Platform: "sftp"}
			if tt.cfg != nil {
				s.PlatformConfigs = map[string]map[string]any{"sftp": tt.cfg}
			}
			if got := s.AllowInsecureTLS(); got != tt.want {
				t.Errorf("AllowInsecureTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}
