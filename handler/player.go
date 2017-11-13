package handler

import (
	"context"

	"github.com/eleme/purchaseMeiTuan/services/player"
)

// The code is used to ensure import safety, delete it as you wish.
var _ = player.GoUnusedProtection__

// PlayerService implements player.PlayerService interface.
type PlayerService struct{}

// NewPlayerService creates a new PlayerService.
func NewPlayerService() *PlayerService {
	return &PlayerService{}
}

// Ping is a handler for thrift service.
func (h *PlayerService) Ping(ctx context.Context) (bool, error) {
	return true, nil
}

// AssignTanks is a handler for thrift service.
func (h *PlayerService) AssignTanks(ctx context.Context, tanks []int32) error {
	return nil
}

// GetNewOrders is a handler for thrift service.
func (h *PlayerService) GetNewOrders(ctx context.Context) ([]*player.Order, error) {
	return nil, nil
}

// LatestState is a handler for thrift service.
func (h *PlayerService) LatestState(ctx context.Context, state *player.GameState) error {
	return nil
}

// UploadMap is a handler for thrift service.
func (h *PlayerService) UploadMap(ctx context.Context, gamemap [][]int32) error {
	return nil
}

// UploadParamters is a handler for thrift service.
func (h *PlayerService) UploadParamters(ctx context.Context, arguments *player.Args_) error {
	return nil
}
