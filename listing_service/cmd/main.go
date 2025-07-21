package main

import (
	"context"
	"errors"
	"fmt"
	"listingService/listingpb"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	listingpb.UnimplementedListingServiceServer
	sql *pgx.Conn
}

var limit int

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	limitStr := os.Getenv("LISTING_LIMIT")
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		log.Fatalf("invalid LISTING_LIMIT: %v", err)
	}
}

const (
	listing = "listing"
)

var acl = map[string][]string{
	// ListingService methods
	"/listingpb.ListingService/GetAllListings": {listing},
	"/listingpb.ListingService/AddListing":     {listing},
	"/listingpb.ListingService/EditListing":    {listing},
	"/listingpb.ListingService/DeleteListing":  {listing},
	"/listingpb.ListingService/AddLike":        {listing},
	"/listingpb.ListingService/RemoveLike":     {listing},
}

// UnaryInterceptor — перехватчик запросов
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
	case "listing-token":
		return listing, nil
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

func (s *server) GetAllListings(ctx context.Context, req *listingpb.GetAllListingsRequest) (*listingpb.GetAllListingsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}

	// Нормализация параметров сортировки
	sortField := "created_at"
	sortOrder := "DESC"

	switch req.SortField {
	case "price":
		sortField = "price"
	case "created_at":
		sortField = "created_at"
	}

	if strings.ToUpper(req.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	// Получаем лайкнутые ID, если передан UserId
	var likedMap map[string]bool
	var likedListingIDs []uuid.UUID

	if req.UserId != "" {
		userUUID, err := uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
		}

		err = s.sql.QueryRow(ctx, `SELECT liked_listings FROM users WHERE id = $1`, userUUID).Scan(&likedListingIDs)
		if err != nil {
			log.Printf("Warning: could not load liked_listings for user %s: %v", req.UserId, err)
			likedListingIDs = []uuid.UUID{}
		}

		likedMap = make(map[string]bool)
		for _, id := range likedListingIDs {
			likedMap[id.String()] = true
		}
	}

	// Если OnlyLiked включен, но список пуст — сразу возвращаем пусто
	if req.OnlyLiked && len(likedListingIDs) == 0 {
		return &listingpb.GetAllListingsResponse{
			Listings:    []*listingpb.Listing{},
			TotalPages:  0,
			CurrentPage: 0,
		}, nil
	}

	// Базовый SQL-запрос
	baseQuery := `
        SELECT 
            l.id, l.title, l.description, l.address, l.price, 
            l.author_id, u.username as author_username,
            l.created_at, l.image_url, l.likes
        FROM listings l
        LEFT JOIN users u ON l.author_id = u.id
    `

	var conditions []string
	var args []interface{}
	argIdx := 1

	// Фильтр по избранным
	if req.OnlyLiked && len(likedListingIDs) > 0 {
		var likedIDs []string
		for _, id := range likedListingIDs {
			likedIDs = append(likedIDs, id.String())
		}
		conditions = append(conditions, fmt.Sprintf("l.id = ANY($%d)", argIdx))
		args = append(args, likedIDs)
		argIdx++
	}

	// Фильтр по автору
	if req.TargetUserId != "" && req.TargetUserId != uuid.Nil.String() {
		conditions = append(conditions, fmt.Sprintf("l.author_id = $%d", argIdx))
		args = append(args, req.TargetUserId)
		argIdx++
	}

	conditions = append(conditions, fmt.Sprintf("l.price >= $%d", argIdx))
	args = append(args, req.MinPrice)
	argIdx++
	conditions = append(conditions, fmt.Sprintf("l.price <= $%d", argIdx))
	args = append(args, req.MaxPrice)
	argIdx++

	// Подсчёт общего количества записей
	countQuery := baseQuery
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	countQuery = "SELECT COUNT(*) FROM (" + countQuery + ") AS filtered_listings"

	var totalItems int
	err := s.sql.QueryRow(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count listings: %v", err)
	}

	// Пагинация
	totalPages := int64((totalItems + limit - 1) / limit)
	if totalPages == 0 {
		totalPages = 1
	}
	if req.Page > totalPages {
		req.Page = totalPages
	}
	offset := (req.Page - 1) * int64(limit)

	// Финальный запрос
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortField, sortOrder, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.sql.Query(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query error: %v", err)
	}
	defer rows.Close()

	// Сборка результата
	var listings []*listingpb.Listing
	for rows.Next() {
		var l listingpb.Listing
		var createdAt time.Time
		var authorUsername *string

		if err := rows.Scan(
			&l.Id,
			&l.Title,
			&l.Description,
			&l.Address,
			&l.Price,
			&l.AuthorId,
			&authorUsername,
			&createdAt,
			&l.ImageUrl,
			&l.Likes,
		); err != nil {
			return nil, status.Errorf(codes.Internal, "scan error: %v", err)
		}

		if authorUsername != nil {
			l.AuthorLogin = *authorUsername
		}

		l.CreatedAt = timestamppb.New(createdAt)
		l.IsYours = (req.UserId != "" && l.AuthorId == req.UserId)

		if likedMap != nil {
			l.IsLiked = likedMap[l.Id]
		}

		listings = append(listings, &l)
	}

	return &listingpb.GetAllListingsResponse{
		Listings:    listings,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
	}, nil
}

