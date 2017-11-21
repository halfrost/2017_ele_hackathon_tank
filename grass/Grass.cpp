#include "Grass.h"
#include <stdlib.h>
#include <time.h>

#include <iostream>

using namespace std;

namespace grass {
  Position::Position() {}

  Position::Position(int x, int y) {
    xPos = x;
    yPos = y;
  }

  Range::Range(Position *s, Position *e) {
    start = s;
    end = e;
  }

  Tank::Tank(int id, Position *pos, Direction dir, int h) {
    tankdId = id;
    position = pos;
    direction = dir;
    hp = h;
  }

  Order::Order(int id, Position *cur, Position *nex, Position *dest) {
    tankId = id;
    current = cur;
    next = nex;
    destination = dest;
  }

  Route::Route(int len, list<Position *> pos) {
    length = len;
    positions = pos;
  }

  int RandomUtil::random(int from, int to) {
    return from + rand() % (to - from + 1);
  }

  // Mock A star
  Route *GrassService::route(Position *tankPos, Position *grassPos, int **gameMap) {
    int distance = RandomUtil::random(20, 40);
    cout << "distance :" << distance << endl;
    Position *position = new Position(RandomUtil::random(tankPos->xPos-1, tankPos->xPos+1), RandomUtil::random(tankPos->yPos-1, tankPos->yPos+1));
    list<Position *> positions;
    positions.push_back(tankPos);
    positions.push_back(position);
    positions.push_back(grassPos);

    Route *route = new Route(distance, positions);
    return route;
  }

  Order *GrassService::gotoTheGrassNearbyTheFlag(Tank *tank, int **gameMap) {

  }

  Position *GrassService::findTheGrassNearbyTheFlag(int **gameMap){

  }

  Order *GrassService::gotoTheNearestGrass(Tank *tank, Range *range, int** gameMap) {
    list<Position *> positions;
    Route *curRoute = new Route(INT_MAX, positions);
    for (int i = range->start->xPos; i < range->end->xPos; i++) {
      for (int j = range->start->yPos; j < range->end->yPos; j++) {
        if (gameMap[i][j] == 2) { // 如果是草丛
          cout << "- find a grass - " << endl;
          Position *grassPos = new Position(i, j);
          Route *route = this->route(tank->position, grassPos, gameMap);
          if (route->length < curRoute->length) { // 选 A Start 距离最短的路径
            curRoute = route;
          } else if (route->length == curRoute->length) { // 如果距离相等随机选一个
            int choice = RandomUtil::random(0, 1);
            cout << choice << endl;
            if (1 == choice) {
              curRoute = route;
            }
          }
        }
      }
    }
    if (INT_MAX == curRoute->length) { // 如果没有草丛
      cout << "- no grass - " << endl;
      Position *grassPos = new Position(RandomUtil::random(range->start->xPos, range->end->xPos), RandomUtil::random(range->start->yPos, range->end->yPos));
      curRoute = this->route(tank->position, grassPos, gameMap);
    }

    curRoute->positions.pop_front(); // 把当前位置 pop 掉
    Order *order = new Order(tank->tankdId, tank->position, curRoute->positions.front(), curRoute->positions.back());
    return order;
  }
}

using namespace grass;

int main() {
  cout << "Test Grass Begin" << endl;

  srand((unsigned)time(NULL));

  int **gameMap;
  gameMap = new int *[30];
  for(int i = 0; i < 30; i++) {
    gameMap[i] = new int[30];
    for (int j = 0; j < 30; j++) {
      gameMap[i][j] = 1;
    }
  }

  GrassService *service = new GrassService();
  Position *tankPos = new Position(3, 5);
  Tank *tank = new Tank(1024, tankPos, UP, 2);
  Position *start = new Position(2, 2);
  Position *end = new Position(15, 15);
  Range *range = new Range(start, end);

  Order *order = service->gotoTheNearestGrass(tank, range, gameMap);

  cout << "tankId: " << order->tankId
  << " current: " << order->current->xPos << ", " << order->current->yPos
  << " next: " << order->next->xPos << ", " << order->next->yPos
  << " destination: " << order->destination->xPos << ", " << order->destination->yPos
  << endl;

  cout << "Test Grass End" << endl;

  return 0;
}
