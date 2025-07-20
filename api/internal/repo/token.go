package repo

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MyClaims определяет структуру данных JWT токена
type MyClaims struct {
	SessionID            uuid.UUID `json:"sessionID"` // Идентификатор сессии
	jwt.RegisteredClaims           // Стандартные поля JWT
}

// TokenData хранит ключ для подписи JWT токенов
type TokenData struct {
	key []byte // Секретный ключ для подписи токенов
}

// TokenRepo определяет методы для работы с JWT токенами
type TokenRepo interface {
	// GetData возвращает ключ для подписи токенов
	GetData() (token []byte)

	// SetData устанавливает ключ для подписи токенов
	SetData(token string)

	// GenerateJWT создает новый JWT токен
	GenerateJWT(sessionID uuid.UUID, sessionLifetime time.Duration) (string, error)

	// ParseJWT проверяет и извлекает данные из JWT токена
	ParseJWT(tokenString string) (*MyClaims, error)
}

// Проверка реализации интерфейса
var _ TokenRepo = &TokenData{}

// NewTokenRepo создает новый экземпляр репозитория токенов
func NewTokenRepo() *TokenData {
	return &TokenData{
		key: make([]byte, 0),
	}
}

// GetData возвращает ключ для подписи токенов
func (p *TokenData) GetData() (token []byte) {
	res := p.key
	return res
}

// SetData устанавливает ключ для подписи токенов
func (p *TokenData) SetData(token string) {
	p.key = []byte(token)
}

// GenerateJWT создает новый JWT токен с указанным ID сессии
func (p *TokenData) GenerateJWT(sessionID uuid.UUID, sessionLifetime time.Duration) (string, error) {
	claims := MyClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	res, err := token.SignedString(p.GetData())
	if err != nil {
		return res, err
	}

	return res, nil
}

// ParseJWT проверяет подпись и извлекает данные из JWT токена
func (p *TokenData) ParseJWT(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return p.GetData(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
