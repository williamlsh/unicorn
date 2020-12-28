package main

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	t.Parallel()

	trueURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan stats)

	var wg sync.WaitGroup
	wg.Add(1)
	go ping(context.Background(), trueURL, 3*time.Second, 3, &wg, ch)
	go wg.Wait()

	rpt := <-ch
	if rpt.err != nil {
		t.Error(err)
	}
}

func TestPingAll(t *testing.T) {
	t.Parallel()

	resp, err := fetch(server.Client(), server.URL)
	if err != nil {
		t.Skip(err)
	}
	defer resp.Body.Close()

	subscriptions, err := parseSubscriptions(resp)
	if err != nil {
		t.Error(err)
	}

	result, err := pingAll(context.Background(), subscriptions)
	if err != nil {
		t.Error(err)
	}

	result.sortAll()
	result.sortByCountry()
}
