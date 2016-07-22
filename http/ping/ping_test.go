package ping_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/http/ping"
	"gopkg.in/h2non/gock.v0"
)

func TestNewPing(t *testing.T) {
	url, _ := ping.NewPing("help")
	if url != nil {
		t.Error("NewPing expected nil but returned string")
	}
	_, err := ping.NewPing("garbage")
	if err == nil {
		t.Error("Newping expected error but it didn't return")
	}
}

func TestPingGet(t *testing.T) {
	var url = "mylg.io"
	gock.New(url).
		Reply(200)

	p, _ := ping.NewPing(url)
	r, _ := p.PingGet()
	if r.StatusCode != 200 {
		t.Error("PingGet expected to get 200 but didn't")
	}
	if r.TotalTime == 0 {
		t.Error("PingGet expected to set totaltime but it didn't")
	}
}
func TestPingPost(t *testing.T) {
	var url = "mylg.io"
	gock.New(url).
		Reply(200)

	p, _ := ping.NewPing(url)
	r, _ := p.PingPost()
	if r.StatusCode != 200 {
		t.Error("PingPost expected to get 200 but didn't")
	}
	if r.TotalTime == 0 {
		t.Error("PingPost expected to set totaltime but it didn't")
	}
}
func TestPingHead(t *testing.T) {
	var url = "mylg.io"
	gock.New(url).
		Reply(200)

	p, _ := ping.NewPing(url)
	r, _ := p.PingHead()
	if r.StatusCode != 200 {
		t.Error("PingHead expected to get 200 but didn't")
	}
	if r.TotalTime == 0 {
		t.Error("PingHead expected to set totaltime but it didn't")
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
