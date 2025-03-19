package main

import (
	"context"
	"fmt"
	"github.com/HJyup/mlt-user/internal/handler"
	"github.com/HJyup/mlt-user/internal/service"
	"github.com/HJyup/mlt-user/internal/store"
	common "github.com/HJyup/mtl-common"
	"github.com/HJyup/mtl-common/consul"
	"github.com/jackc/pgx/v5"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Specification struct {
	ServiceName string `required:"true" default:"user"`
	Address     string `required:"true"`
	Consul      string `required:"true"`
	Environment string `required:"true"`
	DBUser      string `required:"true"`
	DBPassword  string `required:"true"`
	DBName      string `required:"true"`
	DBHost      string `required:"true"`
	DBPort      string `required:"true"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()

	var s Specification
	if err = envconfig.Process("user", &s); err != nil {
		logger.Fatal("Failed to process environment variables", zap.Error(err))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", s.DBUser, s.DBPassword, s.DBName, s.DBPort, s.DBHost)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		logger.Fatal("Failed to parse database config: %v", zap.Error(err))
	}

	dbConn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		logger.Fatal("Failed to connect to the database: %v", zap.Error(err))
	}
	defer func(dbConn *pgx.Conn, ctx context.Context) {
		err = dbConn.Close(ctx)
		if err != nil {
			logger.Fatal("Failed to close database connection: %v", zap.Error(err))
		}
	}(dbConn, ctx)

	registry, err := consul.NewRegistry(s.Consul)
	if err != nil {
		logger.Fatal("Failed to create registry: %v", zap.Error(err))
	}

	instanceID := common.GenerateInstanceID(s.ServiceName)
	if err = registry.Register(instanceID, s.ServiceName, s.Address); err != nil {
		logger.Fatal("Failed to register service: %v", zap.Error(err))
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err = registry.HealthCheck(instanceID); err != nil {
					logger.Error("Failed to health check: %v", zap.Error(err))
				}
			}
		}
	}()
	defer registry.DeRegister(instanceID)

	grpcServer := grpc.NewServer()
	conn, err := net.Listen("tcp", s.Address)
	if err != nil {
		logger.Fatal("Failed to listen on %s: %v", zap.String("port", s.Address), zap.Error(err))
	}
	defer func(conn net.Listener) {
		err = conn.Close()
		if err != nil {
			logger.Warn("Failed to close connection", zap.Error(err))
		}
	}(conn)

	str := store.NewStore(dbConn)
	srv := service.NewService(str, logger)
	handler.NewHandler(grpcServer, srv)

	logger.Info("Starting HTTP server", zap.String("port", s.Address))
	if err = grpcServer.Serve(conn); err != nil {
		logger.Fatal("Failed to serve gRPC: %v", zap.Error(err))
	}
}
