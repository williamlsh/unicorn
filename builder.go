package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// defaultConfig is a default configuration.
// Only remote_addr and password fields will be replaces with subscription values.
var defaultConfig = `
{
	"run_type": "client",
	"local_addr": "127.0.0.1",
	"local_port": 1080,
	"remote_addr": "example.com",
	"remote_port": 443,
	"password": ["password1"],
	"log_level": 1,
	"ssl": {
	  "verify": true,
	  "verify_hostname": true,
	  "cert": "",
	  "cipher": "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES128-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA:AES128-SHA:AES256-SHA:DES-CBC3-SHA",
	  "cipher_tls13": "TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384",
	  "sni": "",
	  "alpn": ["h2", "http/1.1"],
	  "reuse_session": true,
	  "session_ticket": false,
	  "curves": ""
	},
	"tcp": {
	  "no_delay": true,
	  "keep_alive": true,
	  "reuse_port": false,
	  "fast_open": false,
	  "fast_open_qlen": 20
	}
}`

// buildConfig builds configuration from subscription.
func buildConfig(ctx context.Context, trueURL *url.URL, configDir string, folders map[string]struct{}) error {
	dashIndex := strings.Index(trueURL.Fragment, "-")

	folder := trueURL.Fragment[:dashIndex]
	dirpath := path.Join(configDir, folder)
	if _, ok := folders[folder]; !ok {
		folders[folder] = struct{}{}

		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return fmt.Errorf("could not create dir: %w", err)
		}
	}

	filepath := path.Join(dirpath, trueURL.Fragment[dashIndex+1:]+".json")

	// Skip updating configuration file modified within a month.
	if info, err := os.Stat(filepath); !os.IsNotExist(err) && info.ModTime().Before(time.Now().AddDate(0, 1, 0)) {
		fmt.Println("skipping updating configuration file.")
		return nil
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}

	if _, err := strings.NewReplacer("example.com", trueURL.Hostname(), "password1", trueURL.User.Username()).WriteString(file, defaultConfig); err != nil {
		return fmt.Errorf("could not replace strings: %w", err)
	}

	return nil
}

func batchBuild(ctx context.Context, subscriptions []*url.URL) error {
	// Folders names may be duplicate, use a map to avoid creating the same folder again.
	folders := make(map[string]struct{})
	for _, s := range subscriptions {
		if err := buildConfig(ctx, s, configDir, folders); err != nil {
			return err
		}
	}
	return nil
}
