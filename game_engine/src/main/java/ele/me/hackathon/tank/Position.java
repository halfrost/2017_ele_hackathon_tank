package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class Position {
    private int x;
    private int y;

    public Position(int x, int y) {
        this.x = x;
        this.y = y;
    }

    public int getX() {
        return x;
    }


    public int getY() {
        return y;
    }
    /*
                        (0,0) - (0,1) - (0,2)
                                | UP
                LEFT  (1,0) - (1,1) - (1,2)  RIGHT
                                |  DOWN
                              (2,1)

     */
    public Position moveOneStep(Direction dir) {
        switch (dir) {
        case UP:
            return new Position(x - 1 , y);
        case DOWN:
            return new Position(x + 1, y);
        case LEFT:
            return new Position(x, y - 1);
        case RIGHT:
            return new Position(x, y + 1);
        default:
            return null;
        }
    }

    public Position withDrawStep(Direction dir) {
        return moveOneStep(opposite(dir));
    }

    private Direction opposite(Direction dir) {
        switch (dir) {
        case UP:
            return Direction.DOWN;
        case DOWN:
            return Direction.UP;
        case LEFT:
            return Direction.RIGHT;
        case RIGHT:
            return Direction.LEFT;
        default:
            return null;
        }

    }

    @Override
    public boolean equals(Object o) {
        if (this == o)
            return true;
        if (o == null || getClass() != o.getClass())
            return false;

        Position position = (Position) o;

        if (x != position.x)
            return false;
        return y == position.y;

    }

    @Override
    public int hashCode() {
        int result = x;
        result = 31 * result + y;
        return result;
    }

    @Override
    public String toString() {
        return "Position{" +
                "x=" + x +
                ", y=" + y +
                '}';
    }

}
