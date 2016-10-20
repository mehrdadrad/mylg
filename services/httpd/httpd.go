// Package httpd provides web service
package httpd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	// statik is single binary including all web stuff
	_ "github.com/mehrdadrad/mylg/services/dashboard/statik"
)

type TTracker struct {
	ch   chan string
	host string
}

// Route represents a HTTP route
type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

// APIHandler represents API function w/ cli arg
type APIHandler func(w http.ResponseWriter, r *http.Request, cfg cli.Config)

var ttracker = make(map[int]TTracker)

// APIWrapper wraps API func including cli arg
func APIWrapper(handler APIHandler, cfg cli.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, cfg)
	}
}

// API handles API routes
func API(w http.ResponseWriter, r *http.Request, cfg cli.Config) {
	var f = mux.Vars(r)["name"]

	w.Header().Set("Content-Type", "application/json")

	switch f {
	case "ping":
		ping(w, r, &cfg)
	case "init.trace":
		initTrace(w, r, &cfg)
	case "get.trace":
		getTrace(w, r)
	case "close.trace":
		closeTrace(w, r)
	}
}

// Run starts web service
func Run(cfg cli.Config) {
	statikFS, _ := fs.New()
	router := mux.NewRouter().StrictSlash(true)
	routes := []Route{
		{
			"API",
			"GET",
			"/api/{name}",
			APIWrapper(API, cfg),
		},
	}

	for _, route := range routes {
		router.
			Path(route.Path).
			Methods(route.Method).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	router.PathPrefix("/").Handler(http.FileServer(statikFS))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Web.Address, cfg.Web.Port), router)
	if err != nil {
		println(err.Error())
	}
}

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
	var errMsg = "trace channel is not exist"

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
	var errMsg = "trace channel is not exist"

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
