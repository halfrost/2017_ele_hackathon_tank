package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class Shell extends MovableObject {

    public Shell(int id, Position pos, Direction dir, int speed) {
        super(id, pos, dir, speed);
    }

    public void turnTo(Direction dir) {
        throw new UnsupportedOperationException("Shell can't change its direction");
    }

    @Override
    public boolean equals(Object o) {
        if (this == o)
            return true;
        if (o == null || getClass() != o.getClass())
            return false;

        Shell shell = (Shell) o;

        return getId() == shell.getId();

    }

    @Override
    public int hashCode() {
        return getId();
    }

    @Override
    public String toString() {
        return "Shell{" +
                "id=" + getId() +
                ", pos=" + getPos() +
                ", dir=" + getDir() +
                '}';
    }

    public static int compare(Shell x, Shell y) {
        return (x.getId() < y.getId()) ? -1 : ((x.getId() == y.getId()) ? 0 : 1);
    }
}
