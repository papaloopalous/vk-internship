package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"userService/userpb"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	userpb.UnimplementedUserServiceServer
	db *pgx.Conn
}

func (s *server) AddUser(ctx context.Context, req *userpb.NewUserRequest) (*userpb.UserIDResponse, error) {
	id := uuid.New()
	_, err := s.db.Exec(ctx, `
		INSERT INTO users (id, username, pass) 
		VALUES ($1, $2, $3)
	`, id, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &userpb.UserIDResponse{Id: id.String()}, nil
}

func (s *server) CheckCredentials(ctx context.Context, req *userpb.CredentialsRequest) (*userpb.CredentialsResponse, error) {
	var id uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id FROM users 
		WHERE username = $1 AND pass = $2
	`, req.Username, req.Password).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &userpb.CredentialsResponse{Id: id.String()}, nil
}

func (s *server) UserExists(ctx context.Context, req *userpb.UsernameRequest) (*userpb.UserExistsResponse, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM users WHERE username = $1
        )
    `, req.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	return &userpb.UserExistsResponse{Exists: exists}, nil
}

const (
	user = "user"
)

var acl = map[string][]string{
	// UserService methods
	"/user.UserService/AddUser":          {user},
	"/user.UserService/CheckCredentials": {user},
	"/user.UserService/UserExists":       {user},
}

func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	role, err := getRoleByToken(token)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid token")
	}

	allowedRoles, ok := acl[info.FullMethod]
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "method not allowed")
	}

	if !contains(allowedRoles, role) {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	return handler(ctx, req)
}

func getRoleByToken(token string) (string, error) {
	switch token {
	case "user-token":
		return user, nil
	default:
		return "", status.Error(codes.Unauthenticated, "unknown token")
	}
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASS")
	dbName := os.Getenv("POSTGRES_DB")
	serverPort := os.Getenv("USER_ADDR")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}

	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			log.Fatalf("failed to close connection: %v\n", err)
		}
	}()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptor))
	server := &server{db: conn}
	userpb.RegisterUserServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server is running on port %s", serverPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
