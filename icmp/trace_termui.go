package icmp

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"strings"
	"time"

	ui "github.com/gizak/termui"

	"github.com/mehrdadrad/mylg/ripe"
)

// Widgets represents termui widgets
type Widgets struct {
	Hops *ui.List
	ASN  *ui.List
	RTT  *ui.List
	Snt  *ui.List
	Pkl  *ui.List

	Menu   *ui.Par
	Header *ui.Par

	LCRTT *ui.LineChart
	BCPKL *ui.BarChart
}

// Geo represents IP Geo params
type Geo struct {
	// CountrySrc Geo country source
	CountrySrc string
	// CountryDst Geo country destination
	CountryDst string
	// CitySrc Geo country source
	CitySrc string
	// CityDst Geo country source
	CityDst string
	// Latitude Geo source
	LatSrc float64
	// Latitude Geo destination
	LatDst float64
	// Longitude Geo source
	LonSrc float64
	// Longitude Geo destination
	LonDst float64
	// Distance holds src to dst distance
	Distance float64
}

// TermUI prints out trace loop by termui
func (i *Trace) TermUI() (string, error) {
	ui.DefaultEvtStream = ui.NewEvtStream()
	if err := ui.Init(); err != nil {
		return "", err
	}
	defer ui.Close()

	uiTheme(i.uiTheme)

	var (
		done    = make(chan struct{}, 2)
		routers = make([]map[string]Stats, 65)
		stats   = make([]Stats, 65)

		begin    = time.Now()
		w        = initWidgets()
		rp       string
		rChanged bool
	)

	// init widgets parameters w/ trace info
	i.bridgeWidgetsTrace(w)

	// run loop trace route
	resp, err := i.MRun()
	if err != nil {
		return rp, err
	}

	for i := 1; i < 65; i++ {
		routers[i] = make(map[string]Stats, 30)
	}

	screen1, screen2 := w.makeScreens()
	w.eventsHandler(done, screen1, screen2, stats)

	// update header each second
	go w.updateHeader(i, begin, done)

	// init layout
	ui.Body.AddRows(screen1...)
	ui.Body.Align()
	ui.Render(ui.Body)

	go func() {
		var (
			hop, as, holder string
		)
	LOOP:
		for {
			select {
			case <-done:
				break LOOP
			case r, ok := <-resp:
				if !ok {
					break LOOP
				}

				if r.hop != "" {
					hop = r.hop
				} else {
					hop = r.ip
				}

				if r.whois.asn > 0 {
					as = fmt.Sprintf("%.0f", r.whois.asn)
					holder = strings.Fields(r.whois.holder)[0]
				} else {
					as = ""
					holder = ""
				}

				// statistics
				stats[r.num].count++
				w.Snt.Items[r.num] = fmt.Sprintf("%d", stats[r.num].count)

				router := routers[r.num][hop]
				router.count++

				if r.elapsed != 0 {

					// hop level statistics
					calcStatistics(&stats[r.num], r.elapsed)
					// router level statistics
					calcStatistics(&router, r.elapsed)
					// detect router changes
					rChanged = routerChange(hop, w.Hops.Items[r.num])

					w.Hops.Items[r.num] = fmt.Sprintf("[%-2d] %s", r.num, hop)
					w.ASN.Items[r.num] = fmt.Sprintf("%-6s %s", as, holder)
					w.RTT.Items[r.num] = fmt.Sprintf("%-6.2f\t%-6.2f\t%-6.2f\t%-6.2f", r.elapsed, stats[r.num].avg, stats[r.num].min, stats[r.num].max)

					if rChanged {
						w.Hops.Items[r.num] = termUICColor(w.Hops.Items[r.num], "fg-bold")
					}

					lcShift(r, w.LCRTT, ui.TermWidth())

				} else if w.Hops.Items[r.num] == "" {

					w.Hops.Items[r.num] = fmt.Sprintf("[%-2d] %-40s", r.num, "???")
					stats[r.num].pkl++
					router.pkl++

				} else if !strings.Contains(w.Hops.Items[r.num], "???") {

					hop = rmUIMetaData(w.Hops.Items[r.num])
					hop = fmt.Sprintf("[%-2d] %s", r.num, hop)
					w.Hops.Items[r.num] = termUICColor(hop, "fg-red")
					w.RTT.Items[r.num] = fmt.Sprintf("%-6.2s\t%-6.2f\t%-6.2f\t%-6.2f", "?", stats[r.num].avg, stats[r.num].min, stats[r.num].max)
					stats[r.num].pkl++
					router.pkl++

				} else {
					w.Hops.Items[r.num] = fmt.Sprintf("[%-2d] %s", r.num, "???")
					stats[r.num].pkl++
					router.pkl++

				}

				if len(w.BCPKL.DataLabels) > r.num-1 {
					w.BCPKL.DataLabels[r.num-1] = fmt.Sprintf("H%d", r.num)
					w.BCPKL.Data[r.num-1] = int(stats[r.num].pkl)
				} else {
					w.BCPKL.DataLabels = append(w.BCPKL.DataLabels, fmt.Sprintf("H%d", r.num))
					w.BCPKL.Data = append(w.BCPKL.Data, int(stats[r.num].pkl))
				}

				routers[r.num][hop] = router

				w.Pkl.Items[r.num] = fmt.Sprintf("%.1f", float64(stats[r.num].pkl)*100/float64(stats[r.num].count))
				ui.Render(ui.Body)
				// clean up in case of packet loss on the last hop at first try
				if r.last {
					for i := r.num + 1; i < 65; i++ {
						w.Hops.Items[i] = ""
					}
				}
			}
		}
		if _, ok := <-resp; ok {
			close(resp)
		}
	}()

	ui.Loop()

	if i.report {
		rp = report(w, i)
	}

	return rp, nil
}

