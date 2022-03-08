package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/alltom/tomcalendar/pkg/datespec"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Title    string
	DateSpec datespec.DateSpec
}

func (s *Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"text": s.Title,
		"spec": s.DateSpec,
	})
}

var (
	commentPattern         = regexp.MustCompile(`^\s*(#|//).*$`)
	dayOfMonthPattern      = regexp.MustCompile(`^(-?[0-9]+) \*$`)
	yearlyPattern          = regexp.MustCompile(`^([^ ]+) ([0-9]+)$`)
	singleDayPattern       = regexp.MustCompile(`^([^ ]+) ([0-9]+), ([0-9]+)$`)
	everyNthDayPattern     = regexp.MustCompile(`^\*/([0-9]+)$`)
	everyNthWeekdayPattern = regexp.MustCompile(`^(Sunday|Monday|Tuesday|Wednesday|Thursday|Friday|Saturday)/([0-9]+)$`)
)

func Parse(r io.Reader, path string) ([]*Entry, error) {
	var entries []*Entry
	var lastDateSpec datespec.DateSpec
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

func ParseDateSpec(s, path string, lineNumber int) (datespec.DateSpec, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "|") {
		var specs []datespec.DateSpec
		for _, pattern := range strings.Split(s, "|") {
			spec, err := ParseDateSpec(pattern, path, lineNumber)
			if err != nil {
				return nil, err
			}
			specs = append(specs, spec)
		}
		return &datespec.UnionDateSpec{specs}, nil
	} else if s == "*" {
		return &datespec.DailyDateSpec{}, nil
	} else if m := everyNthDayPattern.FindStringSubmatch(s); m != nil {
		count, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse interval %q: %v", path, lineNumber, m[1], err)
		}
		if count <= 0 {
			return nil, fmt.Errorf("%s:%d: day count is %q, but must be at least 1", path, lineNumber, m[1])
		}
		return &datespec.EveryNthDayDateSpec{count}, nil
	} else if weekday := getWeekday(s); weekday != nil {
		return &datespec.WeekdayDateSpec{*weekday}, nil
	} else if m := everyNthWeekdayPattern.FindStringSubmatch(s); m != nil {
		weekday := getWeekday(m[1])
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
		return &datespec.EveryNthWeekdayDateSpec{*weekday, count}, nil
	} else if m := dayOfMonthPattern.FindStringSubmatch(s); m != nil {
		day, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse day of month %q: %v", path, lineNumber, m[1], err)
		}
		if day < -31 || day > 31 || day == 0 {
			return nil, fmt.Errorf("%s:%d: day of month %d is out of range [-31, -1] ∪ [1, 31]", path, lineNumber, day)
		}
		return &datespec.DayOfMonthDateSpec{day}, nil
	} else if m := yearlyPattern.FindStringSubmatch(s); m != nil {
		month := getMonth(m[1])
		if month == nil {
			return nil, fmt.Errorf("%s:%d: unrecognized month %q", path, lineNumber, m[1])
		}
		day, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: could not parse day of month %q: %v", path, lineNumber, m[2], err)
		}
		return &datespec.YearlyDateSpec{*month, day}, nil
	} else if m := singleDayPattern.FindStringSubmatch(s); m != nil {
		month := getMonth(m[1])
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
		return &datespec.SingleDayDateSpec{year, *month, day}, nil
	}
	return nil, fmt.Errorf("%s:%d: unrecognized date specifier: %q", path, lineNumber, s)
}

func getWeekday(s string) *time.Weekday {
	s = strings.ToLower(s)
	for wd := time.Sunday; wd <= time.Saturday; wd++ {
		if strings.ToLower(wd.String()) == s {
			return &wd
		}
	}
	return nil
}

func getMonth(s string) *time.Month {
	s = strings.ToLower(s)
	for m := time.January; m <= time.December; m++ {
		if strings.ToLower(m.String()) == s {
			return &m
		}
	}
	return nil
}
