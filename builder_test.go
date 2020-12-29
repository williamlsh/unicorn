package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildConfig(t *testing.T) {
	t.Parallel()

	resp, err := fetch(server.Client(), server.URL)
	if err != nil {
		t.Skip(err)
	}
	defer resp.Body.Close()

	subscriptions, err := parseSubscriptions(resp)
	if err != nil {
		t.Skip(err)
	}
	folders := make(map[string]struct{})
	tempDir := t.TempDir()
	for _, s := range subscriptions {
		if err := buildConfig(context.TODO(), s, tempDir, folders, false); err != nil {
			t.Error(err)
		}
	}

	if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("path: %s\n", path)
		return nil
	}); err != nil {
		t.Error(err)
	}

	t.Run("Not force build configurations", func(t *testing.T) {
		t.Parallel()
		for _, s := range subscriptions {
			if err := buildConfig(context.TODO(), s, tempDir, folders, false); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Force build configurations", func(t *testing.T) {
		t.Parallel()
		for _, s := range subscriptions {
			if err := buildConfig(context.TODO(), s, tempDir, folders, true); err != nil {
				t.Error(err)
			}
		}
	})
}
