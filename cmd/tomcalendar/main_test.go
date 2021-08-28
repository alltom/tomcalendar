package main

import (
	"strings"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	specs := `
*	Daily
	Daily 2

*/3	Every third day
	Another every third day

// Weekday tests
Sunday	Only Sunday
	Also only Sunday
Friday	Only Friday
Friday/2	Every other Friday

// Union tests
Sunday | Friday	Sunday or Friday

# Day of month tests
1 *	First of the month
15 *	15th of the month
	Ides #romans
-1 *	Last day of the month
-2 *	Second-to-last day of the month

// Yearly tests
December 25	Santa's birthday
	Christmas
March 29
	A random date

# One-day tests
December 24, 1980
	A specific Christmas Eve
	http://example.com/
`
	entries, err := Parse(strings.NewReader(specs), "/path/to/calendar")
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}

	var tests = []struct {
		date     string
		expected []string
	}{
		{"Thu Jan 1, 1970", []string{"First of the month", "Every third day", "Another every third day"}},
		{"Fri Jan 2, 1970", []string{"Only Friday", "Every other Friday", "Sunday or Friday"}},
		{"Sat Jan 3, 1970", []string{}},
		{"Sun Jan 4, 1970", []string{"Every third day", "Another every third day", "Only Sunday", "Also only Sunday", "Sunday or Friday"}},
		{"Fri Jan 9, 1970", []string{"Only Friday", "Sunday or Friday"}},

		{"Sun Mar 1, 2020", []string{"Only Sunday", "Also only Sunday", "First of the month", "Sunday or Friday"}},
		{"Sun Mar 15, 2020", []string{"Only Sunday", "Also only Sunday", "15th of the month", "Ides #romans", "Every third day", "Another every third day", "Sunday or Friday"}},

		{"Sun Mar 29, 2020", []string{"Only Sunday", "Also only Sunday", "A random date", "Sunday or Friday"}},
		{"Fri Dec 25, 2020", []string{"Only Friday", "Every other Friday", "Christmas", "Santa's birthday", "Every third day", "Another every third day", "Sunday or Friday"}},

		{"Tue Dec 24, 1980", []string{"A specific Christmas Eve", "http://example.com/"}},
		{"Wed Dec 30, 2020", []string{"Second-to-last day of the month"}},
		{"Thu Dec 31, 2020", []string{"Every third day", "Another every third day", "Last day of the month"}},
	}
	for _, tt := range tests {
		tt.expected = append(tt.expected, "Daily", "Daily 2")

		t.Run(tt.date, func(t *testing.T) {
			d := ParseDate(tt.date)

			var entryTitles []string
			for _, entry := range entries {
				if entry.DateSpec.OccursOn(d) {
					entryTitles = append(entryTitles, entry.Title)
				}
			}

			for _, expected := range tt.expected {
				if !Contains(entryTitles, expected) {
					t.Errorf("expected %q, but is not present in %q", expected, entryTitles)
				}
			}
			for _, title := range entryTitles {
				if !Contains(tt.expected, title) {
					t.Errorf("did not expect %q", title)
				}
			}
		})
	}
}

func ParseDate(s string) *Date {
	t, err := time.Parse("Mon Jan 2, 2006", s)
	if err != nil {
		panic("invalid date")
	}
	return &Date{t.Year(), t.Month(), t.Day()}
}

func Contains(ss []string, s string) bool {
	for _, s2 := range ss {
		if s2 == s {
			return true
		}
	}
	return false
}
