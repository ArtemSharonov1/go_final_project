package taskHandler

import (
	"database/sql"
	"encoding/json"
	"go_final_project/internal/db"
	"net/http"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	dbInstance, err := db.NewDB()
	if err != nil {
		http.Error(w, `{"error": "Ошибка подключения к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer dbInstance.DB.Close()

	rows, err := dbInstance.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50")
	if err != nil {
		http.Error(w, `{"error": "Ошибка выполнения запроса"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		var repeat sql.NullString
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &repeat); err != nil {
			http.Error(w, `{"error": "Ошибка обработки данных"}`, http.StatusInternalServerError)
			return
		}
		if repeat.Valid {
			t.Repeat = repeat.String
		} else {
			t.Repeat = ""
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error": "Ошибка обработки строк"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}
