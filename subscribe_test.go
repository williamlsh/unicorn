package main

import (
	"testing"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	if _, err := fetch(server.Client(), server.URL); err != nil {
		t.Fatal(err)
	}
}

func TestParseSubscriptions(t *testing.T) {
	t.Parallel()

	resp, err := fetch(server.Client(), server.URL)
	if err != nil {
		t.Skip(err)
	}
	defer resp.Body.Close()

	if _, err := parseSubscriptions(resp); err != nil {
		t.Error(err)
	}
}
