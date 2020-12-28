package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, errors.New("response status not ok")
	}

	if resp.Header.Get("content-type") != "text/plain;charset=utf-8" {
		_ = resp.Body.Close()
		return nil, errors.New("content type not supported")
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
		trueURL, err := url.Parse(scanner.Text())
		if err != nil {
			return nil, err
		}

		if err := validateTrojan(trueURL); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, trueURL)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func validateTrojan(trueURL *url.URL) error {
	if trueURL.Scheme != "trojan" {
		return fmt.Errorf("invalid subscription URL: %s", trueURL.Scheme)
	}

	// Password in trojan URL is username in standard URL scheme.
	if trueURL.User.Username() == "" {
		return fmt.Errorf("no password in subscription URL")
	}

	if trueURL.Hostname() == "" {
		return fmt.Errorf("no host in subscription URL")
	}

	if trueURL.Port() == "" {
		return fmt.Errorf("no port in subscription URL")
	}

	if trueURL.Fragment == "" {
		return fmt.Errorf("no fragment in subscription URL")
	}

	if strings.Index(trueURL.Fragment, "-") < 1 {
		return fmt.Errorf("URL fragment schema invalid: %s", trueURL.Fragment)
	}

	return nil
}
