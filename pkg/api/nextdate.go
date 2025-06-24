package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	TimeFormat    = "20060102"
	maxDailyDays  = 400
	sundayWeekday = 7
	daysInWeek    = 7
	maxMonthDay   = 31
	minMonthDay   = -2
)

var (
	ErrRepeatRequired   = errors.New("repeat is required")
	ErrInvalidTags      = errors.New("invalid tags")
	ErrMissingDaysValue = errors.New("missing days value")
	ErrInvalidWeekday   = errors.New("invalid weekday")
	ErrInvalidDay       = errors.New("invalid day")
)

// nextDayHandler возвращает следующую дату для задачи
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if now == "" {
		now = time.Now().Format(TimeFormat)
	}

	nowTime, err := time.Parse(TimeFormat, now)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, nextDate)
}

// NextDate вычисляет следующую дату задачи
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrRepeatRequired
	}

	date, err := time.Parse(TimeFormat, dateStr)
	if err != nil {
		return "", err
	}

	repParts := strings.Split(repeat, " ")
	if len(repParts) == 0 {
		return "", ErrInvalidTags
	}

	switch repParts[0] {
	case "d":
		return handleDaily(now, date, repParts)
	case "y":
		return handleYearly(now, date)
	case "w":
		return handleWeekly(now, date, repParts)
	case "m":
		return handleMonthly(now, date, repParts)
	default:
		return "", ErrInvalidTags
	}
}

func writeErrorResponse(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}

func writeJSONResponse(w http.ResponseWriter, status int, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(data))
}

func handleDaily(now, date time.Time, repParts []string) (string, error) {
	if len(repParts) < 2 {
		return "", ErrMissingDaysValue
	}

	days, err := strconv.Atoi(repParts[1])
	if err != nil {
		return "", err
	}

	if days < 1 {
		return "", errors.New("days must be greater than 0")
	}
	if days > maxDailyDays {
		return "", errors.New("days must be less than 400 days")
	}

	date = calculateNextDate(now, date, func(d time.Time) time.Time {
		return d.AddDate(0, 0, days)
	})

	return date.Format(TimeFormat), nil
}

func handleYearly(now, date time.Time) (string, error) {
	date = calculateNextDate(now, date, func(d time.Time) time.Time {
		return d.AddDate(1, 0, 0)
	})

	return date.Format(TimeFormat), nil
}

func handleWeekly(now, date time.Time, repParts []string) (string, error) {
	if len(repParts) < 2 {
		return "", ErrMissingDaysValue
	}

	weekdays, err := parseWeekdays(repParts[1])
	if err != nil {
		return "", err
	}

	for {
		currentWeekday := int(date.Weekday())
		if date.After(now) && contains(weekdays, currentWeekday) {
			break
		}
		date = date.AddDate(0, 0, 1)
	}

	return date.Format(TimeFormat), nil
}

func handleMonthly(now, date time.Time, repParts []string) (string, error) {
	if len(repParts) < 2 {
		return "", ErrMissingDaysValue
	}

	days, err := parseIntList(repParts[1])
	if err != nil {
		return "", err
	}

	if !validateMonthDays(days) {
		return "", ErrInvalidDay
	}

	var months []int
	if len(repParts) >= 3 {
		months, err = parseIntList(repParts[2])
		if err != nil {
			return "", err
		}
	}

	for {
		if date.After(now) && isValidMonthDay(date, days, months) {
			break
		}
		date = date.AddDate(0, 0, 1)
	}

	return date.Format(TimeFormat), nil
}

func calculateNextDate(now, date time.Time, nextFunc func(time.Time) time.Time) time.Time {
	for {
		date = nextFunc(date)
		if date.After(now) {
			break
		}
	}
	return date
}

func parseWeekdays(s string) ([]int, error) {
	days, err := parseIntList(s)
	if err != nil {
		return nil, err
	}

	for _, day := range days {
		if day < 1 || day > daysInWeek {
			return nil, ErrInvalidWeekday
		}
	}

	// Конвертируем воскресенье (7) в 0 для time.Weekday
	adjusted := make([]int, 0, len(days))
	for _, day := range days {
		if day == sundayWeekday {
			adjusted = append(adjusted, 0)
		} else {
			adjusted = append(adjusted, day)
		}
	}

	return adjusted, nil
}

func parseIntList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		result = append(result, num)
	}

	return result, nil
}

func validateMonthDays(days []int) bool {
	for _, day := range days {
		if day > maxMonthDay || day == 0 || day < minMonthDay {
			return false
		}
	}
	return true
}

func isValidMonthDay(date time.Time, days, months []int) bool {
	currentMonth := int(date.Month())
	currentDay := date.Day()

	// Проверяем месяц, если указаны конкретные месяцы
	if len(months) > 0 && !contains(months, currentMonth) {
		return false
	}

	// Проверяем день
	for _, day := range days {
		if day > 0 {
			if day == currentDay {
				return true
			}
		} else {
			// Обработка отрицательных дней (с конца месяца)
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			actualDay := lastDay + day + 1
			if actualDay == currentDay {
				return true
			}
		}
	}

	return false
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
