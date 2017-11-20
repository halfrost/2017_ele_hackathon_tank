#include "Grass.h"
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

  Order::Order(int id, string ord, Direction dir) {
    tankId = id;
    order = ord;
    direction = dir;
  }

  Order *GrassService::gotoTheNearestGrass(Tank *tank, Range *range, int** gameMap) {
    string operation = "move";
    Order *order = new Order(tank->tankdId, operation, tank->direction);
    return order;
  }
}

using namespace grass;

int main() {
  cout << "Test Grass Begin" << endl;

  int **gameMap;
  gameMap = new int *[30];
  for(int i = 0; i <30; i++) {
    gameMap[i] = new int[30];
  }

  GrassService *service = new GrassService();
  Position *tankPos = new Position(3, 5);
  Tank *tank = new Tank(1024, tankPos, UP, 2);
  Position *start = new Position(2, 2);
  Position *end = new Position(15, 15);
  Range *range = new Range(start, end);

  Order *order = service->gotoTheNearestGrass(tank, range, gameMap);

  cout << "tankId: " << order->tankId
  << " order: " << order->order
  << " direction: " << order->direction
  << endl;

  cout << "Test Grass End" << endl;

  return 0;
}
