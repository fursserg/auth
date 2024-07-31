package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fursserg/auth/db"
	user "github.com/fursserg/auth/pkg/user_v1"
)

const grpcPort = 50051

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

type server struct {
	user.UnimplementedUserV1Server
}

// Create Создает нового юзера
func (s *server) Create(ctx context.Context, req *user.CreateRequest) (*user.CreateResponse, error) {
	pass := sha256.Sum256([]byte(req.GetPassword()))

	builderInsert := sq.Insert("users").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "password", "role", "status").
		Values(req.GetName(), req.GetEmail(), fmt.Sprintf("%x", pass), req.GetRole(), db.ActiveStatus).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	var userID int64
	err = db.GetConnect().QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	return &user.CreateResponse{
		Id: userID,
	}, nil
}

// Update Обновляет юзера
func (s *server) Update(ctx context.Context, req *user.UpdateRequest) (*emptypb.Empty, error) {
	hasChanges := false

	builderUpdate := sq.Update("users").
		PlaceholderFormat(sq.Dollar)

	if hasChanges, builderUpdate = s.hasChanges(req, builderUpdate); hasChanges {
		builderUpdate = builderUpdate.Set("updated_at", time.Now()).
			Where(sq.Eq{"id": req.GetId()})

		query, args, err := builderUpdate.ToSql()

		if err != nil {
			log.Fatalf("failed to build query: %v", err)
		}

		_, err = db.GetConnect().Exec(ctx, query, args...)
		if err != nil {
			log.Fatalf("failed to update user: %v", err)
		}
	}

	return new(emptypb.Empty), nil
}

// Delete Переводит юзера в статус "удален"
func (s *server) Delete(ctx context.Context, req *user.DeleteRequest) (*emptypb.Empty, error) {
	// Вместо удаления, переводим в специальный статус (храним в БД для истории)
	builderUpdate := sq.Update("users").
		PlaceholderFormat(sq.Dollar).
		Set("status", db.DeletedStatus).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.GetId()})

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	_, err = db.GetConnect().Exec(ctx, query, args...)
	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}

	return new(emptypb.Empty), nil
}

// Get получает одного юзера
func (s *server) Get(ctx context.Context, req *user.GetRequest) (*user.GetResponse, error) {
	builderInsert := sq.Select("id", "name", "email", "role", "created_at", "updated_at").
		PlaceholderFormat(sq.Dollar).
		From("users").
		Where(sq.Eq{"id": req.GetId()}).
		Limit(1)

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	var (
		id, role    int64
		name, email string
		createdAt   time.Time
		updatedAt   sql.NullTime
	)

	err = db.GetConnect().QueryRow(ctx, query, args...).Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
	if err != nil {
		log.Fatalf("failed to select user: %v", err)
	}

	return &user.GetResponse{
		Id:        id,
		Name:      name,
		Email:     email,
		Role:      user.Role(role),
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt.Time),
	}, nil
}

func (s *server) hasChanges(req *user.UpdateRequest, builder sq.UpdateBuilder) (bool, sq.UpdateBuilder) {
	hasChanges := false

	if req.GetName() != nil {
		builder = builder.Set("name", req.GetName().GetValue())
		hasChanges = true
	}

	if req.GetEmail() != nil {
		builder = builder.Set("email", req.GetEmail().GetValue())
		hasChanges = true
	}

	if req.GetRole() != user.Role_UNKNOWN {
		builder = builder.Set("role", req.GetRole())
		hasChanges = true
	}

	return hasChanges, builder
}
