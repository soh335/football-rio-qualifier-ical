package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/soh335/ical"
	"github.com/soh335/icalparser"
)

type progADay struct {
	Date  string `json:"date"`
	Progs []prog `json:"progs"`
}

type prog struct {
	TimeS    string `json:"time_s"`
	TimeSUnd int    `json:"time_s_und"`
	TimeE    string `json:"time_e"`
	TimeEUnd int    `json:"time_e_und"`
	Nodisp   int    `json:"nodisp"`
	Media    string `json:"media"`
	Ttl      string `json:"ttl"`
	Form     string `json:"form"`
	Saiho    string `json:"saiho"`
	Shutsuen string `json:"shutsuen"`
	Biko     string `json:"biko"`
	MatchID  string `json:"matchID"`
}

const (
	timeLayout = "2006-01-02 15:04:05"
)

var (
	loc *time.Location

	tzid    = flag.String("tzid", "Asia/Tokyo", "tzid")
	calname = flag.String("calname", "サッカー男子 リオ五輪アジア最終予選", "calname")
)

func main() {
	flag.Parse()
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	{
		var err error
		loc, err = time.LoadLocation(*tzid)
		if err != nil {
			return err
		}
	}
	var progDays []progADay
	if err := json.NewDecoder(os.Stdin).Decode(&progDays); err != nil {
		return err
	}

	components := []ical.VComponent{}
	for _, progADay := range progDays {
		for _, prog := range progADay.Progs {
			if prog.Form != "live" {
				continue
			}
			component, err := prog.event()
			if err != nil {
				return err
			}
			components = append(components, component)
		}
	}

	cal := ical.NewBasicVCalendar()
	cal.PRODID = *calname
	cal.X_WR_CALNAME = *calname
	cal.X_WR_CALDESC = *calname
	cal.X_WR_TIMEZONE = loc.String()
	cal.VComponent = components

	var b bytes.Buffer

	if err := cal.Encode(&b); err != nil {
		return err
	}

	o, err := icalparser.NewParser(&b).Parse()
	if err != nil {
		return err
	}

	if _, err := icalparser.NewPrinter(o).WriteTo(os.Stdout); err != nil {
		return err
	}

	return nil
}

func (p *prog) event() (*ical.VEvent, error) {
	start, err := time.ParseInLocation(timeLayout, p.TimeS, loc)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation(timeLayout, p.TimeE, loc)
	if err != nil {
		return nil, err
	}
	uid, err := p.uid()
	if err != nil {
		return nil, err
	}
	component := &ical.VEvent{
		UID:         uid,
		DTSTAMP:     start,
		DTSTART:     start,
		DTEND:       end,
		SUMMARY:     p.Ttl,
		DESCRIPTION: p.Shutsuen,
		TZID:        loc.String(),
	}
	return component, nil
}

func (p *prog) uid() (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(p); err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write(b.Bytes())
	return hex.EncodeToString(h.Sum(nil)), nil
}
