package handlers

import (
	"api/internal/logger"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ListingHandler struct {
	Listing repo.ListingRepo
}

func (p *ListingHandler) GetAllListings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	sortField := r.URL.Query().Get(messages.ReqSortField)
	sortOrder := r.URL.Query().Get(messages.ReqSortOrder)
	onlyLikedStr := r.URL.Query().Get(messages.ReqOnlyLiked)
	targetUserId := r.URL.Query().Get(messages.ReqTargetUserID)
	page := r.URL.Query().Get(messages.ReqPage)
	minPrice := r.URL.Query().Get(messages.ReqMinPrice)
	maxPrice := r.URL.Query().Get(messages.ReqMaxPrice)

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogPage: page,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	minPriceInt, err := strconv.Atoi(minPrice)
	if err != nil || minPriceInt < 1 || minPriceInt > 100000000 {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogPrice: minPrice,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	maxPriceInt, err := strconv.Atoi(maxPrice)
	if err != nil || maxPriceInt < 1 || maxPriceInt > 100000000 || maxPriceInt < minPriceInt {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogPrice: maxPrice,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	onlyLiked := onlyLikedStr == "true"

	var targetUser uuid.UUID
	if targetUserId != "" {
		targetUser, err = uuid.Parse(targetUserId)
	}
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidUUID, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidUUID, nil)
		return
	}

	filter := repo.ListingFilter{
		UserID:     userID,
		TargetUser: targetUser,
		SortField:  sortField,
		SortOrder:  sortOrder,
		OnlyLiked:  onlyLiked,
		Page:       pageInt,
		MinPrice:   minPriceInt,
		MaxPrice:   maxPriceInt,
	}

	listings, totalPages, currentPage, err := p.Listing.GetAllListings(filter)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	resp := map[string]interface{}{
		messages.LogListings:    listings,
		messages.LogTotalPages:  totalPages,
		messages.LogCurrentPage: currentPage,
	}

	logger.Info(messages.ServiceListing, messages.LogStatusListingsFetched, map[string]string{
		messages.LogCount: strconv.Itoa(len(listings)),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, resp)
}

func (p *ListingHandler) AddListing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Address     string `json:"address"`
		Price       int    `json:"price"`
		ImageBase64 string `json:"image_base64"`
		ImageName   string `json:"image_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	if req.Title == "" || req.Address == "" || req.ImageBase64 == "" || req.ImageName == "" {
		logger.Error(messages.ServiceListing, messages.LogErrMissingFields, nil)
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrMissingFields, nil)
		return
	}

	if len(req.Title) < 3 || len(req.Title) > 100 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidTitle, map[string]string{
			messages.LogTitleLength: strconv.Itoa(len(req.Title)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidTitle, nil)
		return
	}

	if len(req.Description) > 1000 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidDescription, map[string]string{
			messages.LogDescLength: strconv.Itoa(len(req.Description)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidDescription, nil)
		return
	}

	if len(req.Address) < 5 || len(req.Address) > 200 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidAddress, map[string]string{
			messages.LogAddressLength: strconv.Itoa(len(req.Address)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidAddress, nil)
		return
	}

	if req.Price <= 0 || req.Price > 100000000 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidPrice, map[string]string{
			messages.LogPrice: strconv.Itoa(req.Price),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidPrice, nil)
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidImage, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidImage, nil)
		return
	}

	const maxFileSize = 5 << 20
	if len(imageData) > maxFileSize {
		logger.Error(messages.ServiceListing, messages.LogErrImageTooLarge, map[string]string{
			messages.LogImageSize: strconv.Itoa(len(imageData)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrImageTooLarge, nil)
		return
	}

	filetype := http.DetectContentType(imageData[:512])
	if filetype != "image/jpeg" && filetype != "image/png" {
		logger.Error(messages.ServiceListing, messages.LogErrUnsupportedImageType, map[string]string{
			messages.LogImageType: filetype,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrUnsupportedImageType, nil)
		return
	}

	filename := uuid.New().String() + filepath.Ext(req.ImageName)
	savePath := filepath.Join("uploads", filename)
	if err := os.WriteFile(savePath, imageData, 0644); err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrFileSave, map[string]string{
			messages.LogDetails: err.Error(),
			messages.LogPath:    savePath,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrFileSave, nil)
		return
	}
	imageURL := "/uploads/" + filename

	listing := repo.ListingType{
		Title:       req.Title,
		Description: req.Description,
		Address:     req.Address,
		Price:       req.Price,
		AuthorID:    userID,
		ImageURL:    imageURL,
	}

	id, err := p.Listing.AddListing(listing)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	resp := map[string]interface{}{
		messages.LogID:       id,
		messages.LogImageURL: imageURL,
	}

	logger.Info(messages.ServiceListing, messages.LogStatusListingAdded, map[string]string{
		messages.LogListingID: id.String(),
		messages.LogUserID:    userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusListingAdded, resp)
}

func (p *ListingHandler) EditListing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Address     string    `json:"address"`
		Price       int       `json:"price"`
		ImageBase64 string    `json:"image_base64"`
		ImageName   string    `json:"image_name"`
		ID          uuid.UUID `json:"listing_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	if req.Title == "" || req.Address == "" || req.ImageBase64 == "" || req.ImageName == "" {
		logger.Error(messages.ServiceListing, messages.LogErrMissingFields, nil)
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrMissingFields, nil)
		return
	}

	if len(req.Title) < 3 || len(req.Title) > 100 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidTitle, map[string]string{
			messages.LogTitleLength: strconv.Itoa(len(req.Title)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidTitle, nil)
		return
	}

	if len(req.Description) > 1000 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidDescription, map[string]string{
			messages.LogDescLength: strconv.Itoa(len(req.Description)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidDescription, nil)
		return
	}

	if len(req.Address) < 5 || len(req.Address) > 200 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidAddress, map[string]string{
			messages.LogAddressLength: strconv.Itoa(len(req.Address)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidAddress, nil)
		return
	}

	if req.Price <= 0 || req.Price > 100_000_000 {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidPrice, map[string]string{
			messages.LogPrice: strconv.Itoa(req.Price),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidPrice, nil)
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidImage, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidImage, nil)
		return
	}

	const maxFileSize = 5 << 20
	if len(imageData) > maxFileSize {
		logger.Error(messages.ServiceListing, messages.LogErrImageTooLarge, map[string]string{
			messages.LogImageSize: strconv.Itoa(len(imageData)),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrImageTooLarge, nil)
		return
	}

	filetype := http.DetectContentType(imageData[:512])
	if filetype != "image/jpeg" && filetype != "image/png" {
		logger.Error(messages.ServiceListing, messages.LogErrUnsupportedImageType, map[string]string{
			messages.LogImageType: filetype,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrUnsupportedImageType, nil)
		return
	}

	filename := uuid.New().String() + filepath.Ext(req.ImageName)
	savePath := filepath.Join("uploads", filename)
	if err := os.WriteFile(savePath, imageData, 0644); err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrFileSave, map[string]string{
			messages.LogDetails: err.Error(),
			messages.LogPath:    savePath,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrFileSave, nil)
		return
	}
	imageURL := "/uploads/" + filename

	listing := repo.ListingType{
		ID:          req.ID,
		Title:       req.Title,
		Description: req.Description,
		Address:     req.Address,
		Price:       req.Price,
		AuthorID:    userID,
		ImageURL:    imageURL,
	}

	err = p.Listing.EditListing(listing, userID)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:   err.Error(),
			messages.LogListingID: req.ID.String(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	logger.Info(messages.ServiceListing, messages.LogStatusListingEdited, map[string]string{
		messages.LogListingID: req.ID.String(),
		messages.LogUserID:    userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusListingEdited, map[string]string{
		messages.LogImageURL: imageURL,
	})
}

func (p *ListingHandler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	vars := mux.Vars(r)
	listingIDStr, ok := vars["id"]
	if !ok {
		logger.Error(messages.ServiceListing, messages.LogErrMissingID, nil)
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrMissingID, nil)
		return
	}

	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrInvalidUUID, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidUUID, nil)
		return
	}

	err = p.Listing.DeleteListing(listingID, userID)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:   err.Error(),
			messages.LogListingID: listingID.String(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	logger.Info(messages.ServiceListing, messages.LogStatusListingDeleted, map[string]string{
		messages.LogListingID: listingID.String(),
		messages.LogUserID:    userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusListingDeleted, nil)
}

func (p *ListingHandler) AddLike(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	var req struct {
		ListingID uuid.UUID `json:"listing_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ListingID == uuid.Nil {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	err := p.Listing.AddLike(req.ListingID, userID)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:   err.Error(),
			messages.LogListingID: req.ListingID.String(),
			messages.LogUserID:    userID.String(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	logger.Info(messages.ServiceListing, messages.LogStatusLikeAdded, map[string]string{
		messages.LogListingID: req.ListingID.String(),
		messages.LogUserID:    userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusLikeAdded, nil)
}

func (p *ListingHandler) RemoveLike(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	var req struct {
		ListingID uuid.UUID `json:"listing_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ListingID == uuid.Nil {
		logger.Error(messages.ServiceListing, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	err := p.Listing.RemoveLike(req.ListingID, userID)
	if err != nil {
		logger.Error(messages.ServiceListing, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:   err.Error(),
			messages.LogListingID: req.ListingID.String(),
			messages.LogUserID:    userID.String(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDBQuery, nil)
		return
	}

	logger.Info(messages.ServiceListing, messages.LogStatusLikeRemoved, map[string]string{
		messages.LogListingID: req.ListingID.String(),
		messages.LogUserID:    userID.String(),
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusLikeRemoved, nil)
}
