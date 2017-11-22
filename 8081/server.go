package main

import (
	"astar"
	"fmt"
	"log"
	"math/rand"

	"github.com/eleme/purchaseMeiTuan/player"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	// HOST host
	HOST = "localhost"
	// PORT post
	PORT = "8081"
)

var gameArguments player.Args_
var gameMap [50][50]int32
var astarGameMap [50][50]int32
var nextSteps []*player.Position
var myTankList [5]int32
var myTankTypeList [5]int32
var myTankNum int
var enemyTankList [5]int32
var gameState player.GameState
var roundCount int32 = -1 // 回合数，初始值为 - 1
var gameStates []*player.GameState
var gameMapCenter int
var gameMapDiagonally int
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
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			gameMap[i][j] = -1
		}
	}

	gameMapCenter = len(gamemap) / 2
	gameMapWidth = len(gamemap) / 2
	for i := 0; i < len(gamemap); i++ {
		for j := 0; j < len(gamemap[i]); j++ {
			gameMap[i][j] = gamemap[i][j]
			astarGameMap[i][j] = gamemap[i][j]
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
	myTankNum = len(tanks)
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
	refeshTankState()
	orders := []*player.Order{}
	//fmt.Printf("第 %d 回合 | gameState = %v\n", roundCount, gameState)

	fmt.Printf("myTankNum = %d\n", myTankNum)

	nextSteps = make([]*player.Position, 0)
	for i := 0; i < myTankNum; i++ {
		pos, dir, _ := getTankPosDirHp(myTankList[i])
		if (roundCount%4 == 0) && (pos.X > 8) && (pos.Y > 8) {
			enemyTankPos, myTankPos := getTankListFromGameState()
			fireDir := shot((int)(pos.X), (int)(pos.Y), gameMapWidth, gameMapWidth, enemyTankPos, myTankPos, &player.Position{X: 0, Y: 0})
			order := &player.Order{TankId: myTankList[i], Order: "fire", Dir: fireDir}
			orders = append(orders, order)
			return orders, nil
		}

		// if roundCount%4 == 0 && pos.X > 10 && pos.Y > 10 {
		// 	order := &player.Order{TankId: myTankList[i], Order: "fire", Dir: player.Direction_DOWN}
		// 	orders = append(orders, order)
		// 	fmt.Printf("第 %d 回合 | 【8081】玩家攻击指令 = %v\n", roundCount, orders)
		// }
		// 取余为 0 的坦克
		if myTankList[i]%2 == 0 {
			if (int)(pos.X) == gameMapCenter && (int)(pos.Y) == gameMapCenter {
				order := moveOrder(pos, &player.Position{X: (int32)(gameMapCenter) + (int32)(rand.Intn(3)), Y: (int32)(gameMapCenter) + (int32)(rand.Intn(3))}, myTankList[i], dir)
				orders = append(orders, order)
			} else {
				order := moveOrder(pos, &player.Position{X: (int32)(gameMapCenter), Y: (int32)(gameMapCenter)}, myTankList[i], dir)
				orders = append(orders, order)
			}
		} else { // 取余为 1 的坦克

			order := moveOrder(pos, &player.Position{X: (int32)(gameMapCenter) + (int32)(rand.Intn(gameMapWidth)), Y: (int32)(gameMapCenter) + (int32)(rand.Intn(gameMapWidth))}, myTankList[i], dir)
			orders = append(orders, order)
		}
	}
	fmt.Printf("第 %d 回合 | 【8081】玩家移动指令 = %v \n", roundCount, orders)
	return orders, nil
}

func getTankPosDirHp(tankID int32) (pos *player.Position, dir player.Direction, hp int32) {
	for i := 0; i < len(gameState.Tanks); i++ {
		if gameState.Tanks[i].ID == tankID {
			return gameState.Tanks[i].Pos, gameState.Tanks[i].Dir, gameState.Tanks[i].Hp
		}
	}
	return &player.Position{X: 0, Y: 0}, 0, 0
}

// refeshTankState 刷新己方坦克 list 数组，并且刷新己方 myTankNum
func refeshTankState() {
	for i := 0; i < len(myTankList); i++ {
		isExist := false
		for j := 0; j < len(gameState.Tanks); j++ {
			if myTankList[i] == gameState.Tanks[j].ID {
				isExist = true
			}
		}
		if isExist == false {
			for k := i; k < len(myTankList)-1; k++ {
				myTankList[k] = myTankList[k+1]
			}
		}
	}
	myTankNum = 0
	for count := 0; count < len(myTankList); count++ {
		if myTankList[count] != -1 {
			myTankNum++
		}
	}
}

// refeshAStarMap 刷新 astar 地图
func refeshAStarMap() {
	for i := 0; i < len(gameMap); i++ {
		for j := 0; j < len(gameMap[i]); j++ {
			astarGameMap[i][j] = gameMap[i][j]
		}
	}
	for i := 0; i < len(gameState.Tanks); i++ {
		astarGameMap[gameState.Tanks[i].Pos.X][gameState.Tanks[i].Pos.Y] = 1
	}
	for i := 0; i < len(gameState.Shells); i++ {

		astarGameMap[gameState.Shells[i].Pos.X][gameState.Shells[i].Pos.Y] = 1

		switch gameState.Shells[i].Dir {
		case player.Direction_UP:
			{
				for j := 1; j <= (int)(gameArguments.ShellSpeed); j++ {
					astarGameMap[gameState.Shells[i].Pos.X][(int)(gameState.Shells[i].Pos.Y)-j] = 1
				}
			}

		case player.Direction_DOWN:
			{
				for j := 1; j <= (int)(gameArguments.ShellSpeed); j++ {
					astarGameMap[gameState.Shells[i].Pos.X][(int)(gameState.Shells[i].Pos.Y)+j] = 1
				}
			}

		case player.Direction_LEFT:
			{
				for j := 1; j <= (int)(gameArguments.ShellSpeed); j++ {
					astarGameMap[(int)(gameState.Shells[i].Pos.X)-j][(int)(gameState.Shells[i].Pos.Y)] = 1
				}
			}

		case player.Direction_RIGHT:
			{
				for j := 1; j <= (int)(gameArguments.ShellSpeed); j++ {
					astarGameMap[(int)(gameState.Shells[i].Pos.X)+j][(int)(gameState.Shells[i].Pos.Y)] = 1
				}
			}
		}

	}
}

func moveOrder(tankPos, desPos *player.Position, tankID int32, tankDir player.Direction) (order *player.Order) {

	refeshAStarMap()
	world := astar.InitWorld(astarGameMap)
	p, _, found := astar.Path(world.Start((int)(tankPos.X), (int)(tankPos.Y)), world.End((int)(desPos.X), (int)(desPos.Y)))
	if !found {
		fmt.Printf("333333 tankID = %d\n", tankID)
		return &player.Order{TankId: tankID, Order: "turnTo", Dir: tankDir}
	}
	pT := p[0].(*astar.Tile)
	fmt.Print("Resulting path = \n", world.RenderPath(p))
	var nextStep *astar.Tile
	if (((int32)(pT.X)) == tankPos.X) && (((int32)(pT.Y)) == tankPos.Y) {
		nextStep = p[1].(*astar.Tile)
	} else {
		nextStep = p[len(p)-2].(*astar.Tile)
	}

	fmt.Printf("nextStep = %v | X = %d | Y = %d | 当前tank pos.x = %d | posY = %d\n", nextStep.Kind, nextStep.X, nextStep.Y, tankPos.X, tankPos.Y)
	isEqual, dir := getDir(tankPos, nextStep, tankDir)
	if isEqual == true {
		if len(nextSteps) == 0 {
			nextSteps = append(nextSteps, &player.Position{X: (int32)(nextStep.X), Y: (int32)(nextStep.Y)})
			fmt.Printf("【nextSteps】= %v\n", nextSteps)
		} else {
			nextStepCount := len(nextSteps)
			for i := 0; i < nextStepCount; i++ {
				if nextSteps[i].X == (int32)(nextStep.X) && nextSteps[i].Y == (int32)(nextStep.Y) {
					fmt.Printf("nextStep = %d,%d | nextSteps = %v\n", nextStep.X, nextStep.Y, nextSteps)
					fmt.Printf("222222 tankID = %d\n", tankID)
					return &player.Order{TankId: tankID, Order: "turnTo", Dir: dir}
				}
				fmt.Printf("****添加前nextSteps= %v | 要添加的 nextStep X = %d Y = %d | 方向 = %d\n", nextSteps, nextStep.X, nextStep.Y, tankDir)
				nextSteps = append(nextSteps, &player.Position{X: (int32)(nextStep.X), Y: (int32)(nextStep.Y)})
				fmt.Printf("****添加后nextSteps= %v\n", nextSteps)
			}
		}
		return &player.Order{TankId: tankID, Order: "move", Dir: dir}
	}
	fmt.Printf("111111 tanDir = %d dir = %d\n", tankDir, dir)
	return &player.Order{TankId: tankID, Order: "turnTo", Dir: dir}
}

func getDir(tankPos *player.Position, nextStep *astar.Tile, tankDir player.Direction) (isEqual bool, dir player.Direction) {
	if (int32)(nextStep.X) == tankPos.X {
		if (int32)(nextStep.Y) > tankPos.Y {
			dir = player.Direction_RIGHT
		} else {
			dir = player.Direction_LEFT
		}
	} else {
		if (int32)(nextStep.X) > tankPos.X {
			dir = player.Direction_DOWN
		} else {
			dir = player.Direction_UP
		}
	}

	return tankDir == dir, dir
}

func getTankListFromGameState() (enemyTankPos, myTankPos []*player.Position) {
	for i := 0; i < myTankNum; i++ {
		for j := 0; j < len(gameState.Tanks); j++ {
			if gameState.Tanks[j].ID == myTankList[i] {
				myTankPos = append(myTankPos, gameState.Tanks[j].Pos)
			} else {
				enemyTankPos = append(enemyTankPos, gameState.Tanks[j].Pos)
			}
		}
	}
	return enemyTankPos, myTankPos
}

// 0-1-2-3
// 上下左右查找
// 1. 若有敌方坦克，距离每增加一格减少 1，初始 10；若无则 0；
// 2. 若有己方坦克，距离每增加一格增加 1，初始为 -10；若无则 0；
// 3. 若有目标草丛，距离每增加一格减少 0.5，初始 5；若无则 0；

func shot(x int, y int, width int, height int, enemyTankList []*player.Position, myTankList []*player.Position, grass *player.Position) player.Direction {
	var scoreArr [4]int

	for i := 0; i < len(scoreArr); i++ {
		for step := 1; step < 10; step++ {
			var realX = x
			var realY = y

			switch i {
			case 0:
				realY = y - step
				break
			case 1:
				realY = y + step
				break
			case 2:
				realX = x - step
				break
			case 3:
				realX = x + step
				break
			}
			// 越界跳出循环
			if realX >= width || realY >= height || realX < 0 || realY < 0 {
				break
			}

			// 1.
			boolEnemy := false
			for j := 0; j < len(enemyTankList); j++ {
				if (enemyTankList[i].X == (int32)(realX)) && (enemyTankList[i].Y == (int32)(realY)) {
					boolEnemy = true
					break
				}
			}
			if boolEnemy {
				scoreArr[i] = scoreArr[i] + (10 - step*1)
			}

			// 2.
			boolMy := false
			for k := 0; k < len(myTankList); k++ {
				if (myTankList[i].X == (int32)(realX)) && (myTankList[i].Y == (int32)(realY)) {
					boolMy = true
					break
				}
			}
			if boolMy {
				scoreArr[i] = scoreArr[i] + (-10 + step*1)
			}

			// 3.
			if grass.X == (int32)(realX) && grass.Y == (int32)(realY) {
				scoreArr[i] = scoreArr[i] + (5 - (int)((float32)(step)*0.5))
			}
		}
	}

	// index 取最高分方向开炮
	index := 0
	value := scoreArr[0]
	for i := 1; i < len(scoreArr); i++ {
		if scoreArr[i] > value {
			value = scoreArr[i]
			index = i
		}
	}
	dirs := []player.Direction{player.Direction_UP, player.Direction_DOWN, player.Direction_LEFT, player.Direction_RIGHT}
	return dirs[index]
}

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
