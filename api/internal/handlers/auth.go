package handlers

import (
	"api/internal/encryption"
	"api/internal/logger"
	"api/internal/messages"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// sessionLifetime - время жизни сессии
var sessionLifetime time.Duration

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	User    repo.UserRepo    // Репозиторий пользователей
	Token   repo.TokenRepo   // Репозиторий токенов
	Session repo.SessionRepo // Репозиторий сессий
	secret  string           // Секретный ключ для шифрования
}

var serverSecretKey []byte

func init() {
	serverSecretKey = []byte(viper.GetString("crypto.serverSecretKey"))
	coef := viper.GetInt("session.lifetime") // коэффициент времени жизни сессии (в минутах)
	if coef <= 0 {
		coef = 1
	}

	sessionLifetime = time.Duration(coef) * time.Minute
}

// EncryptionKey обменивается ключами для установки защищенного соединения
func (p *AuthHandler) EncryptionKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientPublic string `json:"clientPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrCryptoParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	secret, err := encryption.DeriveSharedKeyHex(req.ClientPublic)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrKeyDerivation, map[string]string{
			messages.LogDetails: err.Error(),
			"client_pub":        req.ClientPublic,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidPublicKey, nil)
		return
	}

	p.secret = secret

	serverPublic := encryption.GetServerPublicKey()
	logger.Info(messages.ServiceAuth, messages.LogStatusParamsSent, map[string]string{
		"server_pub": serverPublic,
	})

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, map[string]string{
		"serverPublic": serverPublic,
	})
}

// LogIN аутентифицирует пользователя
func (p *AuthHandler) LogIN(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrCryptoParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrEncryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrEncryption, nil)
		return
	}

	userID, err := p.User.CheckPass(username, newPassword)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrAuthFailed, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ClientErrAuth, nil)
		return
	}

	sessionID := uuid.New()
	token, err := p.Token.GenerateJWT(sessionID, sessionLifetime)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	err = p.Session.SetSession(sessionID, userID, sessionLifetime)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	logger.Info(messages.ServiceAuth, messages.LogStatusUserAuth, map[string]string{
		messages.LogUserID: userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusAuth,
		map[string]string{messages.AuthToken: token})
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, username)
	return matched
}

func isValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 64 {
		return false
	}

	return true
}

// Register регистрирует нового пользователя
func (p *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrCryptoParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	if !isValidUsername(username) {
		logger.Error(messages.ServiceAuth, messages.LogErrInvalidUsername, map[string]string{
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidUsername, nil)
		return
	}

	if !isValidPassword(password) {
		logger.Info(messages.ServiceAuth, messages.LogErrInvalidPass, map[string]string{
			"password": password,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidPass, nil)
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrEncryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrEncryption, nil)
		return
	}

	userID, err := p.User.CreateAccount(username, newPassword)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrCreateAccount, nil)
		return
	}

	if userID == uuid.Nil {
		logger.Error(messages.ServiceAuth, messages.LogErrUserExists, map[string]string{
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrUserExists, nil)
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID, sessionLifetime)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrTokenGeneration, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	err = p.Session.SetSession(sessionID, userID, sessionLifetime)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	logger.Info(messages.ServiceAuth, messages.LogStatusUserAuth, map[string]string{messages.LogUserID: userID.String()})
	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusAuth,
		map[string]string{messages.AuthToken: token})
}

// LogOUT завершает сессию пользователя
func (p *AuthHandler) LogOUT(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get(messages.AuthToken)
	if authHeader == "" {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogDetails: messages.LogErrNoAuthToken,
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	token, err := p.Token.ParseJWT(authHeader)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: authHeader,
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	userID, err := p.Session.DeleteSession(token.SessionID)
	if err != nil {
		logger.Error(messages.ServiceAuth, messages.LogErrSessionDelete, map[string]string{
			messages.LogSessionID: token.SessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusLogOut, nil)
	logger.Info(messages.ServiceAuth, messages.LogStatusUserLogOut, map[string]string{messages.LogUserID: userID.String()})
}
