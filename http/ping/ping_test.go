package ping_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/http/ping"
)

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
