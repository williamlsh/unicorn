package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	server = httptest.NewTLSServer(http.HandlerFunc(handleSubscription))
	os.Exit(m.Run())
	server.Close()

}

func handleSubscription(w http.ResponseWriter, r *http.Request) {
	content :=
		`trojan://mypassword1@host:port#US-us1
trojan://mypassword2@host:port#US-us2
trojan://mypassword3@host:port#US-us3
trojan://mypassword4@host:port#UK-uk1
trojan://mypassword5@host:port#UK-uk2
trojan://mypassword6@host:port#UK-uk3
trojan://mypassword7@host:port#JP-jp1
trojan://mypassword8@host:port#JP-jp2
trojan://mypassword9@host:port#JP-jp3
`

	w.Header().Set("content-type", "text/plain;charset=utf-8")

	encoder := base64.NewEncoder(base64.StdEncoding, w)
	replacer := strings.NewReplacer("host:port", server.URL[len("https://"):])
	if _, err := replacer.WriteString(encoder, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
