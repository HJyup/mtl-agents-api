package handler

import (
	"context"
	pb "github.com/HJyup/mtl-common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateConfiguration(ctx context.Context, p *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error)
	UpdateConfiguration(ctx context.Context, p *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error)
	GetConfiguration(ctx context.Context, p *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error)
	DeleteConfiguration(ctx context.Context, p *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error)
}

type Handler struct {
	pb.UnimplementedConfigurationServiceServer

	service Service
}

func NewHandler(grpcServer *grpc.Server, service Service) *Handler {
	handler := &Handler{service: service}

	pb.RegisterConfigurationServiceServer(grpcServer, handler)
	return handler
}

func (h *Handler) CreateConfiguration(ctx context.Context, req *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error) {
	resp, err := h.service.CreateConfiguration(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create configuration: %v", err)
	}
	return resp, nil
}

func (h *Handler) UpdateConfiguration(ctx context.Context, req *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error) {
	resp, err := h.service.UpdateConfiguration(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update configuration: %v", err)
	}
	return resp, nil
}

func (h *Handler) GetConfigurationByUserID(ctx context.Context, req *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	resp, err := h.service.GetConfiguration(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get configuration: %v", err)
	}
	return resp, nil
}

func (h *Handler) DeleteConfigurationByUserID(ctx context.Context, req *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error) {
	resp, err := h.service.DeleteConfiguration(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete configuration: %v", err)
	}
	return resp, nil
}
