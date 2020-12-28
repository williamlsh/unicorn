package main

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	t.Parallel()

	trueURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	rpt, err := ping(context.Background(), trueURL.Host, 3*time.Second, 3)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", rpt)
}
