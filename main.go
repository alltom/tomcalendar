package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	calendarPath = flag.String("calendar", "", "Path to the calendar file to read (uses stdin if empty)")
	queryDate    = flag.String("date", "", `Date whose agenda to return (e.g. "2020-03-22") (the default is the current date)`)
	sinceDate    = flag.String("since", "", "Print the agendas for every day AFTER this date up to the current date (ignored if -date is also present)")
)

var (
	commentPattern         = regexp.MustCompile(`^\s*(#|//).*$`)
	dayOfMonthPattern      = regexp.MustCompile(`^(-?[0-9]+) \*$`)
	yearlyPattern          = regexp.MustCompile(`^([^ ]+) ([0-9]+)$`)
	singleDayPattern       = regexp.MustCompile(`^([^ ]+) ([0-9]+), ([0-9]+)$`)
	everyNthDayPattern     = regexp.MustCompile(`^\*/([0-9]+)$`)
	everyNthWeekdayPattern = regexp.MustCompile(`^(Sunday|Monday|Tuesday|Wednesday|Thursday|Friday|Saturday)/([0-9]+)$`)
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

	var dates []*Date
	if *queryDate != "" {
		t, err := time.Parse("2006-01-02", *queryDate)
		if err != nil {
			log.Fatalf("could not parse date: %v", err)
		}
		dates = []*Date{&Date{t.Year(), t.Month(), t.Day()}}
	} else if *sinceDate != "" {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		t, err := time.Parse("2006-01-02", *sinceDate)
		if err != nil {
			log.Fatalf("could not parse date: %v", err)
		}
		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, now.Location())

		for !startOfDay.Before(t) {
			dates = append(dates, &Date{t.Year(), t.Month(), t.Day()})
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, now.Location())
		}
	} else {
		t := time.Now()
		dates = []*Date{&Date{t.Year(), t.Month(), t.Day()}}
	}

	entries, err := Parse(r, path)
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

func Parse(r io.Reader, path string) ([]*Entry, error) {
	var entries []*Entry
	var lastDateSpec DateSpec
	var lastDateSpecText string
	var lastDateSpecLine int
	var foundEntries bool
	scanner := bufio.NewScanner(r)
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		line = commentPattern.ReplaceAllString(line, "")
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) == 1 {
			dateSpec, err := ParseDateSpec(fields[0], path, i)
			if err != nil {
				return nil, err
			}
			lastDateSpec = dateSpec
			lastDateSpecText = fields[0]
			lastDateSpecLine = i
			foundEntries = false
		} else if len(fields) == 2 {
			if fields[0] == "" {
				if lastDateSpec == nil {
					return nil, fmt.Errorf(`%s:%d: date spec required for first entry`, path, i)
				}
				entries = append(entries, &Entry{fields[1], lastDateSpec})
				foundEntries = true
			} else {
				dateSpec, err := ParseDateSpec(fields[0], path, i)
				if err != nil {
					return nil, err
				}
				lastDateSpec = dateSpec
				lastDateSpecText = fields[0]
				lastDateSpecLine = i
				foundEntries = true
				entries = append(entries, &Entry{fields[1], dateSpec})
			}
		} else {
			return nil, fmt.Errorf("%s:%d: no tab found in %q", path, i, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner: %v", err)
	}

	if lastDateSpec != nil && !foundEntries {
		return nil, fmt.Errorf("%s:%d: date spec %q has no entries", path, lastDateSpecLine, lastDateSpecText)
	}

	return entries, nil
}

func ParseDateSpec(s, path string, lineNumber int) (DateSpec, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "|") {
		var specs []DateSpec
		for _, pattern := range strings.Split(s, "|") {
			spec, err := ParseDateSpec(pattern, path, lineNumber)
			if err != nil {
				return nil, err
			}
			specs = append(specs, spec)
		}
		return &UnionDateSpec{specs}, nil
	} else if s == "*" {
		return &DailyDateSpec{}, nil
	} else if m := everyNthDayPattern.FindStringSubmatch(s); m != nil {
		count, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse interval %q: %v", path, lineNumber, m[1], err)
		}
		if count <= 0 {
			return nil, fmt.Errorf("%s:%d: day count is %q, but must be at least 1", path, lineNumber, m[1])
		}
		return &EveryNthDayDateSpec{count}, nil
	} else if weekday := GetWeekday(s); weekday != nil {
		return &WeekdayDateSpec{*weekday}, nil
	} else if m := everyNthWeekdayPattern.FindStringSubmatch(s); m != nil {
		weekday := GetWeekday(m[1])
		if weekday == nil {
			return nil, fmt.Errorf("%s:%d: could not parse weekday %q", path, lineNumber, m[1])
		}
		count, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse interval %q: %v", path, lineNumber, m[1], err)
		}
		if count <= 0 {
			return nil, fmt.Errorf("%s:%d: week count is %q, but must be at least 1", path, lineNumber, m[2])
		}
		return &EveryNthWeekdayDateSpec{*weekday, count}, nil
	} else if m := dayOfMonthPattern.FindStringSubmatch(s); m != nil {
		day, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse day of month %q: %v", path, lineNumber, m[1], err)
		}
		if day < -31 || day > 31 || day == 0 {
			return nil, fmt.Errorf("%s:%d: day of month %d is out of range [-31, -1] ∪ [1, 31]", path, lineNumber, day)
		}
		return &DayOfMonthDateSpec{day}, nil
	} else if m := yearlyPattern.FindStringSubmatch(s); m != nil {
		month := GetMonth(m[1])
		if month == nil {
			return nil, fmt.Errorf("%s:%d: unrecognized month %q", path, lineNumber, m[1])
		}
		day, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse day of month %q: %v", path, lineNumber, m[2], err)
		}
		return &YearlyDateSpec{*month, day}, nil
	} else if m := singleDayPattern.FindStringSubmatch(s); m != nil {
		month := GetMonth(m[1])
		if month == nil {
			return nil, fmt.Errorf("%s:%d: unrecognized month %q", path, lineNumber, m[1])
		}
		day, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse day of month %q: %v", path, lineNumber, m[2], err)
		}
		if day < 1 || day > 31 {
			return nil, fmt.Errorf("%s:%d: day of month %d is out of range [1, 31]", path, lineNumber, day)
		}
		year, err := strconv.Atoi(m[3])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse year %q: %v", path, lineNumber, m[3], err)
		}
		return &SingleDayDateSpec{year, *month, day}, nil
	}
	return nil, fmt.Errorf("%s:%d: unrecognized date specifier: %q", path, lineNumber, s)
}

type Entry struct {
	Title    string
	DateSpec DateSpec
}

type DateSpec interface {
	OccursOn(*Date) bool
}

func GetWeekday(s string) *time.Weekday {
	s = strings.ToLower(s)
	for wd := time.Sunday; wd <= time.Saturday; wd++ {
		if strings.ToLower(wd.String()) == s {
			return &wd
		}
	}
	return nil
}

func GetMonth(s string) *time.Month {
	s = strings.ToLower(s)
	for m := time.January; m <= time.December; m++ {
		if strings.ToLower(m.String()) == s {
			return &m
		}
	}
	return nil
}
