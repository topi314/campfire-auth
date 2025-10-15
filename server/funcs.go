package server

import (
	"html/template"
	"time"
)

var templateFuncs = template.FuncMap{
	"formatDate":             formatDate,
	"formatDateNice":         formatDateNice,
	"formatTimeToDayTime":    formatTimeToDayTime,
	"formatTimeToRelDayTime": formatTimeToRelDayTime,
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatDateNice(t time.Time) string {
	return t.Format("2 Jan 2006")
}

func formatTimeToDayTime(t time.Time) string {
	return t.Format("Mon _2 Jan 2006 15:04 MST")
}

func formatTimeToRelDayTime(t time.Time) string {
	nowYear, nowMonth, nowDay := time.Now().Date()
	year, month, day := t.Date()

	timeStr := t.Format("15:04 MST")

	switch {
	case year == nowYear && month == nowMonth && day == nowDay:
		return "Today at " + timeStr
	case year == nowYear && month == nowMonth && day == nowDay-1:
		return "Yesterday at " + timeStr
	case year == nowYear && month == nowMonth && day == nowDay+1:
		return "Tomorrow at " + timeStr
	default:
		return formatTimeToDayTime(t)
	}
}
