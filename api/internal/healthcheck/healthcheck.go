package healthcheck

import (
	"api/internal/logger"
	"api/internal/messages"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// GrpcHealthChecker реализует проверку состояния gRPC соединений
type GrpcHealthChecker struct {
	connections map[string]*grpc.ClientConn // Имя соединения -> соединение
	mu          sync.RWMutex                // Мьютекс для безопасного доступа к соединениям
	interval    time.Duration               // Интервал проверки состояния соединений
	cancel      context.CancelFunc          // Функция для отмены контекста
}

func NewHealthChecker(checkInterval time.Duration) *GrpcHealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	checker := &GrpcHealthChecker{
		connections: make(map[string]*grpc.ClientConn),
		interval:    checkInterval,
		cancel:      cancel,
	}
	go checker.startChecking(ctx)
	return checker
}

// Stop останавливает проверку соединений
func (c *GrpcHealthChecker) Stop() {
	log.Println("Stopping gRPC health checker...")
	if c.cancel != nil {
		c.cancel()
	}
}

// AddConnection добавляет новое gRPC соединение для проверки состояния
func (c *GrpcHealthChecker) AddConnection(name string, conn *grpc.ClientConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections[name] = conn
}

// RemoveConnection удаляет gRPC соединение из проверки состояния
func (c *GrpcHealthChecker) startChecking(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkConnections()
		}
	}
}

// checkConnections проверяет состояние всех gRPC соединений
// Если соединение не в состоянии Ready, оно будет перезапущено
func (c *GrpcHealthChecker) checkConnections() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for name, conn := range c.connections {
		state := conn.GetState()
		if state != connectivity.Ready {
			logger.Info(messages.ServiceHealthcheck, fmt.Sprintf(messages.StatusHealth, name, state), nil)
			conn.Connect()
		}
	}
}
