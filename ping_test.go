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
