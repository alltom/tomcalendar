package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alltom/tomcalendar/pkg/parser"
	"io"
	"log"
	"os"
)

type ArrayFlag []string

func (f *ArrayFlag) String() string {
	return "my string representation"
}

func (f *ArrayFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	calendarPaths ArrayFlag
	jsonPath      = flag.String("json", "", "Path to write the JSON version of the calendar to (uses stdout if empty)")
)

func main() {
	flag.Var(&calendarPaths, "calendar", "Path to a calendar file to read (uses stdin if empty)")
	flag.Parse()

	var path string
	var entries []*parser.Entry
	if len(calendarPaths) == 0 {
		var err error
		entries, err = parser.Parse(os.Stdin, path)
		if err != nil {
			log.Fatalf("could not parse calendar: %v", err)
		}
	} else {
		for _, calendarPath := range calendarPaths {
			fentries, err := readEntries(calendarPath)
			if err != nil {
				log.Fatalf("could not read calendar %q: %v", calendarPath, err)
			}
			entries = append(entries, fentries...)
		}
	}

	var w io.Writer
	if *jsonPath == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(*jsonPath)
		if err != nil {
			log.Fatalf("could not open JSON file %q: %v", *jsonPath, err)
		}
		w = f
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(map[string]interface{}{"entries": entries}); err != nil {
		log.Fatalf("could not encode project: %v", err)
	}
}

func readEntries(calendarPath string) ([]*parser.Entry, error) {
	f, err := os.Open(calendarPath)
	if err != nil {
		return nil, fmt.Errorf("open calendar: %v", err)
	}
	defer f.Close()

	entries, err := parser.Parse(f, calendarPath)
	if err != nil {
		return nil, fmt.Errorf("parse calendar: %v", err)
	}
	return entries, nil
}
