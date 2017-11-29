package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class Tank extends  MovableObject {
    private int shellSpeed;
    private Shell shell = null;
    private int hp;

    public Tank(int id, Position p, Direction dir, int speed, int shellSpeed, int hp) {
        super(id, p, dir, speed);
        this.shellSpeed = shellSpeed;
        this.hp = hp;
    }

    /**
    * fire at the given direction.
     * return the new shell if the previous shell is disappeared or the tank never fired.
     * Otherwise return the existing shell.
     */
    public Shell fireAt(Direction dir) {
        if(shell == null || shell.isDestroyed()) {
            shell =   new Shell(getId(), getPos().moveOneStep(dir), dir, shellSpeed);
            return shell;
        }
        else {
            return null;
        }

    }

    public int getHp() {
        return hp;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o)
            return true;
        if (o == null || getClass() != o.getClass())
            return false;
        Tank tank = (Tank) o;

        return getId() == tank.getId();

    }

    @Override
    public int hashCode() {
        return getId();
    }

    public void hit() {
        if(--hp <= 0) {
            destroyed();
        }
    }

    public boolean fired() {
        return (shell != null && !shell.isDestroyed());
    }

    public static int compare(Tank x, Tank y) {
        return (x.getId() < y.getId()) ? -1 : ((x.getId() == y.getId()) ? 0 : 1);
    }
}
