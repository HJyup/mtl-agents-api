package main

import (
	"context"
	"github.com/HJyup/mlt-configuration/internal/handler"
	"github.com/HJyup/mlt-configuration/internal/service"
	"github.com/HJyup/mlt-configuration/internal/store"
	common "github.com/HJyup/mtl-common"
	"github.com/HJyup/mtl-common/consul"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Specification struct {
	ServiceName   string `required:"true" default:"configuration"`
	Address       string `required:"true"`
	Consul        string `required:"true"`
	Environment   string `required:"true"`
	DBLink        string `required:"true"`
	EncryptionKey string `required:"true"`
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
	if err = envconfig.Process("configuration", &s); err != nil {
		logger.Fatal("Failed to process environment variables", zap.Error(err))
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(s.DBLink).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			logger.Fatal("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}()

	if err = client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		logger.Panic("Failed to ping MongoDB", zap.Error(err))
	}

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

	str := store.NewStore(client)
	srv, err := service.NewService(str, logger, s.EncryptionKey)
	if err != nil {
		logger.Fatal("Failed to create service", zap.Error(err))
	}
	handler.NewHandler(grpcServer, srv)

	logger.Info("Starting HTTP server", zap.String("port", s.Address))
	if err = grpcServer.Serve(conn); err != nil {
		logger.Fatal("Failed to serve gRPC: %v", zap.Error(err))
	}
}
