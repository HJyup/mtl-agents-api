package handler

import (
	"context"
	pb "github.com/HJyup/mtl-common/api"
)

type Service interface {
	CreateConfiguration(ctx context.Context, p *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error)
	UpdateConfiguration(ctx context.Context, p *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error)
	GetConfiguration(ctx context.Context, p *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error)
	DeleteConfiguration(ctx context.Context, p *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error)
}
