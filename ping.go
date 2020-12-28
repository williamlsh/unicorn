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
	"time"
)

type stats struct {
	addr                  string
	min, avg, max, stddev float64 // calculation results
}

func ping(ctx context.Context, addr string, timeout time.Duration, count int) (stats, error) {
	if count < 1 {
		return stats{}, errors.New("ping count must be larger than 0")
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

		fmt.Printf("Ping %s: seq=%d time=%.3fms\n", addr, i+1, roundtrip)
	}

	rpt := calculate(results)
	rpt.addr = addr

	return rpt, nil
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
		"",
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
	for _, s := range subscriptions {
		rpt, err := ping(ctx, s.Host, timeout*time.Second, count)
		if err != nil {
			return report{}, err
		}
		fmt.Printf("Report: %+v\n\n", rpt)

		// Filter hosts that is unable to ping.
		if rpt.avg > 0 {
			origReport = append(origReport, rpt)

			// Group reports by country.
			country := s.Fragment[:strings.Index(s.Fragment, "-")]
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
	for _, r := range data {
		fmt.Printf("Ping rank: %+v\n", r)
	}
}

func sortByCountry(data map[string][]stats) {
	for k, g := range data {
		sort.SliceStable(g, func(i, j int) bool {
			return g[i].avg < g[j].avg
		})
		fmt.Printf("\nReport of %s:\n", k)
		for _, r := range g {
			fmt.Printf("%+v\n", r)
		}
	}
}
