package grpc

import (
	"context"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetByID(ctx context.Context, request *GeyByIDRequest) (*GeyByIDResponse, error) {
	return &GeyByIDResponse{
		User: &User{
			Id:   123,
			Name: "from " + s.Name,
		},
	}, nil
}
