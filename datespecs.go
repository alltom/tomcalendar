package main

import (
	"time"
)

// DailyDateSpec takes place every day.
type DailyDateSpec struct{}

// UnionDateSpec takes place on days where any of Specs take place.
type UnionDateSpec struct {
	Specs []DateSpec
}

// EveryNthDayDateSpec takes place every Count days starting on January 1, 1970.
type EveryNthDayDateSpec struct {
	Count int
}

// WeekdayDateSpec takes place every Weekday.
type WeekdayDateSpec struct {
	Weekday time.Weekday
}

// EveryNthWeekdayDateSpec takes place on Weekday every Count weeks, starting the week of January 1, 1970.
type EveryNthWeekdayDateSpec struct {
	Weekday time.Weekday
	Count   int
}

// DayOfMonthDateSpec takes place on the Day'th day of the month when Day is positive, or Day days from the end of the month if it's negative (e.g. in January, -1 would take place on January 31)
type DayOfMonthDateSpec struct {
	Day int
}

// YearlyDateSpec takes place on the given day every year.
type YearlyDateSpec struct {
	Month time.Month
	Day   int
}

// SingleDayDateSpec takes place once, on the given date.
type SingleDayDateSpec struct {
	Year  int
	Month time.Month
	Day   int
}

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (s *UnionDateSpec) OccursOn(date *Date) bool {
	for _, spec := range s.Specs {
		if spec.OccursOn(date) {
			return true
		}
	}
	return false
}

func (s *DailyDateSpec) OccursOn(date *Date) bool {
	return true
}

func (s *EveryNthDayDateSpec) OccursOn(date *Date) bool {
	return daysSinceEpoch(date)%s.Count == 0
}

func (s *WeekdayDateSpec) OccursOn(date *Date) bool {
	return date.LocalStartOfDay().Weekday() == s.Weekday
}

func (s *EveryNthWeekdayDateSpec) OccursOn(date *Date) bool {
	return date.LocalStartOfDay().Weekday() == s.Weekday && (daysSinceEpoch(date)/7)%s.Count == 0
}

func (s *DayOfMonthDateSpec) OccursOn(date *Date) bool {
	if s.Day < 0 {
		return time.Date(date.Year, date.Month+1, s.Day+1, 0, 0, 0, 0, time.Local).Day() == date.Day
	}
	return s.Day == date.Day
}

func (s *YearlyDateSpec) OccursOn(date *Date) bool {
	return s.Month == date.Month && s.Day == date.Day
}

func (s *SingleDayDateSpec) OccursOn(date *Date) bool {
	return s.Year == date.Year && s.Month == date.Month && s.Day == date.Day
}

func (d *Date) LocalStartOfDay() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.Local)
}

func daysSinceEpoch(date *Date) int {
	return daysBetween(date.LocalStartOfDay(), time.Date(1970, time.January, 1, 0, 0, 0, 0, time.Local))
}

// https://dev.to/samwho/get-the-number-of-days-between-two-dates-in-go-5bf3
func daysBetween(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}

	days := -a.YearDay()
	for year := a.Year(); year < b.Year(); year++ {
		days += time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC).YearDay()
	}
	days += b.YearDay()

	return days
}
