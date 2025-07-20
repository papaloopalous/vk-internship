package repo

import (
	"api/internal/proto/sessionpb"
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SessionRepoGRPC реализует взаимодействие с сервисом сессий через gRPC
type SessionRepoGRPC struct {
	db sessionpb.SessionServiceClient // gRPC клиент для взаимодействия с сервисом сессий
}

// Проверка реализации интерфейса SessionRepo
var _ SessionRepo = &SessionRepoGRPC{}

// NewSessionRepo создает новый экземпляр репозитория сессий
func NewSessionRepo(conn *grpc.ClientConn) *SessionRepoGRPC {
	return &SessionRepoGRPC{
		db: sessionpb.NewSessionServiceClient(conn),
	}
}

const (
	sessionToken = "session-token"
)

// GetSession получает информацию о сессии из базы данных
func (r *SessionRepoGRPC) GetSession(sessionID uuid.UUID) (userID uuid.UUID, err error) {
	md := metadata.New(map[string]string{
		authorization: bearer + sessionToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.GetSession(ctx, &sessionpb.SessionIDRequest{
		SessionId: sessionID.String(),
	})
	if err != nil {
		return uuid.Nil, err
	}

	userID, err = uuid.Parse(resp.UserId)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// SetSession создает новую сессию в базе данных
func (r *SessionRepoGRPC) SetSession(sessionID uuid.UUID, userID uuid.UUID, sessionLifetime time.Duration) error {
	md := metadata.New(map[string]string{
		authorization: bearer + sessionToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	expiresAt := time.Now().Add(sessionLifetime).Unix()

	_, err := r.db.SetSession(ctx, &sessionpb.SetSessionRequest{
		SessionId: sessionID.String(),
		UserId:    userID.String(),
		ExpiresAt: expiresAt,
	})
	return err
}

// DeleteSession удаляет сессию из базы данных и возвращает ID пользователя
func (r *SessionRepoGRPC) DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error) {
	md := metadata.New(map[string]string{
		authorization: bearer + sessionToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.DeleteSession(ctx, &sessionpb.SessionIDRequest{
		SessionId: sessionID.String(),
	})
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(resp.UserId)
}