func (s *server) AddListing(ctx context.Context, req *listingpb.AddListingRequest) (*listingpb.AddListingResponse, error) {
	id := uuid.New()
	createdAt := time.Now()

	_, err := s.sql.Exec(ctx, `
        INSERT INTO listings (id, title, description, address, price, author_id, created_at, image_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `,
		id,
		req.Title,
		req.Description,
		req.Address,
		req.Price,
		req.AuthorId,
		createdAt,
		req.ImageUrl,
	)
	if err != nil {
		return nil, err
	}

	return &listingpb.AddListingResponse{
		Id: id.String(),
	}, nil
}

func (s *server) EditListing(ctx context.Context, req *listingpb.EditListingRequest) (*listingpb.Empty, error) {
	var authorID string
	err := s.sql.QueryRow(ctx, `SELECT author_id FROM listings WHERE id = $1`, req.Id).Scan(&authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "listing not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to query listing: %v", err)
	}

	if authorID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "you are not the owner of this listing")
	}

	_, err = s.sql.Exec(ctx, `
        UPDATE listings
        SET title = $1, description = $2, address = $3, price = $4, image_url = $5
        WHERE id = $6
    `,
		req.Title,
		req.Description,
		req.Address,
		req.Price,
		req.ImageUrl,
		req.Id,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update listing: %v", err)
	}
	return &listingpb.Empty{}, nil
}

func (s *server) DeleteListing(ctx context.Context, req *listingpb.DeleteListingRequest) (*listingpb.Empty, error) {
	var authorID string
	err := s.sql.QueryRow(ctx, `SELECT author_id FROM listings WHERE id = $1`, req.Id).Scan(&authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "listing not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to query listing: %v", err)
	}

	if authorID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "you are not the owner of this listing")
	}

	_, err = s.sql.Exec(ctx, `DELETE FROM listings WHERE id = $1`, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete listing: %v", err)
	}
	return &listingpb.Empty{}, nil
}

func (s *server) AddLike(ctx context.Context, req *listingpb.AddLikeRequest) (*listingpb.Empty, error) {
	_, err := s.sql.Exec(ctx, `
        UPDATE listings SET likes = likes + 1 WHERE id = $1
    `, req.ListingId)
	if err != nil {
		return nil, err
	}
	_, err = s.sql.Exec(ctx, `
        UPDATE users SET liked_listings = array_append(liked_listings, $1) WHERE id = $2
    `, req.ListingId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &listingpb.Empty{}, nil
}

func (s *server) RemoveLike(ctx context.Context, req *listingpb.RemoveLikeRequest) (*listingpb.Empty, error) {
	_, err := s.sql.Exec(ctx, `
        UPDATE listings SET likes = likes - 1 WHERE id = $1 AND likes > 0
    `, req.ListingId)
	if err != nil {
		return nil, err
	}
	_, err = s.sql.Exec(ctx, `
        UPDATE users SET liked_listings = array_remove(liked_listings, $1) WHERE id = $2
    `, req.ListingId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &listingpb.Empty{}, nil
}

func main() {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASS")
	dbName := os.Getenv("POSTGRES_DB")

	serverPort := os.Getenv("LISTING_ADDR")

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
	server := &server{
		sql: conn,
	}
	listingpb.RegisterListingServiceServer(grpcServer, server)

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
