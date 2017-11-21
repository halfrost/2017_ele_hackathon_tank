#include <string>
#include <list>

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
    Position *current;
    Position *next;
    Position *destination; // 目标坐标（旗子／草丛）

    Order(int id, Position *cur, Position *nex, Position *dest);
  };

  class Route {
  public:
    int length;
    list<Position *> positions;

    Route(int len, list<Position *> pos);
  };

  class RandomUtil {
  public:
    static int random(int from, int to);
  };

  class GrassService {
  public:
    Order *gotoTheNearestGrass(Tank *tank, Range *range, int **gameMap);
    Order *gotoTheGrassNearbyTheFlag(Tank *tank, int **gameMap);
    Position *findTheGrassNearbyTheFlag(int **gameMap);
    Route *route(Position *tankPos, Position *grassPos, int **gameMap);
  };
}
