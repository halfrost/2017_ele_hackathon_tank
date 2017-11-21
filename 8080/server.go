package main

import (
	"astar"
	"fmt"
	"log"
	"math/rand"

	"github.com/eleme/purchaseMeiTuan/player"

	"git.apache.org/thrift.git/lib/go/thrift"
)

var gameArguments player.Args_
var gameMap [30][30]int32
var myTankList [5]int32
var myTankTypeList [5]int32
var enemyTankList [5]int32
var gameState player.GameState
var roundCount int32 = -1 // 回合数，初始值为 - 1
var gameStates []*player.GameState
var gameMapWidth int

// PlayerService struct
type PlayerService struct{}

// Ping is a handler for thrift service.
func (p *PlayerService) Ping() (bool, error) {
	return true, nil
}

// UploadParamters is a handler for thrift service.
// 接收初始参数,把参数存储到本地
func (p *PlayerService) UploadParamters(arguments *player.Args_) error {
	gameArguments = player.Args_{}
	gameArguments.TankSpeed = arguments.TankSpeed
	gameArguments.ShellSpeed = arguments.ShellSpeed
	gameArguments.TankHP = arguments.TankHP
	gameArguments.TankScore = arguments.TankScore
	gameArguments.FlagScore = arguments.FlagScore
	gameArguments.MaxRound = arguments.MaxRound
	gameArguments.RoundTimeoutInMs = arguments.RoundTimeoutInMs
	return nil
}

// UploadMap is a handler for thrift service.
// 接收二维地图，存储地图到本地
func (p *PlayerService) UploadMap(gamemap [][]int32) error {
	for i := 0; i < 30; i++ {
		for j := 0; j < 30; j++ {
			gameMap[i][j] = -1
		}
	}

	gameMapWidth = len(gameMap)/2 + 1
	for i := 0; i < len(gamemap); i++ {
		for j := 0; j < len(gamemap[i]); j++ {
			gameMap[i][j] = gamemap[i][j]
		}
	}
	return nil
}

// AssignTanks is a handler for thrift service.
// 接收己方坦克list，保存到本地
func (p *PlayerService) AssignTanks(tanks []int32) error {
	for i := 0; i < 5; i++ {
		myTankList[i] = -1
	}
	for i := 0; i < len(tanks); i++ {
		myTankList[i] = tanks[i]
	}
	return nil
}

// LatestState is a handler for thrift service.
// 获取最新的状态
func (p *PlayerService) LatestState(state *player.GameState) error {
	roundCount++
	gameStates = make([]*player.GameState, 200)
	gameState = player.GameState{}
	gameState.Tanks = state.Tanks
	gameState.Shells = state.Shells
	gameState.YourFlagNo = state.YourFlagNo
	gameState.EnemyFlagNo = state.EnemyFlagNo
	gameState.FlagPos = state.FlagPos

	gameStates[roundCount] = state
	return nil
}

// GetNewOrders is a handler for thrift service.
// 给己方坦克下达指令
func (p *PlayerService) GetNewOrders() ([]*player.Order, error) {
	orders := []*player.Order{}
	if roundCount%4 == 0 {
		order := &player.Order{TankId: myTankList[rand.Intn(len(gameState.Tanks))], Order: "fire", Dir: player.Direction_DOWN}
		orders = append(orders, order)
		return orders, nil
	}
	selectTank := gameState.Tanks[rand.Intn(len(gameState.Tanks))]
	order := moveOrder(selectTank.Pos, &player.Position{X: (int32)(gameMapWidth), Y: (int32)(gameMapWidth)}, myTankList[(rand.Int31n(4))], selectTank.Dir)
	orders = append(orders, order)
	return orders, nil
}

func moveOrder(tankPos, desPos *player.Position, tankID int32, tankDir player.Direction) (order *player.Order) {
	world := astar.InitWorld(gameMap)
	p, _, found := astar.Path(world.Start((int)(tankPos.X), (int)(tankPos.Y)), world.End((int)(desPos.X), (int)(desPos.Y)))
	if !found {
		return &player.Order{TankId: tankID, Order: "fire", Dir: tankDir}
	}
	pT := p[0].(*astar.Tile)
	var nextStep *astar.Tile
	if (((int32)(pT.X)) == tankPos.X) && (((int32)(pT.Y)) == tankPos.Y) {
		nextStep = p[1].(*astar.Tile)
	} else {
		nextStep = p[len(p)-2].(*astar.Tile)
	}
	isEqual, dir := getDir(tankPos, nextStep)
	if isEqual == true {
		return &player.Order{TankId: tankID, Order: "move", Dir: dir}
	}
	return &player.Order{TankId: tankID, Order: "turnTo", Dir: dir}
}

func getDir(tankPos *player.Position, nextStep *astar.Tile) (isEqual bool, dir player.Direction) {
	if (int32)(nextStep.X) == tankPos.X {
		isEqual = false
		if (int32)(nextStep.Y) > tankPos.Y {
			dir = player.Direction_RIGHT
		} else {
			dir = player.Direction_LEFT
		}
	} else {
		isEqual = true
		if (int32)(nextStep.X) > tankPos.X {
			dir = player.Direction_DOWN
		} else {
			dir = player.Direction_UP
		}
	}
	return isEqual, dir
}

const (
	// HOST host
	HOST = "localhost"
	// PORT post
	PORT = "8080"
)

func main() {

	handler := &PlayerService{}
	processor := player.NewPlayerServiceProcessor(handler)
	serverTransport, err := thrift.NewTServerSocket(HOST + ":" + PORT)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	// transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	// protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	server := thrift.NewTSimpleServer2(processor, serverTransport)
	fmt.Println("Running at:", HOST+":"+PORT)
	server.Serve()
}
