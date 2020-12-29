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

	if count < 1 {
		ch <- stats{err: errors.New("ping count must be larger than 0")}
		return
	}

	addr := trueURL.Host
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

type primaryData struct {
	basics
	groups
}

// groups is grouped stats by countries
type groups map[string]basics

// basics is non-classified stats.
type basics []stats

func (b basics) String() string {
	var a string
	for _, s := range b {
		a += fmt.Sprintln(s)
	}
	return a
}

func (d primaryData) sortAll() basics {
	sort.SliceStable(d.basics, func(i, j int) bool {
		return d.basics[i].avg < d.basics[j].avg
	})

	// Print the ranks.
	fmt.Printf("\nBasic ranks:\n%s", d.basics)
	return d.basics
}

func (d primaryData) sortByCountry() groups {
	for k, g := range d.groups {
		sort.SliceStable(g, func(i, j int) bool {
			return g[i].avg < g[j].avg
		})
		d.groups[k] = g

		// Print the ranks.
		fmt.Printf("\nGrouped ranks of %s:\n%s", k, g)
	}
	return d.groups
}

func pingAll(ctx context.Context, subscriptions []*url.URL) (primaryData, error) {
	var basicStats basics
	groupedStats := make(groups)
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

	for result := range ch {
		if result.err != nil {
			return primaryData{}, result.err
		}

		// Filter hosts that is unable to ping.
		if result.avg > 0 {
			basicStats = append(basicStats, result)

			// Group stats by country.
			country := result.url.Fragment[:strings.Index(result.url.Fragment, "-")]
			if _, ok := groupedStats[country]; !ok {
				groupedStats[country] = make(basics, 0)
			}
			group := groupedStats[country]
			group = append(group, result)
			groupedStats[country] = group
		}
	}

	return primaryData{basicStats, groupedStats}, nil
}
