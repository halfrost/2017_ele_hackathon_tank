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

  Order::Order(int id, Position *cur, Position *nex) {
    tankId = id;
    current = cur;
    next = nex;
  }

  Route::Route(int len, list<Position *> pos) {
    length = len;
    positions = pos;
  }

  int RandomUtil::random(int from, int to) {
    return from + rand() % (to - from + 1);
  }

  Route *GrassService::route(Position *tankPos, Position *grassPos, int **gameMap) {
    int distance = RandomUtil::random(20, 40);
    cout << "distance :" << distance << endl;
    Position *position = new Position(RandomUtil::random(tankPos->xPos-1, tankPos->xPos+1), RandomUtil::random(tankPos->yPos-1, tankPos->yPos+1));
    list<Position *> positions;
    positions.push_front(position);

    Route *route = new Route(distance, positions);
    return route;
  }

  Order *GrassService::gotoTheNearestGrass(Tank *tank, Range *range, int** gameMap) {
    list<Position *> positions;
    Route *curRoute = new Route(INT_MAX, positions);
    for (int i = range->start->xPos; i < range->end->xPos; i++) {
      for (int j = range->start->yPos; j < range->end->yPos; j++) {
        if (gameMap[i][j] == 2) { // 如果是草丛
          cout << "here" << endl;
          Position *grassPos = new Position(i, j);
          Route *route = this->route(tank->position, grassPos, gameMap);
          if (route->length < curRoute->length) {
            curRoute = route;
          }
        }
      }
    }
    Order *order = new Order(tank->tankdId, tank->position, curRoute->positions.front());
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
      gameMap[i][j] = 2;
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
  << endl;

  cout << "Test Grass End" << endl;

  return 0;
}
