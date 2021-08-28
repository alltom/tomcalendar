package main

import (
	"alltom/tomcalendar/pkg/datespec"
	"alltom/tomcalendar/pkg/parser"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	calendarPath = flag.String("calendar", "", "Path to the calendar file to read (uses stdin if empty)")
	queryDate    = flag.String("date", "", `Date whose agenda to return (e.g. "2020-03-22") (the default is the current date)`)
	sinceDate    = flag.String("since", "", "Print the agendas for every day AFTER this date up to the current date (ignored if -date is also present)")
)

func main() {
	flag.Parse()

	var r io.Reader
	var path string
	if *calendarPath == "" {
		r = os.Stdin
	} else {
		f, err := os.Open(*calendarPath)
		if err != nil {
			log.Fatalf("could not open calendar: %v", err)
		}
		r = f
		path = *calendarPath
	}

	var dates []*datespec.Date
	if *queryDate != "" {
		t, err := time.Parse("2006-01-02", *queryDate)
		if err != nil {
			log.Fatalf("could not parse date: %v", err)
		}
		dates = []*datespec.Date{&datespec.Date{t.Year(), t.Month(), t.Day()}}
	} else if *sinceDate != "" {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		t, err := time.Parse("2006-01-02", *sinceDate)
		if err != nil {
			log.Fatalf("could not parse date: %v", err)
		}
		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, now.Location())

		for !startOfDay.Before(t) {
			dates = append(dates, &datespec.Date{t.Year(), t.Month(), t.Day()})
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, now.Location())
		}
	} else {
		t := time.Now()
		dates = []*datespec.Date{&datespec.Date{t.Year(), t.Month(), t.Day()}}
	}

	entries, err := parser.Parse(r, path)
	if err != nil {
		log.Fatalf("could not parse calendar: %v", err)
	}
	for _, entry := range entries {
		for _, date := range dates {
			if entry.DateSpec.OccursOn(date) {
				fmt.Printf("%s\n", entry.Title)
				break
			}
		}
	}
}
