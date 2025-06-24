package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vladimirpaholov/td/pkg/db"
	"github.com/vladimirpaholov/td/pkg/response"
)

// taskHandler обрабатывает операции с задачей
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		response.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// deleteTaskHandler обработчик удаления задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		response.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		response.Error(w, "Failed to delete task: "+err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(w, http.StatusOK)
}

// getTaskHandler возвращает задачу по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		response.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		response.Error(w, "Failed to get task: "+err.Error(), http.StatusBadRequest)
		return
	}

	response.JSON(w, task, http.StatusOK)
}

// updateTaskHandler обработчик обновления задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		response.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateTask(&task); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		response.Error(w, "Failed to update task: "+err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(w, http.StatusOK)
}

// addTaskHandler обработчик добавления новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		response.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateTask(&task); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		response.Error(w, "Failed to add task: "+err.Error(), http.StatusBadRequest)
		return
	}

	response.JSON(w, map[string]int64{"id": id}, http.StatusCreated)
}

// validateTask проверяет валидность данных задачи
func validateTask(task *db.Task) error {
	if task.Title == "" {
		return errors.New("Task title is required")
	}

	if task.ID == "" && task.Date == "" {
		task.Date = time.Now().Format(TimeFormat)
	}

	return validateAndAdjustDate(task)
}

// validateAndAdjustDate проверяет и корректирует дату задачи
func validateAndAdjustDate(task *db.Task) error {
	if task.Date == "" {
		return nil
	}

	parsedDate, err := time.Parse(TimeFormat, task.Date)
	if err != nil {
		return fmt.Errorf("Invalid date format: %v", err)
	}

	now := normalizeDate(time.Now())
	parsedDate = normalizeDate(parsedDate)

	if parsedDate.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(TimeFormat)
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("Failed to calculate next date: %v", err)
			}
			task.Date = next
		}
	}

	return nil
}

// normalizeDate обнуляет время в дате для корректного сравнения
func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
