package main

import (
	"context"
	"embed"
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
	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatal("Error creating validator", zap.Error(err))
	}

	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			protovalidate_middleware.UnaryServerInterceptor(validator),
		))
	// proto.RegisterUserServiceServer(s, service.NewUserServiceServer(dbpool))
	proto.RegisterExpenseServiceServer(s, service.NewExpensesServiceServer(dbpool))
	logger.Info("Server started on port 8090")
	if err := s.Serve(listener); err != nil {
		logger.Fatal("Error serving server", zap.Error(err))
	}
}

func RunGatewayServer(wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := proto.RegisterExpenseServiceHandlerFromEndpoint(ctx, mux, ":8090", opts)
	if err != nil {
		logger.Fatal("Error starting gateway server", zap.Error(err))
	}

	logger.Info("Gateway server started on port 8080")
	if err = http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("Error serving gateway server", zap.Error(err))
	}
}
