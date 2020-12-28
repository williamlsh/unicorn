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

	subscriptions, err := subscribe(httpProxy, subURL)
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return batchBuild(ctx, subscriptions)
	})

	g.Go(func() error {
		rpt, err := makeReport(ctx, subscriptions)
		if err != nil {
			return err
		}

		sortOrigin(rpt.origin)
		sortByCountry(rpt.byCountry)
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("failed: %v\n", err)
	}
}
