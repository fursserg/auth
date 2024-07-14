package main

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"

	"github.com/brianvoe/gofakeit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	user "github.com/fursserg/auth/pkg/user_v1"
)

const grpcPort = 50051

type server struct {
	user.UnimplementedUserV1Server
}

func (s *server) Create(ctx context.Context, req *user.CreateRequest) (*user.CreateResponse, error) {
	log.Printf("User create data: %+v", req)

	return &user.CreateResponse{
		Id: gofakeit.Int64(),
	}, nil
}

func (s *server) Update(ctx context.Context, req *user.UpdateRequest) (*emptypb.Empty, error) {
	log.Printf("User update data: %+v", req)

	res := new(emptypb.Empty)
	return res, nil
}

func (s *server) Delete(ctx context.Context, req *user.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("User delete id: %d", req.GetId())

	res := new(emptypb.Empty)
	return res, nil
}

func (s *server) Get(ctx context.Context, req *user.GetRequest) (*user.GetResponse, error) {
	log.Printf("User get id: %d", req.GetId())

	return &user.GetResponse{
		Id:        req.GetId(),
		Name:      gofakeit.Name(),
		Email:     gofakeit.Email(),
		Role:      1,
		CreatedAt: timestamppb.New(gofakeit.Date()),
		UpdatedAt: timestamppb.New(gofakeit.Date()),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	user.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