func (i *Trace) bridgeWidgetsTrace(w *Widgets) {
	// barchart
	w.LCRTT.BorderLabel = fmt.Sprintf("RTT: %s", i.host)
	// title
	t := fmt.Sprintf(
		"──[ myLG ]── traceroute to %s (%s), %d hops max, elapsed: 0s",
		i.host,
		i.ip,
		i.maxTTL,
	)
	t += strings.Repeat(" ", 20)
	w.Header.Text = t
}

// lcShift shifs line chart once it filled out
func lcShift(r HopResp, lc *ui.LineChart, width int) {
	if r.last {
		t := time.Now()
		lc.Data = append(lc.Data, r.elapsed)
		lc.DataLabels = append(lc.DataLabels, t.Format("04:05"))
		if len(lc.Data) > (ui.TermWidth()/2)-10 {
			lc.Data = lc.Data[1:]
			lc.DataLabels = lc.DataLabels[1:]
		}
	}
}

func rttWidget() *ui.LineChart {
	lc := ui.NewLineChart()
	lc.Height = 18
	lc.Mode = "dot"

	return lc
}

func pktLossWidget() *ui.BarChart {
	bc := ui.NewBarChart()
	bc.BorderLabel = "Packet Loss per hop"
	bc.Height = 18
	bc.TextColor = ui.ColorGreen
	bc.BarColor = ui.ColorRed
	bc.NumColor = ui.ColorYellow

	return bc
}

func headerWidget() *ui.Par {
	h := ui.NewPar("")
	h.Height = 1
	h.Width = ui.TermWidth()
	h.Y = 1
	h.TextBgColor = ui.ColorCyan
	h.TextFgColor = ui.ColorBlack
	h.Border = false

	return h
}

func menuWidget() *ui.Par {
	var items = []string{
		"Press [q] to quit",
		"[r] to reset statistics",
		"[1,2] to change display mode",
	}

	m := ui.NewPar(strings.Join(items, ", "))
	m.Height = 1
	m.Width = 20
	m.Y = 1
	m.Border = false

	return m
}

