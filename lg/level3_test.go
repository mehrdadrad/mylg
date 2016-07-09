package lg_test

import (
	"github.com/mehrdadrad/mylg/lg"
	"gopkg.in/h2non/gock.v0"
	"testing"
)

func TestGetDefaultNode(t *testing.T) {
	var level3 lg.Level3
	if level3.GetDefaultNode() != "Los Angeles, CA" {
		t.Error("Level3 default node expected Los Angeles, CA but", level3.GetDefaultNode)
	}
}

func TestFetchNodes(t *testing.T) {
	gock.New("http://lookingglass.level3.net").
		Get("/ping/lg_ping_main.php").
		Reply(200).
		BodyString(`
				<th style="width:100px;">Packet Size</th><th style="width:100px;">Packet Count</th><th>IPv6</th><th>&nbsp;</th>
				</tr>
				<tr><td><SELECT name="sitename">
				<OPTGROUP Label="EMEA">
				<OPTION value="ear1.ams1">Amsterdam, Netherlands</OPTION>
				<OPTION value="bar1.bcl1">Barcellona, Spain</OPTION>
				<OPTION value="bear1.xrs2">Belgrade, Serbia</OPTION>
				</SELECT></td><td>
		`)
	var level3 lg.Level3
	nodes := level3.FetchNodes()
	if len(nodes) != 3 {
		t.Error("expected to have 3 nodes but they are", len(nodes))
	}
	for _, n := range nodes {
		if n != "ear1.ams1" && n != "bar1.bcl1" && n != "bear1.xrs2" {
			t.Error("expected to see the correct value but it is", n)
		}
	}
}

func TestPing(t *testing.T) {
	gock.New("http://lookingglass.level3.net").
		Post("/ping/lg_ping_output.php").
		Reply(200).
		BodyString(`</div></div>Ping results from Los Angeles, CA to 8.8.8.8(google-public-dns-a.google.com)<br><pre>` +
			`<font face="terminal" SIZE="1" color="#000000">icmp_seq=1 ttl=61 time=0.325 ms<br><br>---- target` +
			`statistics ----<br>1 packets transmitted, 1 packets received, 0% packet loss<br>` +
			`rtt min/avg/median/max/mdev/stddev = 0.297/0.311/0.307/0.325/0.097/0.01 ms<br></font></pre><br></div></body></html>
		`)
	var level3 lg.Level3
	level3.Set("127.0.0.1", "ipv4")
	p, err := level3.Ping()
	if err != nil {
		t.Error(err.Error())
	}
	if len(p) != 270 {
		t.Error("expected to see 270 length result but it is", len(p))
	}
}
