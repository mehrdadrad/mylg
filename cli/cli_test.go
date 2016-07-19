package cli_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/cli"
)

var c *cli.Readline

func TestInitPrompt(t *testing.T) {
	c = cli.Init("local", "test")
	p := c.GetPrompt()
	if p != "local" {
		t.Error("Init cli prompt failed")
	}
}

func TestSetPrompt(t *testing.T) {
	c = cli.Init("local", "test")
	c.SetPrompt("mylg")
	p := c.GetPrompt()
	if p != "mylg" {
		t.Error("Set cli prompt failed")
	}
}
