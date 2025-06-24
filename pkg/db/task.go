package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"` // Формат "20060102"
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	const query = `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to add task: %w", err)
	}
	return res.LastInsertId()
}

func GetTask(id string) (*Task, error) {
	const query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	var task Task
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

func Tasks(limit int) ([]Task, error) {
	const query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date DESC LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func TasksSearch(search string, limit int) ([]Task, error) {
	const query = `SELECT id, date, title, comment, repeat FROM scheduler 
                  WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`
	searchPattern := "%" + search + "%"
	rows, err := db.Query(query, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func TasksDate(date time.Time, limit int) ([]Task, error) {
	const query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?`
	dateStr := date.Format("20060102")
	rows, err := db.Query(query, dateStr, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by date: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// DeleteTask удаляет задачу по ID
func DeleteTask(id string) error {
	const query = `DELETE FROM scheduler WHERE id = ?`

	res, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task ID %s: %w", id, err)
	}

	return checkRowsAffected(res, id)
}

// checkRowsAffected проверяет количество затронутых строк

func checkRowsAffected(res sql.Result, id string) error {
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for ID %s: %w", id, err)
	}
	if count == 0 {
		return fmt.Errorf("%w: ID %s", ErrTaskNotFound, id)
	}
	return nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	const query = `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task ID %s: %w", task.ID, err)
	}

	return checkRowsAffected(res, task.ID)
}

// UpdateDate обновляет дату для конкретной задачи
func UpdateDate(nextDate string, id string) error {
	const query = `UPDATE scheduler SET date = ? WHERE id = ?`

	res, err := db.Exec(query, nextDate, id)
	if err != nil {
		return fmt.Errorf("failed to update date for task ID %s: %w", id, err)
	}

	return checkRowsAffected(res, id)
}
