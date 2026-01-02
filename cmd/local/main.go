package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// Input Service - Adapters In
	httpAdapter "jiaa-server-core/internal/input/adapter/in/http"
	kafkaIn "jiaa-server-core/internal/input/adapter/in/kafka"

	// Input Service - Adapters Out
	inputGrpcOut "jiaa-server-core/internal/input/adapter/out/grpc"
	kafkaOut "jiaa-server-core/internal/input/adapter/out/kafka"
	"jiaa-server-core/internal/input/adapter/out/memory"

	// Input Service - Services
	inputService "jiaa-server-core/internal/input/service"

	// Output Service - Adapters In
	outputGrpcIn "jiaa-server-core/internal/output/adapter/in/grpc"

	// Output Service - Adapters Out
	outputGrpcOut "jiaa-server-core/internal/output/adapter/out/grpc"

	// Output Service - Services
	outputService "jiaa-server-core/internal/output/service"
)

// Config 통합 서버 설정
type Config struct {
	// Input Service
	HTTPPort      string
	KafkaBrokers  string
	ActivityTopic string
	StateTopic    string

	// Output Service
	OutputGRPCPort string

	// External Services (Dev 1, Dev 3, Dev 5)
	PhysicalControlAddr string
	ScreenControlAddr   string
	IntelligenceAddr    string // Dev 5 (Intelligence Worker)
}

func main() {
	log.Println("============================================")
	log.Println("  JIAA Server Core - Local Development Mode")
	log.Println("============================================")

	config := loadConfig()

	// ========================================
	// Output Service 초기화 (먼저 시작 - gRPC 서버)
	// ========================================
	log.Println("[LOCAL] Initializing Output Service...")

	// Output - Driven Adapters (→ Dev 1, Dev 3)
	physicalExecutor := outputGrpcOut.NewPhysicalExecutorAdapterLazy(config.PhysicalControlAddr)
	screenExecutor := outputGrpcOut.NewScreenExecutorAdapterLazy(config.ScreenControlAddr)

	// Output - Service
	sabotageExecutorService := outputService.NewSabotageExecutorService(physicalExecutor, screenExecutor)

	// Output - Driving Adapter (gRPC Server)
	outputServer := outputGrpcIn.NewSabotageServer(config.OutputGRPCPort, sabotageExecutorService)

	// Start Output gRPC Server
	if err := outputServer.Start(); err != nil {
		log.Fatalf("[LOCAL] Failed to start Output gRPC server: %v", err)
	}
	log.Printf("[LOCAL] Output Service started on gRPC port %s", config.OutputGRPCPort)

	// ========================================
	// Input Service 초기화
	// ========================================
	log.Println("[LOCAL] Initializing Input Service...")

	// Input - Driven Adapters
	blacklistAdapter := memory.NewBlacklistAdapterWithDefaults()

	// Kafka Producer (optional)
	dataRelayAdapter, err := kafkaOut.NewDataRelayAdapter(config.KafkaBrokers, config.ActivityTopic)
	if err != nil {
		log.Printf("[LOCAL] Warning: Kafka producer not available: %v", err)
	}

	// gRPC Clients (→ Output Service, Dev 1, Dev 3, Dev 5)
	sabotageCommandAddr := "localhost:" + config.OutputGRPCPort // 같은 프로세스의 Output Service
	sabotageAdapter := inputGrpcOut.NewSabotageCommandAdapterLazy(sabotageCommandAddr)
	physicalAdapter := inputGrpcOut.NewPhysicalControlAdapterLazy(config.PhysicalControlAddr)
	screenAdapter := inputGrpcOut.NewScreenControlAdapterLazy(config.ScreenControlAddr)
	intelligenceAdapter := inputGrpcOut.NewIntelligenceAdapterLazy(config.IntelligenceAddr)

	// Input - Services
	reflexService := inputService.NewReflexService(blacklistAdapter, sabotageAdapter, dataRelayAdapter)
	commandRouterService := inputService.NewCommandRouterService(physicalAdapter, screenAdapter)
	solutionRouterService := inputService.NewSolutionRouterService(screenAdapter)

	// Intelligence Adapter 로깅 (Dev 5 연결 확인)
	log.Printf("[LOCAL] Intelligence Worker (Dev 5) address: %s", config.IntelligenceAddr)
	_ = intelligenceAdapter // TODO: Emergency 프로토콜에서 사용 예정

	// Input - Driving Adapters
	activityHandler := httpAdapter.NewActivityHandler(reflexService)

	// Kafka Consumer (optional)
	var stateConsumer *kafkaIn.StateConsumer
	if config.KafkaBrokers != "" {
		stateConsumer, err = kafkaIn.NewStateConsumer(
			config.KafkaBrokers,
			"core-decision-service-local",
			config.StateTopic,
			commandRouterService,
		)
		if err != nil {
			log.Printf("[LOCAL] Warning: Kafka consumer not available: %v", err)
		} else if err := stateConsumer.Start(); err != nil {
			log.Printf("[LOCAL] Warning: Failed to start Kafka consumer: %v", err)
		}
	}

	// ========================================
	// HTTP Server (Echo)
	// ========================================
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	activityHandler.RegisterRoutes(e)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":         "ok",
			"mode":           "local",
			"input_service":  "running",
			"output_service": "running",
		})
	})

	// Solution Router (Dev 5 → Dev 3)
	e.POST("/api/v1/solution", func(c echo.Context) error {
		var req struct {
			ClientID string `json:"client_id"`
			Markdown string `json:"markdown"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid request"})
		}
		if err := solutionRouterService.RouteAIResult(req.ClientID, req.Markdown); err != nil {
			return c.JSON(500, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Start HTTP Server
	go func() {
		log.Printf("[LOCAL] Input Service started on HTTP port %s", config.HTTPPort)
		if err := e.Start(":" + config.HTTPPort); err != nil {
			log.Printf("[LOCAL] HTTP server stopped: %v", err)
		}
	}()

	// ========================================
	// Startup Complete
	// ========================================
	log.Println("============================================")
	log.Printf("  HTTP API: http://localhost:%s", config.HTTPPort)
	log.Printf("  gRPC:     localhost:%s", config.OutputGRPCPort)
	log.Println("  Press Ctrl+C to stop")
	log.Println("============================================")

	// ========================================
	// Graceful Shutdown
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[LOCAL] Shutting down...")

	// Cleanup
	if stateConsumer != nil {
		stateConsumer.Stop()
	}
	if dataRelayAdapter != nil {
		dataRelayAdapter.Close()
	}
	outputServer.Stop()
	sabotageAdapter.Close()
	physicalAdapter.Close()
	screenAdapter.Close()
	physicalExecutor.Close()
	screenExecutor.Close()

	log.Println("[LOCAL] Shutdown complete")
}

// loadConfig 환경 변수에서 설정 로드
func loadConfig() Config {
	return Config{
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		KafkaBrokers:        getEnv("KAFKA_BROKERS", ""),
		ActivityTopic:       getEnv("ACTIVITY_TOPIC", "client-activity"),
		StateTopic:          getEnv("STATE_TOPIC", "command-state"),
		OutputGRPCPort:      getEnv("OUTPUT_GRPC_PORT", "50053"),
		PhysicalControlAddr: getEnv("PHYSICAL_CONTROL_ADDR", "localhost:50051"),
		ScreenControlAddr:   getEnv("SCREEN_CONTROL_ADDR", "localhost:50052"),
		IntelligenceAddr:    getEnv("INTELLIGENCE_ADDR", "localhost:50051"), // Dev 5
	}
}

// getEnv 환경 변수 조회 (기본값 지원)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
