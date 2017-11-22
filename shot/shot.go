package shot

// 0-1-2-3
// 上下左右查找
// 1. 若有敌方坦克，距离每增加一格减少 1，初始 10；若无则 0；
// 2. 若有己方坦克，距离每增加一格增加 1，初始为 -10；若无则 0；
// 3. 若有目标草丛，距离每增加一格减少 0.5，初始 5；若无则 0；

func shot (x int, y int, width int, height int, enemyTankList []Position, myTankTypeList []Position, grass: Position) Direction {
	var scoreArr [4]int32
	for i := 0; i < len(scoreArr); ++i {
		for step := 1; step < 10; ++step {
			var realX int = x
			var realY int = y

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
			if (realX >= width || realY >= height || realX < 0 || realY < 0) {
				break
			}

			// 1.
			var boolEnemy bool = false
			for i := 0; i < len(enemyTankList); ++i {
				if (enemyTankList[i].x == realX && enemyTankList[i].y == realY) {
					boolEnemy = true
					break
				}
			}
			if (boolEnemy) {
				scoreArr[i] = scoreArr[i] + (10 - step * 1)
			}

			// 2.
			var boolMy bool = false
			for i := 0; i < len(myTankTypeList); ++i {
				if (myTankTypeList[i].x == realX && myTankTypeList[i].y == realY) {
					boolMy = true
					break
				}
			}
			if (boolMy) {
				scoreArr[i] = scoreArr[i] + (-10 + step * 1)
			}

			// 3.
			if (grass.x == realX && grass.y == realY) {
				scoreArr[i] = scoreArr[i] + (5 - step * 0.5)
			}
		}
	}

	// 取最高分方向开炮
	var index int = 0
	var value int32 = scoreArr[0]
	for i := 1; i < len(scoreArr); ++i {
		if (scoreArr[i] > value) {
			value = scoreArr[i]
			index = i
		}
	}
    return (index + 1)
}
