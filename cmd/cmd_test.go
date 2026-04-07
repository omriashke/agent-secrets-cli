package cmd

import "testing"

func TestParseUserHost_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantUser string
		wantHost string
	}{
		{"deploy@myserver.com", "deploy", "myserver.com"},
		{"root@192.168.1.1", "root", "192.168.1.1"},
		{"user@host.internal", "user", "host.internal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			user, host, err := parseUserHost(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user != tt.wantUser {
				t.Errorf("user = %q, want %q", user, tt.wantUser)
			}
			if host != tt.wantHost {
				t.Errorf("host = %q, want %q", host, tt.wantHost)
			}
		})
	}
}

func TestParseUserHost_Invalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"nousersign"},
		{"@nouser"},
		{"nohost@"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, _, err := parseUserHost(tt.input)
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}
