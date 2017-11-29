package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public enum Direction {
    UP(1),
    DOWN(2),
    LEFT(3),
    RIGHT(4);

    private final int value;

    Direction(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }

    public static Direction findByValue(int value) {
        switch (value) {
        case 1:
            return UP;
        case 2:
            return DOWN;
        case 3:
            return LEFT;
        case 4:
            return RIGHT;
        default:
            return null;
        }
    }

    public boolean negative(Direction dir) {
        return (dir.value + this.value) == 3 || (dir.value + this.value) == 7;
    }
}
