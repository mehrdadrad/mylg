package cli_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/cli"
)

var c *cli.Readline

func TestSetPrompt(t *testing.T) {
	c = cli.Init("test")
	c.SetPrompt("mylg")
	p := c.GetPrompt()
	if p != "mylg" {
		t.Error("Set cli prompt failed")
	}
}

func TestFlagHelp(t *testing.T) {
	_, flag := cli.Flag("www.mylg.io help")
	if _, ok := flag["help"]; !ok {
		t.Error("flag expected help request but not exist")
	}
}

func TestFlagSimple(t *testing.T) {
	url, flag := cli.Flag("www.mylg.io")
	if _, ok := flag["help"]; ok {
		t.Error("flag expected none help")
	}
	if url != "www.mylg.io" {
		t.Error("flag unexpected url")
	}
}

func TestFlagStringValue(t *testing.T) {
	url, flag := cli.Flag("www.mylg.io -option=value")
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
}

func TestFlagIntValue(t *testing.T) {
	url, flag := cli.Flag("www.mylg.io -option=1976")
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
}

func TestFlagBoolianOptions(t *testing.T) {
	url, flag := cli.Flag("www.mylg.io -n -nr")
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
}

func TestFlagDashInMiddle(t *testing.T) {
	// test dash in middle at value
	url, flag := cli.Flag("www.mylg.io -p 1-100")
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
		t.Error("flag unexpected url")
	}
}

func TestFlagTargetWithSpace(t *testing.T) {
	rule, flag := cli.Flag("tcp and port 443 -i en0")
	if _, ok := flag["i"]; !ok {
		t.Error("")
	}
	if rule != "tcp and port 443" {
		t.Error("")
	}
}
