namespace grass {

  class Position {
  private:
    int xPos;
    int yPos;

  public:
    Position(int x, int y);
  };

  class Range {
  private:
    Position start;
    Position end;

  public:
    Range(int s, int e);
  };

  class Route {
  private:
    int tankId;
    Position grassPosition;
    int step;
    Position next;
  public:
  };


  class Grass {

  };

  class GrassService {
  public:
    Order* findNearbyGrass(Range *range) {

    }
  };
}
