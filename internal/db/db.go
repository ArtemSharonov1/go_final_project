package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite
	"log"
	"os"
)

const dbFileName = "scheduler.db"

// DB представляет собой обёртку для работы с базой данных
type DB struct {
	DB *sql.DB // Экспортируемое поле
}

// NewDB создаёт новое подключение к базе данных и возвращает экземпляр DB
func NewDB() (*DB, error) {
	// Получаем путь к исполняемому файлу
	dbFile, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пути к приложению: %v", err)
	}

	// Определяем путь к файлу базы данных
	dbFile = "./internal/db/scheduler.db"
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		// Если файл базы данных не существует, устанавливаем флаг install
		install = true
	}

	// Открываем базу данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии базы данных: %v", err)
	}

	// Если флаг install установлен, создаем таблицу и индекс
	if install {
		err = createTableAndIndex(db)
		if err != nil {
			return nil, fmt.Errorf("ошибка при создании таблицы и индекса: %v", err)
		}
		log.Println("База данных успешно создана и настроена.")
	}

	return &DB{DB: db}, nil
}

// Close закрывает соединение с базой данных
func (d *DB) Close() error {
	if err := d.DB.Close(); err != nil { // Используем экспортируемое поле DB
		log.Printf("Ошибка при закрытии базы данных: %v", err)
		return err
	}
	return nil
}

// createTableAndIndex создаёт таблицу и индекс, если они не существуют
func createTableAndIndex(db *sql.DB) error {
	// SQL-запрос для создания таблицы scheduler
	query := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);
    `

	// Выполняем SQL-запрос
	_, err := db.Exec(query)
	return err
}

// структура задачи
type Task struct {
	ID      int
	Date    string
	Title   string
	Comment string
	Repeat  string
}
