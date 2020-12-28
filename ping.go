package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type stats struct {
	err                   error
	url                   *url.URL
	min, avg, max, stddev float64 // calculation results
}

func (s stats) String() string {
	return fmt.Sprintf("Ping stats of %s min: %.2fms max: %.2fms avg: %.2fms stddev: %.2fms", s.url.Host, s.min, s.max, s.avg, s.stddev)
}

func ping(ctx context.Context, trueURL *url.URL, timeout time.Duration, count int, wg *sync.WaitGroup, ch chan<- stats) {
	defer wg.Done()
	addr := trueURL.Host

	if count < 1 {
		ch <- stats{err: errors.New("ping count must be larger than 0")}
		return
	}

	results := make([]float64, count)
	for i := 0; i < count; i++ {
		startAt := time.Now()

		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			// Skip timed out ping.
			results = append(results, 0)
			fmt.Printf("Ping timeout: seq=%d addr=%s err=%v\n", i+1, addr, err)
			continue
		}
		_ = conn.Close()

		endAt := time.Now()
		roundtrip := float64(endAt.UnixNano()-startAt.UnixNano()) / 1e6
		results[i] = roundtrip

		fmt.Printf("Ping %s seq=%d time=%.3fms\n", addr, i+1, roundtrip)
	}

	rpt := calculate(results)
	rpt.url = trueURL
	ch <- rpt
}

func calculate(results []float64) stats {
	max := results[0]
	min := results[0]
	sum := 0.0
	stddev := 0.0
	for _, v := range results {
		if max < v {
			max = v
		}
		if min > v {
			min = v
		}
		sum += v
	}
	avg := sum / float64(count)

	for _, v := range results {
		stddev += math.Pow(v-avg, 2)
	}
	stddev = math.Sqrt(stddev / float64(count))

	return stats{
		nil,
		nil,
		min,
		avg,
		max,
		stddev,
	}
}

type report struct {
	origin    []stats            // Non-classified stats
	byCountry map[string][]stats // Grouped stats by countries
}

func makeReport(ctx context.Context, subscriptions []*url.URL) (report, error) {
	var origReport []stats
	reportByCountry := make(map[string][]stats)

	ch := make(chan stats)

	var wg sync.WaitGroup
	for _, s := range subscriptions {
		wg.Add(1)
		go ping(ctx, s, timeout*time.Second, count, &wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for rpt := range ch {
		if rpt.err != nil {
			return report{}, rpt.err
		}

		// Filter hosts that is unable to ping.
		if rpt.avg > 0 {
			origReport = append(origReport, rpt)

			// Group reports by country.
			country := rpt.url.Fragment[:strings.Index(rpt.url.Fragment, "-")]
			if _, ok := reportByCountry[country]; !ok {
				reportByCountry[country] = make([]stats, 0)
			}
			group := reportByCountry[country]
			group = append(group, rpt)
			reportByCountry[country] = group
		}
	}

	return report{origReport, reportByCountry}, nil
}

func sortOrigin(data []stats) {
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].avg < data[j].avg
	})

	fmt.Println("\nOriginal report:")
	for _, r := range data {
		fmt.Println(r)
	}
}

func sortByCountry(data map[string][]stats) {
	for k, g := range data {
		sort.SliceStable(g, func(i, j int) bool {
			return g[i].avg < g[j].avg
		})

		fmt.Printf("\nReport of %s:\n", k)
		for _, r := range g {
			fmt.Println(r)
		}
	}
}
