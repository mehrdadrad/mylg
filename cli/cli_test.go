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

func TestFlag(t *testing.T) {
	// test help
	url, flag := cli.Flag("www.mylg.io help")
	if _, ok := flag["help"]; !ok {
		t.Error("flag expected help request but not exist")
	}
	// test simple arg
	url, flag = cli.Flag("www.mylg.io")
	if _, ok := flag["help"]; ok {
		t.Error("flag expected none help")
	}
	if url != "www.mylg.io" {
		t.Error("flag unexpected url")
	}
	// test option and string value
	url, flag = cli.Flag("www.mylg.io -option=value")
	if _, ok := flag["option"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["option"] != "value" {
		t.Error("flag unexpected string value")
	}
	if url != "www.mylg.io" {
		t.Error("")
		println("url", url)
	}
	url, flag = cli.Flag("www.mylg.io -option value")
	if _, ok := flag["option"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["option"] != "value" {
		t.Error("flag unexpected string value")
	}
	if url != "www.mylg.io" {
		t.Error("")
		println("url", url)
	}
	// test option and int value
	url, flag = cli.Flag("www.mylg.io -option=1976")
	if _, ok := flag["option"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["option"] != 1976 {
		t.Error("flag unexpected int value")
	}
	if url != "www.mylg.io" {
		t.Error("")
		println("url", url)
	}
	url, flag = cli.Flag("www.mylg.io -option 1976")
	if _, ok := flag["option"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["option"] != 1976 {
		t.Error("flag unexpected int value")
	}
	if url != "www.mylg.io" {
		t.Error("")
		println("url", url)
	}
	// test two boolean options sequentially
	url, flag = cli.Flag("www.mylg.io -n -nr")
	if _, ok := flag["n"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["n"] != "" {
		t.Error("flag unexpected int value")
	}
	if _, ok := flag["nr"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["nr"] != "" {
		t.Error("flag unexpected int value")
	}
	if url != "www.mylg.io" {
		t.Error("")
		println("url", url)
	}
	// test dash in middle at value
	url, flag = cli.Flag("www.mylg.io -p 1-100")
	if _, ok := flag["p"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["p"] != "1-100" {
		t.Error("flag unexpected int value")
	}
	// test dash at target
	url, flag = cli.Flag("www.my-lg.io -p 1-100")
	if _, ok := flag["p"]; !ok {
		t.Error("flag unexpected option")
	} else if flag["p"] != "1-100" {
		t.Error("flag unexpected int value")
	}
	if url != "www.my-lg.io" {
		t.Error("")
		println("url", url)
	}

}
