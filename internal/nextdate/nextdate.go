package nextdate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// everyDay обрабатывает правило d
func everyDay(now, date time.Time, daysStr string) (string, error) {
	d, err := strconv.Atoi(daysStr)
	if err != nil || d > 400 || d <= 0 {
		return "", fmt.Errorf("неверное правило повторения в d")
	}

	// Прибавляем день
	date = date.AddDate(0, 0, d)

	// Проверяем, если дата всё ещё не в будущем
	for !date.After(now) {
		date = date.AddDate(0, 0, d)
	}

	// Корректировка 29 февраля
	if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()) {
		date = time.Date(date.Year(), time.March, 1, 0, 0, 0, 0, time.Local)
	}

	return date.Format("20060102"), nil
}

// everyWeek обрабатывает правило w
func everyWeek(now, date time.Time, daysStr string) (string, error) {
	days := strings.Split(daysStr, ",")
	validDays := make(map[int]bool)

	// Преобразуем значения дней недели в числа и проверяем корректность
	for _, day := range days {
		d, err := strconv.Atoi(day)
		if err != nil || d < 1 || d > 7 {
			return "", fmt.Errorf("неверный день недели: %s", day)
		}
		validDays[d] = true
	}

	// Начинаем проверку с завтрашнего дня
	date = date.AddDate(0, 0, 1) // Проверяем начиная с завтрашнего дня
	for {
		weekDay := int(date.Weekday())
		if weekDay == 0 {
			weekDay = 7 // Обработка воскресенья (0) к 7
		}

		if validDays[weekDay] && date.After(now) {
			return date.Format("20060102"), nil
		}
		date = date.AddDate(0, 0, 1)
	}
}

// everyMonth обрабатывает правило m
func everyMonth(now, date time.Time, daysStr string) (string, error) {
	days := strings.Split(daysStr, ",")
	validDays := []int{}

	// Преобразуем значения в числа и проверяем их корректность
	for _, day := range days {
		d, err := strconv.Atoi(day)
		if err != nil || d < -31 || d > 31 || d == 0 {
			return "", fmt.Errorf("неверный день месяца: %s", day)
		}
		validDays = append(validDays, d)
	}

	for {
		// Определяем последний день месяца
		for _, day := range validDays {
			newDate := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
			lastDay := newDate.AddDate(0, 1, -1).Day()

			if day < 0 {
				day = lastDay + day + 1
			}

			if day > lastDay {
				day = lastDay
			}

			newDate = time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, time.Local)
			if newDate.After(now) {
				return newDate.Format("20060102"), nil
			}
		}
		date = date.AddDate(0, 1, 0) // Переход к следующему месяцу
	}
}

// everyYear обрабатывает правило y
func everyYear(now, date time.Time, _ string) (string, error) {
	// Добавляем ровно 1 год
	date = date.AddDate(1, 0, 0)

	// Корректируем 29 февраля в невисокосные годы
	if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()) {
		date = time.Date(date.Year(), time.March, 1, 0, 0, 0, 0, time.Local)
	}

	// Если новая дата всё ещё не после now, добавляем ещё 1 год
	for !date.After(now) {
		date = date.AddDate(1, 0, 0)

		// Коррекция 29 февраля снова
		if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()) {
			date = time.Date(date.Year(), time.March, 1, 0, 0, 0, 0, time.Local)
		}
	}

	return date.Format("20060102"), nil
}

// isLeapYear проверяет, является ли год високосным
func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

// NextDate вычисляет следующую дату выполнения задачи
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	// Парсим входную дату
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}

	if repeat == "" {
		return "", errors.New("отсутствует правило повторения")
	}

	// Определяем тип правила повторения и вызываем соответствующую функцию
	switch {
	case repeat == "y":
		return everyYear(now, date, "")
	case strings.HasPrefix(repeat, "d "):
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила повторения")
		}
		return everyDay(now, date, parts[1])
	case strings.HasPrefix(repeat, "w "):
		parts := strings.SplitN(repeat, " ", 2)
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила повторения w")
		}
		return everyWeek(now, date, parts[1])
	case strings.HasPrefix(repeat, "m "):
		parts := strings.SplitN(repeat, " ", 2)
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила повторения m")
		}
		return everyMonth(now, date, parts[1])
	default:
		return "", errors.New("неподдерживаемый формат правила повторения")
	}
}
