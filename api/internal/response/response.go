package response

import (
	"api/internal/logger"
	"encoding/json"
	"net/http"
)

// APIResponse определяет структуру ответа API
type APIResponse struct {
	Success bool        `json:"success"`        // Флаг успешности операции
	Code    int         `json:"code"`           // HTTP код ответа
	Message string      `json:"message"`        // Сообщение для пользователя
	Data    interface{} `json:"data,omitempty"` // Данные ответа (опционально)
}

// WriteAPIResponse формирует и отправляет JSON-ответ клиенту
// Устанавливает заголовки ответа, сериализует данные и логирует ошибки при неудаче
func WriteAPIResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: success,
		Code:    statusCode,
		Message: message,
		Data:    data,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("api", "failed to write a response", map[string]string{"error: ": err.Error()})
	}
}
