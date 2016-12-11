package httpd

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
)

// initTrace returns an id and creates a gorouting
func initTrace(w http.ResponseWriter, r *http.Request, cfg *cli.Config) {
	r.ParseForm()
	args := r.FormValue("a")

	id := rand.Intn(1000)
	ttracker[id] = TTracker{ch: make(chan string, 1)}

	t, err := icmp.NewTrace(args, *cfg)
	if err != nil {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, -1, err.Error())
		return
	}

	go func() {
		defer func() {
			recover()
		}()

		resp, _ := t.MRun()
		defer close(resp)

		for {
			select {
			case r, _ := <-resp:
				ttracker[id].ch <- r.Marshal()
			}
		}
	}()

	fmt.Fprintf(w, `{"id": %d, "err": ""}`, id)
}

func getTrace(w http.ResponseWriter, r *http.Request) {
	var errMsg = "trace channel does not exist"

	r.ParseForm()
	id := r.FormValue("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, i, err.Error())
		return
	}

	if _, ok := ttracker[i]; !ok {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, i, errMsg)
		return
	}

	if v, ok := <-ttracker[i].ch; ok {
		fmt.Fprint(w, v)
	} else {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, i, errMsg)
	}
}

func closeTrace(w http.ResponseWriter, r *http.Request) {
	var errMsg = "trace channel does not exist"

	r.ParseForm()
	id := r.FormValue("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, i, err.Error())
		return
	}

	if _, ok := ttracker[i]; !ok {
		fmt.Fprintf(w, `{"id": %d, "err": "%s"}`, i, errMsg)
		return
	}

	close(ttracker[i].ch)
	delete(ttracker, i)

	fmt.Fprintf(w, `{"id": %d, "err": ""}`, i)
}
