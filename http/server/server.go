// Package server provides web service
package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Route represents a HTTP route
type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

var routes = []Route{
	Route{
		"API",
		"POST",
		"/api/{name}",
		API,
	},
	{
		"root",
		"GET",
		"/",
		Root,
	},
}

// Root handles root request
func Root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.URL.Path)
}

// API handles API routes
func API(w http.ResponseWriter, r *http.Request) {
	var f = mux.Vars(r)["name"]
	switch f {
	case "ping":
		w.Write([]byte(`{"rtt": 10,"pl":0}`))
	}
}

// Run starts web service
func Run() {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Path(route.Path).
			Methods(route.Method).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./html/")))
	http.ListenAndServe("127.0.0.1:8080", router)
}
