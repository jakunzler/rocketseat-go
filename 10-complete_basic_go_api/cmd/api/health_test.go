package main

import "testing"

func TestHealthURL(t *testing.T) {
	tests := map[string]string{
		":8080":          "http://127.0.0.1:8080/healthz",
		"0.0.0.0:8080":   "http://127.0.0.1:8080/healthz",
		"127.0.0.1:9090": "http://127.0.0.1:9090/healthz",
	}
	for addr, want := range tests {
		if got := healthURL(addr); got != want {
			t.Fatalf("healthURL(%q) = %q, want %q", addr, got, want)
		}
	}
}
