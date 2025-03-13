package main

import (
	"fmt"
	"go_final_project/internal/taskHandler"
	"log"
	"net/http"

	"go_final_project/internal/db"
	"go_final_project/internal/handler"
	"go_final_project/internal/task"
)

const defaultPort = "7540"
const webDir = "./web"

func main() {
	// Создаём подключение к базе данных
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Ошибка при закрытии базы данных: %v", err)
		}
	}()

	http.HandleFunc("/api/nextdate", handler.NextDateHandler)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			task.GetTaskHandler(database.DB)(w, r)
		case http.MethodPost:
			task.AddTaskHandler(database.DB)(w, r)
		case http.MethodPut:
			task.UpdateTaskHandler(database.DB)(w, r)
		case http.MethodDelete:
			task.DeleteTaskHandler(database.DB)(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			task.PostTaskDoneHandler(database.DB)(w, r)
		case http.MethodDelete:
			task.DeleteTaskDoneHandler(database.DB)(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/tasks", taskHandler.TasksHandler) // Новый маршрут для получения задач
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	fmt.Printf("Запускаем сервер на http://localhost:%s\n", defaultPort)
	err = http.ListenAndServe(":"+defaultPort, nil)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
	fmt.Println("Завершаем работу")
}
