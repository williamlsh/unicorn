package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

// newClient returns a new HTTP client with HTTP proxy forwarding if httpProxy is not nil
// else it returns a HTTP default client with proxy.
func newClient(httpProxy string) (*http.Client, error) {
	// If httpProxy is not not specified, don't use http proxy to forward request.
	if httpProxy == "" {
		fmt.Println("No proxy mod")
		return http.DefaultClient, nil
	}

	fmt.Println("Proxy mod")
	proxy, err := url.Parse(httpProxy)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	client.Transport = &http.Transport{
		// Specify HTTP proxy url to forward client request.
		Proxy: http.ProxyURL(proxy),
	}

	return client, nil
}

// fetch fetches subscription with provided HTTP client and target URL.
func fetch(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// parseSubscriptions parses the response body and returns the final url.URLs.
// The response body content is base64 encoded containing text of URLs line by line.
// A subscription URL scheme is as such: trojan://password@remote_host:remote_port#country-xxx
func parseSubscriptions(resp *http.Response) ([]*url.URL, error) {
	var subscriptions []*url.URL

	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	scanner := bufio.NewScanner(decoder)
	for scanner.Scan() {
		trueURL, err := parseTextToURL(scanner.Text())
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, trueURL)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func parseTextToURL(text string) (*url.URL, error) {
	trueURL, err := url.Parse(text)
	if err != nil {
		return nil, err
	}
	if trueURL.Scheme != "trojan" {
		return nil, fmt.Errorf("invalid subscription URL: %s", trueURL.Scheme)
	}
	return trueURL, nil
}
