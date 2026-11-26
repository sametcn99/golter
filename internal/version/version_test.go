package version

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want bool
	}{
		{"0.1.0", "0.1.1", true},
		{"0.1.1", "0.1.0", false},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "1.0.0", true},
		{"1.0.0", "0.9.9", false},
		{"0.1", "0.1.1", true},
		{"0.1.1", "0.1", false},
		{"0.1.0", "0.1.0.1", true},
		{"2026.03.31.235959", "2026.04.01.000000", true},
		{"2026.04.01.120000", "2026.04.01.130000", true},
		{"2026.04.01.130000", "2026.04.01.120000", false},
	}

	for _, tt := range tests {
		if got := isNewer(tt.v1, tt.v2); got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestReleaseCheck(t *testing.T) {
	for _, tc := range []struct {
		name, body, current string
		status              int
		available, fail     bool
	}{
		{"newer", `{"tag_name":"v2.1","html_url":"https://example.org/release"}`, "2.0", 200, true, false},
		{"current", `{"tag_name":"v2.1"}`, "v2.1", 200, false, false},
		{"offline status", "unavailable", "1", 503, false, true},
		{"invalid json", "{", "1", 200, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); io.WriteString(w, tc.body) }))
			defer server.Close()
			latest, url, available, err := checkForUpdates(server.Client(), server.URL, tc.current)
			if (err != nil) != tc.fail || available != tc.available {
				t.Fatalf("%q %q %t %v", latest, url, available, err)
			}
			if available && (latest != "2.1" || url != "https://example.org/release") {
				t.Fatal(latest, url)
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
		defer server.Close()
		client := server.Client()
		client.Timeout = 50 * time.Millisecond
		_, _, _, err := checkForUpdates(client, server.URL, "1")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})
}