func (w *Widgets) eventsHandler(done chan struct{}, s1, s2 []*ui.Row, stats []Stats) {
	// exit
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		done <- struct{}{}
		done <- struct{}{}
		ui.StopLoop()
	})

	// change display mode to one
	ui.Handle("/sys/kbd/1", func(e ui.Event) {
		ui.Clear()
		ui.Body.Rows = ui.Body.Rows[:0]
		ui.Body.AddRows(s1...)
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	// change display mode to two
	ui.Handle("/sys/kbd/2", func(e ui.Event) {
		ui.Clear()
		ui.Body.Rows = ui.Body.Rows[:0]
		ui.Body.AddRows(s2...)
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	// resize
	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	// reset statistics and display
	ui.Handle("/sys/kbd/r", func(ui.Event) {
		for i := 1; i < 65; i++ {
			w.Hops.Items[i] = ""
			w.ASN.Items[i] = ""
			w.RTT.Items[i] = ""
			w.Snt.Items[i] = ""
			w.Pkl.Items[i] = ""

			stats[i].count = 0
			stats[i].avg = 0
			stats[i].min = 0
			stats[i].max = 0
			stats[i].pkl = 0
		}
		w.LCRTT.Data = w.LCRTT.Data[:0]
		w.LCRTT.DataLabels = w.LCRTT.DataLabels[:0]
	})

}

func (w *Widgets) makeScreens() ([]*ui.Row, []*ui.Row) {
	// screens1 - trace statistics
	screen1 := []*ui.Row{
		ui.NewRow(
			ui.NewCol(12, 0, w.Header),
		),
		ui.NewRow(
			ui.NewCol(12, 0, w.Menu),
		),
		ui.NewRow(
			ui.NewCol(5, 0, w.Hops),
			ui.NewCol(2, 0, w.ASN),
			ui.NewCol(1, 0, w.Pkl),
			ui.NewCol(1, 0, w.Snt),
			ui.NewCol(3, 0, w.RTT),
		),
	}
	// screen2 - trace line chart
	screen2 := []*ui.Row{
		ui.NewRow(
			ui.NewCol(12, 0, w.Header),
		),
		ui.NewRow(
			ui.NewCol(12, 0, w.Menu),
		),
		ui.NewRow(
			ui.NewCol(6, 0, w.LCRTT),
		),
		ui.NewRow(
			ui.NewCol(6, 0, w.BCPKL),
		),
	}

	return screen1, screen2
}

func (w *Widgets) updateHeader(i *Trace, begin time.Time, done chan struct{}) {
	var (
		c       = time.Tick(1 * time.Second)
		geo     Geo
		unit            = "miles"
		eRadius float64 = 3961
	)

	geo.CitySrc = "..."
	geo.CityDst = "..."

	go getGeo(i.ip, &geo)

	if i.km {
		unit = "km"
		eRadius = 6373
	}
LOOP:
	for {
		select {
		case <-done:
			break LOOP
		case <-c:
			h := strings.Split(w.Header.Text, "hops max")
			if len(h) < 1 {
				break LOOP
			}
			seconds := fmt.Sprintf("%.0fs", time.Since(begin).Seconds())
			du, _ := time.ParseDuration(seconds)
			s := fmt.Sprintf("%shops max, elapsed: %s %s (%s) -> %s (%s) %.0f %s ",
				h[0],
				du.String(),
				geo.CitySrc,
				geo.CountrySrc,
				geo.CityDst,
				geo.CountryDst,
				distance(geo, eRadius),
				unit,
			)
			w.Header.Text = s
			ui.Render(ui.Body)
		}
	}
}

func getGeo(ipDst net.IP, geo *Geo) {
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
		p.Set(ipDst.String())
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

	}
}
func initWidgets() *Widgets {
	var (
		hops = ui.NewList()
		asn  = ui.NewList()
		rtt  = ui.NewList()
		snt  = ui.NewList()
		pkl  = ui.NewList()

		lists = []*ui.List{hops, asn, rtt, snt, pkl}
	)

	for _, l := range lists {
		l.Items = make([]string, 65)
		l.X = 0
		l.Y = 0
		l.Height = 35
		l.Border = false
	}

	// title
	hops.Items[0] = fmt.Sprintf("[%s](fg-bold)", "Host")
	asn.Items[0] = fmt.Sprintf("[ %-6s %-6s](fg-bold)", "ASN", "Holder")
	rtt.Items[0] = fmt.Sprintf("[%-6s %-6s %-6s %-6s](fg-bold)", "Last", "Avg", "Best", "Wrst")
	snt.Items[0] = "[Sent](fg-bold)"
	pkl.Items[0] = "[Loss%](fg-bold)"

	return &Widgets{
		Hops: hops,
		ASN:  asn,
		RTT:  rtt,
		Snt:  snt,
		Pkl:  pkl,

		Menu:   menuWidget(),
		Header: headerWidget(),
		LCRTT:  rttWidget(),
		BCPKL:  pktLossWidget(),
	}
}

func uiTheme(t string) {

	switch t {
	case "light":
		ui.ColorMap["bg"] = ui.ColorWhite
		ui.ColorMap["fg"] = ui.ColorBlack
		ui.ColorMap["label.fg"] = ui.ColorBlack | ui.AttrBold
		ui.ColorMap["linechart.axes.fg"] = ui.ColorBlack
		ui.ColorMap["linechart.line.fg"] = ui.ColorGreen
		ui.ColorMap["border.fg"] = ui.ColorBlue
	default:
		// dark theme
		ui.ColorMap["bg"] = ui.ColorBlack
		ui.ColorMap["fg"] = ui.ColorWhite
		ui.ColorMap["label.fg"] = ui.ColorWhite | ui.AttrBold
		ui.ColorMap["linechart.axes.fg"] = ui.ColorWhite
		ui.ColorMap["linechart.line.fg"] = ui.ColorGreen
		ui.ColorMap["border.fg"] = ui.ColorCyan
	}
	ui.Clear()
}

func report(w *Widgets, i *Trace) string {
	var (
		r      string
		format = "%-45s %-25s %-5s %-6s %s\n"
	)

	r = fmt.Sprintf("──[ myLG ]── traceroute to %s (%s)\n",
		i.host,
		i.ip,
	)

	r += fmt.Sprintf(format,
		"Host",
		"ASN    Holder",
		"Sent",
		"Lost%",
		"Last       Avg     Best    Wrst",
	)

	for i := 1; i < 65; i++ {
		if w.Hops.Items[i] != "" {

			w.Hops.Items[i] = rmUIMetaData(w.Hops.Items[i])
			w.Hops.Items[i] = trimLongStr(w.Hops.Items[i], 40)

			r += fmt.Sprintf(format,
				w.Hops.Items[i],
				w.ASN.Items[i],
				w.Snt.Items[i],
				w.Pkl.Items[i],
				w.RTT.Items[i],
			)
		}
	}
	return r
}
func trimLongStr(s string, l int) string {
	if len(s) > l {
		return s[:l] + "..."
	}
	return s
}

func rmUIMetaData(m string) string {
	var rgx = []string{`\[+\d+\s*\]\s`, `\]\(.*\)`}
	for _, r := range rgx {
		re := regexp.MustCompile(r)
		m = re.ReplaceAllString(m, "")
	}
	return m
}

func termUICColor(m, color string) string {
	if !strings.Contains(m, color) {
		m = fmt.Sprintf("[%s](%s)", m, color)
	}
	return m
}

func distance(geo Geo, r float64) float64 {

	geo.LonSrc = d2r(geo.LonSrc)
	geo.LonDst = d2r(geo.LonDst)
	geo.LatSrc = d2r(geo.LatSrc)
	geo.LatDst = d2r(geo.LatDst)

	deltaLon := geo.LonDst - geo.LonSrc
	deltaLat := geo.LatDst - geo.LatSrc

	a := hsin(deltaLat) + math.Cos(geo.LatDst)*math.Cos(geo.LatSrc)*hsin(deltaLon)
	c := 2 * r * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c
}
func d2r(i float64) float64 {
	return i * math.Pi / 180
}
func hsin(i float64) float64 {
	return math.Pow(math.Sin(i/2), 2)
}
