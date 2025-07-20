package repo

import (
	"time"

	"github.com/google/uuid"
)

const (
	authorization = "authorization"
	bearer        = "Bearer "
)

// UserRepo определяет методы для работы с пользователями в системе
type UserRepo interface {
	// CheckPass проверяет учетные данные пользователя
	CheckPass(username string, pass string) (userID uuid.UUID, err error)

	// CreateAccount создает новую учетную запись
	CreateAccount(username string, pass string) (userID uuid.UUID, err error)
}

// SessionRepo определяет методы для работы с сессиями
type SessionRepo interface {
	// GetSession получает информацию о сессии
	GetSession(sessionID uuid.UUID) (userID uuid.UUID, err error)

	// SetSession создает новую сессию
	SetSession(sessionID uuid.UUID, userID uuid.UUID, sessionLifetime time.Duration) error

	// DeleteSession удаляет сессию
	DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error)
}

type ListingType struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	Price       int       `json:"price"`
	AuthorID    uuid.UUID `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
	ImageURL    string    `json:"image_url"`
	Likes       int       `json:"likes"`
	IsYours     bool      `json:"is_yours"`
	IsLiked     bool      `json:"is_liked"`
	AuthorLogin string    `json:"author_login"`
}

// ListingRepo определяет методы для работы с объявлениями
type ListingRepo interface {
	// GetAllListings получает все объявления
	GetAllListings(userID uuid.UUID, targetUser uuid.UUID, sortField string, sortOrder string, onlyLiked bool, page int) (listing []ListingType, totalPages int64, cuurentPage int64, err error)

	// AddListing добавляет новое объявление
	AddListing(listing ListingType) (id uuid.UUID, err error)

	// EditListing редактирует существующее объявление
	EditListing(listing ListingType, userID uuid.UUID) error

	// DeleteListing удаляет объявление
	DeleteListing(id uuid.UUID, userID uuid.UUID) error

	// AddLike добавляет объявление в список избранного
	AddLike(listingID uuid.UUID, userID uuid.UUID) error

	// RemoveLike удаляет объявление из списка избранного
	RemoveLike(listingID uuid.UUID, userID uuid.UUID) error
}
