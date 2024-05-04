package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var logger *zap.Logger

//go:embed db/migrations/*.sql
var embedMigrations embed.FS

func RegisterGrpcServers(s *grpc.Server, dbpool *pgxpool.Pool) {
	// TODO: Implement
}

func RegisterHttpHandlers(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) {
	// TODO: Implement
}

func main() {
	logger = zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") != "production" {
		logger = zap.Must(zap.NewDevelopment())
	}
	defer logger.Sync()

	logger.Info("Starting database connection pool")
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("Error creating database connection pool", zap.Error(err))
	}
	defer dbpool.Close()

	logger.Info("Connection pool started. Preparing to run migrations")
	goose.SetLogger(GooseZapLogger(logger))
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatal("Error setting dialect", zap.Error(err))
	}
	dbconn := stdlib.OpenDBFromPool(dbpool)
	if err := goose.Up(dbconn, "db/migrations"); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go RunGrpcServer(dbpool, &wg)
	go RunGatewayServer(&wg)

	wg.Wait()
	logger.Info("Exiting")
}

func RunGrpcServer(dbpool *pgxpool.Pool, wg *sync.WaitGroup) {
	defer wg.Done()
	serverPort := getEnvOrDefault("GRPC_SERVER_PORT", "8090")

	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatal("Error creating validator", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			protovalidate_middleware.UnaryServerInterceptor(validator),
		))

	RegisterGrpcServers(s, dbpool)

	logger.Infof("Server started on port %s", serverPort)
	if err := s.Serve(listener); err != nil {
		logger.Fatal("Error serving server", zap.Error(err))
	}
}

func RunGatewayServer(wg *sync.WaitGroup) {
	servicePort := getEnvOrDefault("HTTP_SERVER_PORT", "8080")

	defer wg.Done()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	RegisterHttpHandlers(ctx, mux, opts)

	logger.Infof("Gateway server started on port % s", servicePort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", servicePort), mux); err != nil {
		logger.Fatal("Error serving gateway server", zap.Error(err))
	}
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
