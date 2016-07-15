package ripe_test

import (
	"github.com/mehrdadrad/mylg/ripe"
	"gopkg.in/h2non/gock.v0"
	"testing"
)

func TestRipePrefixAPISCode(t *testing.T) {
	gock.New(ripe.RIPEAPI).
		Reply(200).
		JSON(map[string]string{"status": "ok"})

	var p ripe.Prefix
	p.Set("8.8.8.0/24")
	if !p.GetData() {
		t.Error("failed on http 200")
	}

	gock.New(ripe.RIPEAPI).
		Reply(400).
		JSON(map[string]string{"status": "ok"})

	if p.GetData() {
		t.Error("failed on none http 200")
	}
}

func TestRipeASNAPISCode(t *testing.T) {
	gock.New(ripe.RIPEAPI).
		Reply(200).
		JSON(map[string]string{"status": "ok"})

	var a ripe.ASN
	a.Set("577")
	if !a.GetData() {
		t.Error("failed on http 200")
	}

	gock.New(ripe.RIPEAPI).
		Reply(400).
		JSON(map[string]string{"status": "ok"})

	if a.GetData() {
		t.Error("failed on none http 200")
	}
}
