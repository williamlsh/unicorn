package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	subURL, httpProxy, configDir string
	timeout                      time.Duration
	count                        int
)

func init() {
	flag.StringVar(&subURL, "url", "", "Subscription URL")
	flag.StringVar(&httpProxy, "http-proxy", "", "HTTP proxy to retrieve subscription")
	flag.StringVar(&configDir, "dir", "tmp", "Top level of configuration directory")
	flag.DurationVar(&timeout, "timeout", 10, "Single ping timeout in s")
	flag.IntVar(&count, "count", 3, "Ping counts for every server")
}

func main() {
	flag.Parse()

	if subURL == "" {
		log.Fatal("Input subscriptin URL is empty")
	}

	client, err := newClient(httpProxy)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := fetch(client, subURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	subscriptions, err := parseSubscriptions(resp)
	if err != nil {
		fmt.Println(err)
		// Return to honor any deferred call.
		return
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		// Folders names may be duplicate, use a map to avoid creating the same folder.
		folders := make(map[string]struct{})
		for _, s := range subscriptions {
			if err := buildConfig(ctx, s, configDir, folders); err != nil {
				return err
			}
		}
		return nil
	})
	g.Go(func() error {
		for _, s := range subscriptions {
			report, err := ping(ctx, s.Host, timeout*time.Second, count)
			if err != nil {
				return err
			}
			fmt.Printf("Report: %#v\n\n", report)
		}
		return nil
	})

	fmt.Println("Done", g.Wait())
}
