package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON отправляет данные в формате JSON с указанным статус-кодом
func JSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON encoding failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Error отправляет ошибку в формате JSON с указанным статус-кодом
func Error(w http.ResponseWriter, message string, statusCode int) {
	JSON(w, map[string]string{"error": message}, statusCode)
}

// Success отправляет успешный пустой ответ в формате JSON
func Success(w http.ResponseWriter, statusCode int) {
	JSON(w, map[string]string{}, statusCode)
}
