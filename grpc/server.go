package grpc

import "context"

type Server struct {
	UnimplementedUserServiceServer
}

func (s *Server) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	return &GetByIDResponse{
		User: &User{
			Id:   123,
			Name: "zl",
		},
	}, nil
}
