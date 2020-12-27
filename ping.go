package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"time"
)

type report struct {
	min, avg, max, stddev float64 // calculation results
}

func ping(ctx context.Context, addr string, timeout time.Duration, count int) (report, error) {
	if count < 1 {
		return report{}, errors.New("ping count must be larger than 0")
	}

	results := make([]float64, count)
	for i := 0; i < count; i++ {
		startAt := time.Now()

		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			// Skip timed out ping.
			results = append(results, 0)
			fmt.Printf("Ping %s timeout: %v\n", addr, err)
			break
		}
		_ = conn.Close()

		endAt := time.Now()
		roundtrip := float64(endAt.UnixNano()-startAt.UnixNano()) / 1e6
		results[i] = roundtrip

		fmt.Printf("Ping %s: seq=%d time=%.3fms\n", addr, i+1, roundtrip)
	}

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

	return report{
		min,
		max,
		avg,
		stddev,
	}, nil
}
