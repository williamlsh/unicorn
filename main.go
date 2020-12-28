package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
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
		var reports []report
		reportsByCountry := make(map[string][]report)
		for _, s := range subscriptions {
			rpt, err := ping(ctx, s.Host, timeout*time.Second, count)
			if err != nil {
				return err
			}
			fmt.Printf("Report: %+v\n\n", rpt)

			// Filter hosts that is unable to ping.
			if rpt.avg > 0 {
				reports = append(reports, rpt)

				// Group reports by country.
				country := s.Fragment[:strings.Index(s.Fragment, "-")]
				if _, ok := reportsByCountry[country]; !ok {
					reportsByCountry[country] = make([]report, 0)
				}
				group := reportsByCountry[country]
				group = append(group, rpt)
				reportsByCountry[country] = group
			}
		}

		sortAll(reports)

		for k, g := range reportsByCountry {
			sort.SliceStable(g, func(i, j int) bool {
				return g[i].avg < g[j].avg
			})
			fmt.Printf("Report of %s:\n", k)
			for _, r := range g {
				fmt.Printf("%+v\n", r)
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("failed: %v\n", err)
	}
}
