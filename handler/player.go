package handler

import (
	"context"

	"github.com/eleme/purchaseMeiTuan/services/player"
)

// The code is used to ensure import safety, delete it as you wish.
var _ = player.GoUnusedProtection__

// playerService implements player.playerService interface.
type playerService struct{}

// NewplayerService creates a new playerService.
func NewplayerService() *playerService {
	return &playerService{}
}

// AddTodo is a handler for thrift service.
func (h *playerService) AddTodo(ctx context.Context, title string) error {
	// TODO
	return nil
}

// CompleteTodo is a handler for thrift service.
func (h *playerService) CompleteTodo(ctx context.Context, id int64) error {
	// TODO
	return nil
}

// ListTodo is a handler for thrift service.
func (h *playerService) ListTodo(ctx context.Context) ([]*player.TTodo, error) {
	// TODO
	var v []*player.TTodo
	return v, nil
}

// Nay is a handler for thrift service.
func (h *playerService) Nay(ctx context.Context, id int32) (bool, error) {
	// TODO
	var v bool
	return v, nil
}

// Ping is a handler for thrift service.
func (h *playerService) Ping(ctx context.Context) (bool, error) {
	// TODO
	var v bool
	return v, nil
}

// Yay is a handler for thrift service.
func (h *playerService) Yay(ctx context.Context, id int32) (bool, error) {
	// TODO
	var v bool
	return v, nil
}
