package httpd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mehrdadrad/mylg/ripe"
)

func getGeo(w http.ResponseWriter, r *http.Request) {
	var geo = struct {
		CitySrc    string
		CityDst    string
		CountrySrc string
		CountryDst string

		LatSrc float64
		LonSrc float64
		LatDst float64
		LonDst float64
	}{}

	r.ParseForm()
	ipDst := r.FormValue("ip")

	// find station public ip address and geo
	ip, err := ripe.MyIPAddr()
	// find src, dst geo
	if err == nil {
		p := new(ripe.Prefix)
		p.Set(ip)
		p.GetGeoData()
		for _, g := range p.GeoData.Data.Locations {
			if g.City != "" {
				geo.CitySrc = g.City
				geo.CountrySrc = g.Country
				geo.LatSrc = g.Latitude
				geo.LonSrc = g.Longitude
				break
			}
		}
		p.Set(ipDst)
		p.GetGeoData()
		for _, g := range p.GeoData.Data.Locations {
			if g.City != "" {
				geo.CityDst = g.City
				geo.CountryDst = g.Country
				geo.LatDst = g.Latitude
				geo.LonDst = g.Longitude
				break
			}
		}
		b, _ := json.Marshal(geo)
		fmt.Fprintf(w, string(b))
	}
}
