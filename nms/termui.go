package nms

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"

	ui "github.com/gizak/termui"
)

const (
	maxRowTermUI = 45
)

// Widgets represents termui widgets
type Widgets struct {
	header   *ui.Par
	footer   *ui.Par
	menu     *ui.Par
	ifName   *ui.List
	ifStatus *ui.List
	ifDescr  *ui.List
	ifTIn    *ui.List
	ifTOut   *ui.List
	ifPIn    *ui.List
	ifPOut   *ui.List
	ifDIn    *ui.List
	ifDOut   *ui.List
	ifEIn    *ui.List
	ifEOut   *ui.List
}

func initWidgets() *Widgets {
	return &Widgets{
		header:   ui.NewPar(""),
		footer:   ui.NewPar(""),
		menu:     ui.NewPar(""),
		ifName:   ui.NewList(),
		ifStatus: ui.NewList(),
		ifDescr:  ui.NewList(),
		ifTIn:    ui.NewList(),
		ifTOut:   ui.NewList(),
		ifPIn:    ui.NewList(),
		ifPOut:   ui.NewList(),
		ifDIn:    ui.NewList(),
		ifDOut:   ui.NewList(),
		ifEIn:    ui.NewList(),
		ifEOut:   ui.NewList(),
	}
}

func (w *Widgets) updateFrame(c *Client, err string) {
	var (
		h = fmt.Sprintf("──[ myLG ]── Quick NMS SNMP - %s [%s](fg-red,fg-bold)",
			c.SNMP.Host,
			err,
		)
		m = "Press [q] to quit"
	)

	if c := ui.TermWidth() - len(h) + 2 + 18; c > 0 {
		h = h + strings.Repeat(" ", c)
	}

	w.header.Width = ui.TermWidth()
	w.header.Height = 1
	w.header.Y = 0
	w.header.Text = h
	w.header.TextBgColor = ui.ColorCyan
	w.header.TextFgColor = ui.ColorBlack
	w.header.Border = false

	w.footer.Width = ui.TermWidth()
	w.footer.Height = 1
	w.footer.Text = strings.Repeat("─", ui.TermWidth()-6)
	w.footer.TextBgColor = ui.ColorDefault
	w.footer.TextFgColor = ui.ColorCyan
	w.footer.Border = false

	w.menu.Width = ui.TermWidth()
	w.menu.Height = 1
	w.menu.Y = 1
	w.menu.Text = m
	w.menu.TextFgColor = ui.ColorDefault
	w.menu.Border = false
}

func (c *Client) snmpShowInterfaceTermUI(filter string, flag map[string]interface{}) error {
	var (
		spin   = spinner.New(spinner.CharSets[26], 220*time.Millisecond)
		span   = []int{1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1}
		s1, s2 [][]string
		rows   []*ui.Row
		idxs   []int
		err    error
	)

	spin.Prefix = "initializing "
	spin.Start()

	if len(strings.TrimSpace(filter)) > 1 {
		idxs = c.snmpGetIdx(filter)
	}

	s1, err = c.snmpGetInterfaces(idxs)
	if err != nil {
		spin.Stop()
		return err
	}
	if len(s1)-1 < 1 {
		spin.Stop()
		return fmt.Errorf("could not find any interface")
	}

	spin.Stop()

	if len(s1) > maxRowTermUI {
		return fmt.Errorf("result can not fit on the screen please try filter")
	}

	ui.DefaultEvtStream = ui.NewEvtStream()
	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()
	w := initWidgets()

	wList := []*ui.List{
		w.ifName,
		w.ifStatus,
		w.ifDescr,
		w.ifTIn,
		w.ifTOut,
		w.ifPIn,
		w.ifPOut,
		w.ifDIn,
		w.ifDOut,
		w.ifEIn,
		w.ifEOut,
	}

	for i, l := range wList {
		l.Items = make([]string, maxRowTermUI+5)
		l.X = 0
		l.Y = 0
		l.Height = len(s1)
		l.Border = false
		l.PaddingLeft = 1

		rows = append(rows, ui.NewCol(span[i], 0, l))
	}

	for i, v := range s1[0] {
		wList[i].Items[0] = fmt.Sprintf("[%s](fg-magenta,fg-bold)", v)
	}

	// initialize cells
	for i, v := range s1[1:] {
		w.ifName.Items[i+1] = v[0]
		w.ifStatus.Items[i+1] = ifStatus(v[1])
		w.ifDescr.Items[i+1] = v[2]
		for _, l := range wList[3:] {
			l.Items[i+1] = "-"
		}
	}

	w.updateFrame(c, "")

	screen := []*ui.Row{
		ui.NewRow(
			ui.NewCol(12, 0, w.header),
		),
		ui.NewRow(
			ui.NewCol(12, 0, w.menu),
		),
		ui.NewRow(rows...),
		ui.NewRow(
			ui.NewCol(12, 0, w.footer),
		),
	}

	ui.Handle("/timer/1s", func(e ui.Event) {
		t := e.Data.(ui.EvtTimer)
		if t.Count%10 != 0 {
			return
		}

		s2, err = c.snmpGetInterfaces(idxs)
		if err != nil {
			w.updateFrame(c, "error: "+err.Error())
			ui.Render(ui.Body)
			return
		} else if strings.Contains(w.header.Text, "error") {
			w.updateFrame(c, "")
		}

		for i := range s2[1:] {
			rows := normalize(s1[i+1], s2[i+1], 10)
			for c := range wList {
				wList[c].Items[i+1] = rows[c]
			}
		}

		copy(s1, s2)
		ui.Render(ui.Body)
	})

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		w.updateFrame(c, "")
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	ui.Body.AddRows(screen...)
	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Loop()
	return nil
}

func resetCounters(s [][]string) {
	for i := range s[1:] {
		for j := range s[i][3:] {
			s[i+1][j+3] = "0.0"
		}
	}
}
