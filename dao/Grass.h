#include <string>

using namespace std;

namespace grass {

  class Position {
  public:
    int xPos;
    int yPos;

    Position();
    Position(int x, int y);
  };

  class Range {
  public:
    Position *start;
    Position *end;

    Range(Position *s, Position *e);
  };

  enum Direction {
      UP = 1,
      DOWN = 2,
      LEFT = 3,
      RIGHT = 4
  };

  class Tank {
  public:
    int tankdId;
    Position *position;
    Direction direction;
    int hp;

    Tank(int id, Position *pos, Direction dir, int h);
  };

  class Order {
  public:
    int tankId;
    string order;
    Direction direction;

    Order(int tankdId, string ord, Direction dir);
  };

  class GrassService {
  public:
    Order *gotoTheNearestGrass(Tank *tank, Range *range, int **gameMap);
  };
}
