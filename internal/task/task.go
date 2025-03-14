package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/internal/nextdate"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Cтруктура запроса на добавление или обновление задачи
type TaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Cтруктура для ответа с ID созданной задачи или ошибкой
type TaskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// Обработка на добавление задачи
func AddTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Метод не поддерживается: %s", r.Method)
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		var req TaskRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("Ошибка десериализации JSON", w, http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			sendErrorResponse(w, "Не указан заголовок задачи")
			return
		}

		now := time.Now()
		dateStr := req.Date
		if dateStr == "" {
			dateStr = now.Format("20060102")
		}

		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			sendErrorResponse(w, "Неверный формат даты")
			return
		}

		// Если дата в прошлом и есть повторение, ищем следующую дату
		if date.Before(now) && req.Repeat != "" {
			nextDate, err := nextdate.NextDate(now, dateStr, req.Repeat)
			if err != nil {
				sendErrorResponse(w, fmt.Sprintf("Ошибка в правиле повторения: %v", err))
				return
			}
			dateStr = nextDate
		}

		if date.Before(now) {
			dateStr = now.Format("20060102")
		}

		// Вставляем задачу в базу данных
		query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
		res, err := db.Exec(query, dateStr, req.Title, req.Comment, req.Repeat)
		if err != nil {
			sendErrorResponse(w, fmt.Sprintf("Ошибка при добавлении задачи: %v", err))
			return
		}

		// Получаем ID созданной задачи
		id, err := res.LastInsertId()
		if err != nil {
			sendErrorResponse(w, fmt.Sprintf("Ошибка при получении ID задачи: %v", err))
			return
		}

		sendSuccessResponse(w, id)
	}
}

// Обработчик получения задачи по ID
func GetTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			log.Println("Ошибка: не указан идентификатор задачи")
			http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		log.Printf("Получен запрос на задачу с ID: %s", id)

		var task Task
		err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).
			Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

		if err == sql.ErrNoRows {
			log.Printf("Задача с id=%s не найдена в базе", id)
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Ошибка при выполнении SQL-запроса: %v", err)
			http.Error(w, `{"error": "Ошибка сервера"}`, http.StatusInternalServerError)
			return
		}

		log.Printf("Задача найдена: %+v", task)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

// Обработчик обновления задачи
func UpdateTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TaskRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			sendErrorResponseErr(w, "Ошибка десериализации JSON", err)
			return
		}

		if req.ID == "" {
			sendErrorResponseErr(w, "Не указан идентификатор задачи", err)
			return
		}

		if req.Title == "" {
			sendErrorResponseErr(w, "Не указан заголовок задачи", err)
			return
		}

		now := time.Now()
		dateStr := req.Date
		if dateStr == "" {
			dateStr = now.Format("20060102")
		}

		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			sendErrorResponseErr(w, "Неверный формат даты", err)
			return
		}

		if date.Before(now) && req.Repeat != "" {
			nextDate, err := nextdate.NextDate(now, dateStr, req.Repeat)
			if err != nil {
				sendErrorResponseErr(w, "Ошибка в правиле повторения", err)
				return
			}
			dateStr = nextDate
		}

		if date.Before(now) {
			dateStr = now.Format("20060102")
		}

		query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
		res, err := db.Exec(query, dateStr, req.Title, req.Comment, req.Repeat, req.ID)
		if err != nil {
			sendErrorResponseErr(w, "Ошибка при обновлении задачи", err)
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {

			sendErrorResponseErr(w, "Ошибка при получении информации об обновлении", err)
			return
		}

		if rowsAffected == 0 {
			sendErrorResponseErr(w, "Задача не найдена", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

// Обработчик удаления задачи по ID
func DeleteTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			sendErrorResponse(w, "Не указан идентификатор задачи")
			return
		}

		query := `DELETE FROM scheduler WHERE id = ?`
		res, err := db.Exec(query, id)
		if err != nil {
			sendErrorResponseErr(w, "Ошибка при удалении задачи: %v", err)
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			sendErrorResponseErr(w, "Ошибка при получении информации об удалении", err)
			return
		}

		if rowsAffected == 0 {
			sendErrorResponseErr(w, "Задача не найдена", err)
			return
		}

		// Успешный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

// Обработчик завершения задачи
func PostTaskDoneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			sendErrorResponse(w, "Не указан идентификатор задачи")
			return
		}

		var task Task
		err := db.QueryRow(`SELECT id, date, repeat FROM scheduler WHERE id = ?`, id).
			Scan(&task.ID, &task.Date, &task.Repeat)
		if err == sql.ErrNoRows {
			sendErrorResponseErr(w, "Задача не найдена", err)
			return
		} else if err != nil {
			sendErrorResponseErr(w, "Ошибка при запросе задачи", err)
			return
		}

		// Если повторений нет, удаляем задачу
		if task.Repeat == "" {
			_, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
			if err != nil {
				sendErrorResponseErr(w, "Ошибка при удалении задачи", err)
				return
			}
		} else {
			// Рассчитываем следующую дату выполнения
			now := time.Now()
			nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				sendErrorResponseErr(w, "Ошибка в правиле повторения", err)
				return
			}

			// Обновляем дату в базе
			_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, id)
			if err != nil {
				sendErrorResponseErr(w, "Ошибка при обновлении даты", err)
				return
			}
		}

		// Успешный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

// Обработчик удаления задачи по ID
func DeleteTaskDoneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			sendErrorResponse(w, "Не указан идентификатор задачи")
			return
		}

		query := `DELETE FROM scheduler WHERE id = ?`
		res, err := db.Exec(query, id)
		if err != nil {
			sendErrorResponseErr(w, "Ошибка при удалении задачи", err)
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			sendErrorResponseErr(w, "Ошибка при получении информации об удалении: %v", err)
			return
		}

		if rowsAffected == 0 {
			sendErrorResponse(w, "Задача не найдена")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

// sendErrorResponse отправляет JSON-ответ с ошибкой
func sendErrorResponse(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(TaskResponse{Error: errorMsg})
}

func sendErrorResponseErr(w http.ResponseWriter, msg string, err error) {
	log.Printf("%s: %v", msg, err)
	sendErrorResponse(w, fmt.Sprintf("%s: %v", msg, err))
}

// sendSuccessResponse отправляет JSON-ответ с ID созданной задачи
func sendSuccessResponse(w http.ResponseWriter, id int64) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TaskResponse{ID: id})
}
