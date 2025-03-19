package service

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/HJyup/mtl-common/api"
	"github.com/HJyup/mtl-common/utils"
	"go.uber.org/zap"
)

var (
	ErrEmptyValues = errors.New("empty values")
	ErrEmptyUserID = errors.New("user id is empty")
)

type Store interface {
	CreateUser(ctx context.Context, username, email, password string) (string, error)
	AuthUser(ctx context.Context, email, password string) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	DeleteUser(ctx context.Context, id string) error
}

type Service struct {
	store  Store
	logger *zap.Logger
}

func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{store: store, logger: logger}
}

func (svc *Service) CreateUser(ctx context.Context, p *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if p == nil || p.GetUsername() == "" || p.GetEmail() == "" || p.GetPassword() == "" {
		return nil, ErrEmptyValues
	}

	userID, err := svc.store.CreateUser(ctx, p.Username, p.Email, p.Password)
	if err != nil {
		svc.logger.Error("failed to create user",
			zap.String("username", p.Username),
			zap.String("email", p.Email),
			zap.Error(err))
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &pb.CreateUserResponse{
		UserId: userID,
	}, nil
}

func (svc *Service) AuthUser(ctx context.Context, p *pb.AuthUserRequest) (*pb.AuthUserResponse, error) {
	if p == nil || p.GetEmail() == "" || p.GetPassword() == "" {
		return nil, ErrEmptyValues
	}

	user, err := svc.store.AuthUser(ctx, p.Email, p.Password)
	if err != nil {
		svc.logger.Error("failed to auth user",
			zap.String("email", p.Email),
			zap.Error(err))
		return nil, fmt.Errorf("authenticate user: %w", err)
	}

	token, err := utils.CreateToken(user.ID, p.Email, user.Username)
	if err != nil {
		svc.logger.Error("failed to create token",
			zap.String("user_id", user.ID),
			zap.Error(err))
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &pb.AuthUserResponse{
		Token: token,
	}, nil
}

func (svc *Service) GetUser(ctx context.Context, p *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if p == nil || p.GetUserId() == "" {
		return nil, ErrEmptyUserID
	}

	user, err := svc.store.GetUser(ctx, p.GetUserId())
	if err != nil {
		svc.logger.Warn("failed to get user",
			zap.String("user_id", p.GetUserId()),
			zap.Error(err))
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &pb.GetUserResponse{
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (svc *Service) DeleteUser(ctx context.Context, p *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if p == nil || p.GetUserId() == "" {
		return nil, ErrEmptyUserID
	}

	err := svc.store.DeleteUser(ctx, p.GetUserId())
	if err != nil {
		svc.logger.Warn("failed to delete user",
			zap.String("user_id", p.GetUserId()),
			zap.Error(err))
		return nil, fmt.Errorf("delete user: %w", err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "user deleted",
	}, nil
}
