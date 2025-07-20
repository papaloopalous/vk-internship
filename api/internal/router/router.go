package router

import (
	"api/internal/encryption"
	"api/internal/handlers"
	"api/internal/healthcheck"
	"api/internal/logger"
	"api/internal/middleware"
	"api/internal/repo"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// tokenRepo хранит глобальный репозиторий для работы с JWT токенами
var tokenRepo = repo.NewTokenRepo()

var coef1 int // коэффициент для таймаута соединения с микросервисами
var coef2 int // коэффициент для частоты проверки состояния соединений

// Адреса микросервисов
var userAddr, sessionAddr, listingAddr string

// init инициализирует секретный ключ для JWT токенов
func init() {
	tokenRepo.SetData("secret jwt key")
	coef1 = viper.GetInt("api.timeout")
	coef2 = viper.GetInt("api.healthcheckInterval")
	userAddr = viper.GetString("user.addr")
	sessionAddr = viper.GetString("session.addr")
	listingAddr = viper.GetString("listing.addr")
}

func gracefulStop(healthcheck *healthcheck.GrpcHealthChecker) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	healthcheck.Stop()
}

// CreateNewRouter создает и настраивает роутер приложения
// Устанавливает соединения с микросервисами, инициализирует обработчики и настраивает маршруты
func CreateNewRouter() *mux.Router {
	// Создаем контекст с таймаутом для установки соединений
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(coef1)*time.Second)
	defer cancel()

	// Устанавливаем соединения с микросервисами
	userConn, err := grpc.DialContext(ctx, userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}

	sessionConn, err := grpc.DialContext(ctx, sessionAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to session service: %v", err)
	}

	listingConn, err := grpc.DialContext(ctx, listingAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to listing service: %v", err)
	}

	// Инициализируем проверку здоровья сервисов
	healthChecker := healthcheck.NewHealthChecker(time.Duration(coef2) * time.Second)

	go gracefulStop(healthChecker)

	// Добавляем соединения в проверку состояния
	healthChecker.AddConnection("user-service", userConn)
	healthChecker.AddConnection("session-service", sessionConn)
	healthChecker.AddConnection("listing-service", listingConn)

	logger.InitLogger("logs")

	// Создаем репозитории для работы с данными
	userRepo := repo.NewUserRepo(userConn)
	sessionRepo := repo.NewSessionRepo(sessionConn)
	listingRepo := repo.NewListingRepo(listingConn)

	// Создаем обработчики запросов
	authHandler := &handlers.AuthHandler{
		User:    userRepo,
		Token:   tokenRepo,
		Session: sessionRepo,
	}

	middlewareHandler := &middleware.MiddlewareHandler{
		User:    userRepo,
		Session: sessionRepo,
		Token:   tokenRepo,
	}

	listingHandler := &handlers.ListingHandler{
		Listing: listingRepo,
	}

	// Создаем основной роутер
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("statusOK"))
	}).Methods("GET")

	// Настраиваем раздачу статических файлов
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Маршруты для шифрования и аутентификации
	router.HandleFunc("/api/key-exchange", authHandler.EncryptionKey).Methods("POST")
	router.HandleFunc("/api/crypto-params", encryption.GetCryptoParams).Methods("GET")

	router.HandleFunc("/api/login", authHandler.LogIN).Methods("POST")
	router.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/logout", authHandler.LogOUT).Methods("DELETE")

	// Маршруты для авторизованных пользователей
	userRouter := router.NewRoute().Subrouter()
	userRouter.Use(middlewareHandler.CheckSes)
	userRouter.HandleFunc("/api/listings", listingHandler.AddListing).Methods("POST")
	userRouter.HandleFunc("/api/edit", listingHandler.EditListing).Methods("POST")
	userRouter.HandleFunc("/api/listings/{id}", listingHandler.DeleteListing).Methods("DELETE")
	userRouter.HandleFunc("/api/addlike", listingHandler.AddLike).Methods("POST")
	userRouter.HandleFunc("/api/removelike", listingHandler.RemoveLike).Methods("POST")

	// Маршруты для всех пользователей
	allUserRouter := router.NewRoute().Subrouter()
	allUserRouter.Use(middlewareHandler.CheckSesWithNilOnError)
	allUserRouter.HandleFunc("/api/listings", listingHandler.GetAllListings).Methods("GET")

	// Маршруты для статических страниц
	router.HandleFunc("/", handlers.OutIndex)
	router.HandleFunc("/register", handlers.OutRegister)
	router.HandleFunc("/login", handlers.OutLogin)
	router.HandleFunc("/listing", handlers.OutListing)
	router.HandleFunc("/edit", handlers.OutEdit)

	return router
}
