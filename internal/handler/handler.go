package handler

import (
	"fmt"
	"go_final_project/internal/nextdate"
	"log"
	"net/http"
	"time"
)

// NextDateHandler обрабатывает запросы к /api/nextdate
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Проверка на пустые значения
	if nowStr == "" || dateStr == "" || repeat == "" {
		fmt.Println("необходимо указать все параметры: now, date, repeat", http.StatusBadRequest)
		http.Error(w, "необходимо указать все параметры: now, date, repeat", http.StatusBadRequest)
		return
	}

	// Парсинг текущей даты
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		log.Printf("Ошибка парсинга текущей даты: %v", err)
		http.Error(w, "неверный формат текущей даты", http.StatusBadRequest)
		return
	}

	// Вызов функции NextDate
	nextDate, err := nextdate.NextDate(now, dateStr, repeat)
	if err != nil {
		log.Printf("Ошибка в NextDate: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возврат результата
	fmt.Fprint(w, nextDate)
}
