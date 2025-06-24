package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vladimirpaholov/td/pkg/db"
	"github.com/vladimirpaholov/td/pkg/response"
)

// completeHandler отмечает задачу как выполненную:
// - удаляет задачу без правила повтора
// - обновляет дату для повторяющихся задач
func completeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		response.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	if err := handleTaskCompletion(id); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(w, http.StatusOK)
}

// handleTaskCompletion содержит основную логику обработки выполнения задачи
func handleTaskCompletion(id string) error {
	task, err := db.GetTask(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.Repeat == "" {
		return deleteNonRepeatingTask(id)
	}
	return updateRepeatingTaskDate(id, task)
}

// deleteNonRepeatingTask удаляет неповторяющуюся задачу
func deleteNonRepeatingTask(id string) error {
	if err := db.DeleteTask(id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// updateRepeatingTaskDate обновляет дату для повторяющейся задачи
func updateRepeatingTaskDate(id string, task *db.Task) error {
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		return fmt.Errorf("failed to calculate next date: %w", err)
	}

	if err := db.UpdateDate(nextDate, id); err != nil {
		return fmt.Errorf("failed to update task date: %w", err)
	}
	return nil
}
