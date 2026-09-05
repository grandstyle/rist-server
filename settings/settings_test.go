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
