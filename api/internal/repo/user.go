package repo

import (
	"api/internal/proto/userpb"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserRepoGRPC реализует взаимодействие с сервисом пользователей через gRPC
type UserRepoGRPC struct {
	db userpb.UserServiceClient // gRPC клиент для взаимодействия с сервисом пользователей
}

// Проверка реализации интерфейса UserRepo
var _ UserRepo = &UserRepoGRPC{}

// NewUserRepo создает новый экземпляр репозитория пользователей
func NewUserRepo(conn *grpc.ClientConn) *UserRepoGRPC {
	return &UserRepoGRPC{
		db: userpb.NewUserServiceClient(conn),
	}
}

const (
	userToken = "user-token"
)

// CreateAccount создает новую учетную запись
func (r *UserRepoGRPC) CreateAccount(username string, pass string) (uuid.UUID, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	existsResp, err := r.db.UserExists(ctx, &userpb.UsernameRequest{Username: username})
	if err != nil {
		return uuid.Nil, err
	}
	if existsResp.Exists {
		return uuid.Nil, err
	}
	resp, err := r.db.AddUser(ctx, &userpb.NewUserRequest{
		Username: username,
		Password: pass,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.MustParse(resp.Id), nil
}

// CheckPass проверяет учетные данные пользователя
func (r *UserRepoGRPC) CheckPass(username string, pass string) (uuid.UUID, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.CheckCredentials(ctx, &userpb.CredentialsRequest{
		Username: username,
		Password: pass,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.MustParse(resp.Id), nil
}
