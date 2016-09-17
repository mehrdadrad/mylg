// Package httpd provides web service
package httpd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"net/http"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	// statik is single binary including all web stuff
	_ "github.com/mehrdadrad/mylg/services/dashboard/statik"
)

// Route represents a HTTP route
type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

// APIHandler represents API function w/ cli arg
type APIHandler func(w http.ResponseWriter, r *http.Request, cfg cli.Config)

// APIWrapper wraps API func including cli arg
func APIWrapper(handler APIHandler, cfg cli.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, cfg)
	}
}

// API handles API routes
func API(w http.ResponseWriter, r *http.Request, cfg cli.Config) {
	var (
		f      = mux.Vars(r)["name"]
		errStr string
	)
	switch f {
	case "ping":
		r.ParseForm()
		host := r.FormValue("host")
		p, err := icmp.NewPing(host+" -c 1", cfg)
		if err != nil {
			return
		}
		resp := p.Run()
		r := <-resp
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		w.Write([]byte(fmt.Sprintf(`{"rtt": %.2f,"pl":0, "err": "%s"}`, r.RTT, errStr)))
	}
}

// Run starts web service
func Run(cfg cli.Config) {
	statikFS, _ := fs.New()
	router := mux.NewRouter().StrictSlash(true)
	routes := []Route{
		{
			"API",
			"POST",
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
