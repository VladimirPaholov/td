package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vladimirpaholov/td/pkg/db"
	"github.com/vladimirpaholov/td/pkg/response"
)

const (
	defaultLimit    = 50
	maxLimit        = 100
	dateInputFormat = "02.01.2006"
)

// tasksHandler обрабатывает запросы на получение списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.FormValue("search")
	limit := getLimit(r)

	var tasks []db.Task
	var err error

	if search == "" {
		tasks, err = db.Tasks(limit)
	} else {
		tasks, err = handleSearchRequest(search, limit)
	}

	if err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем в формат, ожидаемый тестами
	taskMaps := make([]map[string]string, len(tasks))
	for i, task := range tasks {
		taskMaps[i] = map[string]string{
			"id":      task.ID,
			"date":    task.Date,
			"title":   task.Title,
			"comment": task.Comment,
			"repeat":  task.Repeat,
		}
	}

	response.JSON(w, map[string]interface{}{"tasks": taskMaps}, http.StatusOK)
}

// handleSearchRequest обрабатывает поисковые запросы
func handleSearchRequest(search string, limit int) ([]db.Task, error) {
	if date, err := time.Parse(dateInputFormat, search); err == nil {
		return db.TasksDate(date, limit)
	}
	return db.TasksSearch(search, limit)
}

// getLimit извлекает и валидирует параметр limit
func getLimit(r *http.Request) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultLimit
	}

	if limit > maxLimit {
		return maxLimit
	}

	return limit
}
