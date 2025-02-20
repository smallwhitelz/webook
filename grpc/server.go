package grpc

import (
	"context"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetByID(ctx context.Context, request *GeyByIDRequest) (*GeyByIDResponse, error) {
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		log.Println(rest.String())
	}
	return &GeyByIDResponse{
		User: &User{
			Id:   123,
			Name: "from " + s.Name,
		},
	}, nil
}
