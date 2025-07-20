package middleware

import (
	"api/internal/logger"
	"api/internal/messages"
	"api/internal/repo"
	"api/internal/response"
	"context"
	"net/http"

	"github.com/google/uuid"
)

// MiddlewareHandler содержит репозитории для проверки аутентификации и авторизации
type MiddlewareHandler struct {
	User    repo.UserRepo    // Репозиторий пользователей
	Token   repo.TokenRepo   // Репозиторий токенов
	Session repo.SessionRepo // Репозиторий сессий
}

// contextKey определяет тип ключа для контекста
type contextKey string

// userKey - ключ для хранения ID пользователя в контексте
const userKey contextKey = "UserKey"

// CheckSes проверяет сессию и права доступа пользователя
func (p *MiddlewareHandler) CheckSes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(messages.AuthToken)
		if authHeader == "" {
			response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoToken, nil)
			logger.Info(messages.ServiceMiddleware, messages.LogErrNoAuthToken, nil)
			return
		}

		token, err := p.Token.ParseJWT(authHeader)
		if err != nil {
			response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrBadToken, nil)
			logger.Error(messages.ServiceMiddleware, messages.LogErrParseToken, map[string]string{
				messages.LogDetails: err.Error(),
			})
			return
		}

		userID, err := p.Session.GetSession(token.SessionID)
		if err != nil {
			response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ClientErrNoSession, nil)
			logger.Error(messages.ServiceMiddleware, messages.LogErrSessionNotFound, map[string]string{
				messages.LogSessionID: token.SessionID.String(),
				messages.LogDetails:   err.Error(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), userKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CheckSesWithNilOnError при ошибке авторизации кладёт uuid.Nil в контекст
func (p *MiddlewareHandler) CheckSesWithNilOnError(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(messages.AuthToken)
		var userID uuid.UUID

		if authHeader == "" {
			userID = uuid.Nil
		} else {
			token, err := p.Token.ParseJWT(authHeader)
			if err != nil {
				userID = uuid.Nil
			} else {
				id, err := p.Session.GetSession(token.SessionID)
				if err != nil {
					userID = uuid.Nil
				} else {
					userID = id
				}
			}
		}

		ctx := context.WithValue(r.Context(), userKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetContext извлекает ID пользователя из контекста
func GetContext(ctx context.Context) (userID uuid.UUID) {
	userID = ctx.Value(userKey).(uuid.UUID)
	return userID
}
