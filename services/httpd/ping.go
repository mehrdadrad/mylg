package httpd

import (
	"fmt"
	"net/http"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
)

// ping tries to ping a host (count = 1)
func ping(w http.ResponseWriter, r *http.Request, cfg *cli.Config) {
	var errStr string

	r.ParseForm()
	host := r.FormValue("host")

	if len(host) < 5 {
		fmt.Fprintf(w, `{"err": "%s"}`, "host is invalid")
		return
	}

	p, err := icmp.NewPing(host+" -c 1", *cfg)
	if err != nil {
		return
	}

	resp := p.Run()
	rs := <-resp
	if rs.Error != nil {
		errStr = rs.Error.Error()
	}
	fmt.Fprintf(w, `{"rtt": %.2f,"pl":0, "err": "%s"}`, rs.RTT, errStr)
}
