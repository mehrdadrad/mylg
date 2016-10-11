package ping_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/http/ping"
)

func TestNewPing(t *testing.T) {
	cfg, _ := cli.ReadDefaultConfig()
	url, _ := ping.NewPing("help", cfg)
	if url != nil {
		t.Error("NewPing expected nil but returned string")
	}
	_, err := ping.NewPing(".", cfg)
	if err == nil {
		t.Error("Newping expected error but it didn't return")
	}
}

func TestPing(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "test")
	}

	ts := httptest.NewServer(http.HandlerFunc(testHandler))
	defer ts.Close()

	cfg, _ := cli.ReadDefaultConfig()
	p, _ := ping.NewPing(ts.URL, cfg)
	r, _ := p.Ping()

	if r.StatusCode != 200 {
		t.Error("PingGet expected to get 200 but didn't, I got:", r.StatusCode)
	}

	if r.TotalTime == 0 {
		t.Error("PingGet expected to set totaltime but it didn't")
	}
}

func TestNormalize(t *testing.T) {
	n := ping.Normalize("google.com")
	if n != "http://google.com" {
		t.Error("Normalize retured unexpected value")
	}
	n = ping.Normalize("http://google.com")
	if n != "http://google.com" {
		t.Error("Normalize retured unexpected value")
	}
	n = ping.Normalize("https://google.com")
	if n != "https://google.com" {
		t.Error("Normalize retured unexpected value")
	}
}
