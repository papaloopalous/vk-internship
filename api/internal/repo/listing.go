package repo

import (
	"api/internal/proto/listingpb"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ListingRepoGRPC struct {
	service listingpb.ListingServiceClient // gRPC клиент для взаимодействия с сервисом листингов
}

// Проверка реализации интерфейса ListingRepo
var _ ListingRepo = &ListingRepoGRPC{}

func NewListingRepo(conn *grpc.ClientConn) *ListingRepoGRPC {
	return &ListingRepoGRPC{
		service: listingpb.NewListingServiceClient(conn),
	}
}

const (
	listingToken = "listing-token"
)

// GetAllListings получает все объявления
// userID - ID пользователя, для которого получаем объявления
// targetUser - ID пользователя, чьи объявления получаем
func (r *ListingRepoGRPC) GetAllListings(userID uuid.UUID, targetUser uuid.UUID, sortField string, sortOrder string, onlyLiked bool, page int) (listing []ListingType, totalPages int64, cuurentPage int64, err error) {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := r.service.GetAllListings(ctx, &listingpb.GetAllListingsRequest{
		TargetUserId: targetUser.String(),
		SortField:    sortField,
		SortOrder:    sortOrder,
		OnlyLiked:    onlyLiked,
		UserId:       userID.String(),
		Page:         int64(page),
	})

	if err != nil {
		return nil, 0, 0, err
	}

	for _, item := range resp.Listings {
		parsedID, err := uuid.Parse(item.Id)
		if err != nil {
			continue
		}

		parsedAuthorID, err := uuid.Parse(item.AuthorId)
		if err != nil {
			continue
		}

		listing = append(listing, ListingType{
			ID:          parsedID,
			Title:       item.Title,
			Description: item.Description,
			Address:     item.Address,
			Price:       int(item.Price),
			AuthorID:    parsedAuthorID,
			CreatedAt:   item.CreatedAt.AsTime(),
			ImageURL:    item.ImageUrl,
			Likes:       int(item.Likes),
			IsYours:     item.IsYours,
			IsLiked:     item.IsLiked,
			AuthorLogin: item.AuthorLogin,
		})
	}

	if len(listing) == 0 {
		listing = []ListingType{}
	}

	return listing, resp.TotalPages, resp.CurrentPage, nil
}

// AddListing добавляет новое объявление
func (r *ListingRepoGRPC) AddListing(listing ListingType) (id uuid.UUID, err error) {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := r.service.AddListing(ctx, &listingpb.AddListingRequest{
		Title:       listing.Title,
		Description: listing.Description,
		Address:     listing.Address,
		Price:       int64(listing.Price),
		AuthorId:    listing.AuthorID.String(),
		ImageUrl:    listing.ImageURL,
	})

	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(resp.Id)
}

// EditListing редактирует существующее объявление
func (r *ListingRepoGRPC) EditListing(listing ListingType, userID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err := r.service.EditListing(ctx, &listingpb.EditListingRequest{
		Id:          listing.ID.String(),
		Title:       listing.Title,
		Description: listing.Description,
		Address:     listing.Address,
		Price:       int64(listing.Price),
		ImageUrl:    listing.ImageURL,
		UserId:      userID.String(),
	})

	return err
}

// DeleteListing удаляет объявление
func (r *ListingRepoGRPC) DeleteListing(id uuid.UUID, userID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err := r.service.DeleteListing(ctx, &listingpb.DeleteListingRequest{
		Id:     id.String(),
		UserId: userID.String(),
	})

	return err
}

// AddLike добавляет объявление в список избранного
func (r *ListingRepoGRPC) AddLike(listingID uuid.UUID, userID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err := r.service.AddLike(ctx, &listingpb.AddLikeRequest{
		ListingId: listingID.String(),
		UserId:    userID.String(),
	})

	return err
}

// RemoveLike удаляет объявление из списка избранного
func (r *ListingRepoGRPC) RemoveLike(listingID uuid.UUID, userID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + listingToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err := r.service.RemoveLike(ctx, &listingpb.RemoveLikeRequest{
		ListingId: listingID.String(),
		UserId:    userID.String(),
	})

	return err
}
