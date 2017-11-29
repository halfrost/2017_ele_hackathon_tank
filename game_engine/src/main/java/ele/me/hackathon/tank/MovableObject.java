package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 03/11/2017.
 */
public class MovableObject {
    private int id;
    private Position pos;
    private Direction dir;
    private int speed;

    private boolean destroyed = false;

    public MovableObject(int id, Position pos, Direction dir, int speed) {
        this.pos = pos;
        this.dir = dir;
        this.speed = speed;
        this.id = id;
    }

    public void turnTo(Direction dir) {
        this.dir = dir;
    }

    /**
     * Because the object itself doesn't know the Map, it does not know if the movement is validate.
     * So it can only evaluate the moveOneStep track and let the GameStateMachine to verify it .
                      (0,0) - (0,1) - (0,2)
                                | UP
                LEFT  (1,0) - (1,1) - (1,2)  RIGHT
                                |  DOWN
                              (2,1)
      * @return
     */
    public Position[] evaluateMoveTrack() {
        Position[] track = new Position[speed];
        Position prePos = pos;
        for(int i = 0; i < speed; i++) {
            track[i] = prePos.moveOneStep(dir);
            prePos = track[i];
        }
        return track;
    }

    public void moveOneStep() {
        this.pos = pos.moveOneStep(dir);
    }
    public void withdrawOneStep() {
        this.pos = pos.withDrawStep(dir);
    }

    /**
     * move the object to given position.
     * @param position
     */
    public void moveTo(Position position) {
        this.pos = position;
    }

    public void destroyed() {
        this.destroyed = true;
    }

    public boolean isDestroyed() {
        return destroyed;
    }

    public Position getPos() {
        return pos;
    }

    public Direction getDir() {
        return dir;
    }

    public int getId() {
        return id;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o)
            return true;
        if (o == null || getClass() != o.getClass())
            return false;

        MovableObject that = (MovableObject) o;

        if (speed != that.speed)
            return false;
        if (destroyed != that.destroyed)
            return false;
        if (pos != null ? !pos.equals(that.pos) : that.pos != null)
            return false;
        return dir == that.dir;

    }

    @Override
    public int hashCode() {
        int result = pos != null ? pos.hashCode() : 0;
        result = 31 * result + (dir != null ? dir.hashCode() : 0);
        result = 31 * result + speed;
        result = 31 * result + (destroyed ? 1 : 0);
        return result;
    }

    protected void setSpeed(int speed) {
        this.speed = speed;
    }

    @Override
    public String toString() {
        return "MovableObject{" +
                "id=" + id +
                ", pos=" + pos +
                ", dir=" + dir +
                ", speed=" + speed +
                ", destroyed=" + destroyed +
                '}';
    }

}
