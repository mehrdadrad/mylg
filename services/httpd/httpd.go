// Package httpd provides web service
package httpd

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/rakyll/statik/fs"
	"net/http"

	"github.com/mehrdadrad/mylg/cli"
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
	case "geo":
		getGeo(w, r)
	}
}

// Run starts web service
func Run(cfg cli.Config) {
	//statikFS, _ := fs.New()
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
	//router.PathPrefix("/").Handler(http.FileServer(statikFS))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/Users/mehrdad/golang/src/github.com/mehrdadrad/myLG/services/dashboard/assets/")))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Web.Address, cfg.Web.Port), router)
	if err != nil {
		println(err.Error())
	}
}
