package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	server = httptest.NewTLSServer(http.HandlerFunc(handleSubscription))
	os.Exit(m.Run())
	server.Close()

}

func handleSubscription(w http.ResponseWriter, r *http.Request) {
	sampleConfig :=
		`trojan://mypassword1@localhost:443#US-us1
trojan://mypassword2@localhost:443#US-us2
trojan://mypassword3@localhost:443#US-us3
trojan://mypassword4@localhost:443#UK-uk1
trojan://mypassword5@localhost:443#UK-uk2
trojan://mypassword6@localhost:443#UK-uk3
trojan://mypassword7@localhost:443#JP-jp1
trojan://mypassword8@localhost:443#JP-jp2
trojan://mypassword9@localhost:443#JP-jp3
`

	w.Header().Set("content-type", "text/plain;charset=utf-8")
	if _, err := base64.NewEncoder(base64.StdEncoding, w).Write([]byte(sampleConfig)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
