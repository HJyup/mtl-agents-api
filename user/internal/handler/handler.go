package handler

import (
	"context"
	pb "github.com/HJyup/mtl-common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateUser(ctx context.Context, p *pb.CreateUserRequest) (*pb.CreateUserResponse, error)
	AuthUser(ctx context.Context, p *pb.AuthUserRequest) (*pb.AuthUserResponse, error)
	GetUser(ctx context.Context, p *pb.GetUserRequest) (*pb.GetUserResponse, error)
	DeleteUser(ctx context.Context, p *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error)
}

type Handler struct {
	pb.UnimplementedUserServiceServer

	service Service
}

func NewHandler(grpcServer *grpc.Server, service Service) *Handler {
	handler := &Handler{service: service}

	pb.RegisterUserServiceServer(grpcServer, handler)
	return handler
}

func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	resp, err := h.service.CreateUser(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	return resp, nil
}

func (h *Handler) AuthUser(ctx context.Context, req *pb.AuthUserRequest) (*pb.AuthUserResponse, error) {
	resp, err := h.service.AuthUser(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to auth user: %v", err)
	}
	return resp, nil
}

func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	resp, err := h.service.GetUser(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	return resp, nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	resp, err := h.service.DeleteUser(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return resp, nil
}
