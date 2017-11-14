package models

import (
	"github.com/eleme/purchaseMeiTuan/services/player"
)

// TankModel struct defines
type TankModel struct {
	ID        int32
	Position  player.Position
	Direction player.Direction
	Hp        int32
	Type      int32
}
