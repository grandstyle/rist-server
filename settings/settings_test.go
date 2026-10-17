package settings

import (
	"testing")

func TestParseOutTarget(t *testing.T) {
	tests := []struct {
		input       string
		defaultPort int
		wantHost    string
		wantPort    int
		wantErr     bool
	}{
	
{"", 10000, "127.0.0.1", 10000, false},
	
{"   ", 10000, "127.0.0.1", 10000, false},
	
{"10000", 10000, "127.0.0.1", 10000, false},
	{"5000", 10000, "127.0.0.1", 5000, false},
	{"192.168.5.39:10000", 10000, "192.168.5.39", 10000, false},
	{"192.168.1.100:8000", 10000, "192.168.1.100", 8000, false},
	{"127.0.0.1:10000", 10000, "127.0.0.1", 10000, false},
	{":10000", 10000, "127.0.0.1", 10000, false},
	{"192.168.5.39", 10000, "192.168.5.39", 10000, false},
	{"70000", 10000, "", 0, true},
	{"-5", 10000, "", 0, true},
	{"192.168.1.50:99999", 10000, "", 0, true},
	{"192.168.1.50:abc", 10000, "", 0, true},
	}

	for _, tc := range tests {
		gotHost, gotPort, err := ParseOutTarget(tc.input, tc.defaultPort)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseOutTarget(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseOutTarget(%q) unexpected error: %v", tc.input, err)
			}
			if gotHost != tc.wantHost {
				t.Errorf("ParseOutTarget(%q) gotHost = %q, want %q", tc.input, gotHost, tc.wantHost)
			}
			if gotPort != tc.wantPort {
				t.Errorf("ParseOutTarget(%q) gotPort = %d, want %d", tc.input, gotPort, tc.wantPort)
			}
		}
	}
}

func TestParseRistPort(t *testing.T) {
	tests := []struct {
		input       string
		defaultPort int
		wantPort    int
		wantErr     bool
	}{
		{"", 1968, 1968, false},
		{"   ", 1968, 1968, false},
		{"1968", 1968, 1968, false},
		{"2000", 1968, 2000, false},
		{"192.168.1.50:1968", 1968, 1968, false},
		{"192.168.1.50:7890", 1968, 7890, false},
		{":1968", 1968, 1968, false},
		{":7890", 1968, 7890, false},
		{"0.0.0.0:1968", 1968, 1968, false},
		{"70000", 1968, 0, true},
		{"-1", 1968, 0, true},
		{"192.168.1.50:99999", 1968, 0, true},
		{"invalid", 1968, 1968, false}, // falls nur text/ip ohne port -> defaultPort
	}

	for _, tc := range tests {
		gotPort, err := ParseRistPort(tc.input, tc.defaultPort)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseRistPort(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseRistPort(%q) unexpected error: %v", tc.input, err)
			}
			if gotPort != tc.wantPort {
				t.Errorf("ParseRistPort(%q) gotPort = %d, want %d", tc.input, gotPort, tc.wantPort)
			}
		}
	}
}

