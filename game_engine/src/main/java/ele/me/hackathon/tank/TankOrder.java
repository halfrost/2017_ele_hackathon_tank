package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class TankOrder {
    private int tankId;
    private String order;
    private Direction parameter;

    public TankOrder(int tankId, String order, Direction parameter) {
        this.order = order;
        this.parameter = parameter;
        this.tankId = tankId;
    }

    public int getTankId() {
        return tankId;
    }

    public String getOrder() {
        return order;
    }

    public Direction getParameter() {
        return parameter;
    }

    @Override
    public String toString() {
        return "TankOrder{" +
                "tankId=" + tankId +
                ", order='" + order + '\'' +
                ", parameter='" + parameter + '\'' +
                '}';
    }

    public static boolean isValid(TankOrder o) {
        if (o.getOrder() == null) {
            System.out.println("Invalid order : " + o);
            return false;
        }
        switch (o.getOrder()) {
        case "move":
            break;
        case "fire":
        case "turnTo":
            if (o.getParameter() == null) {
                System.out.println("Invalid order : " + o);
                return false;
            }
            break;
        default:
            System.out.println("Invalid order : " + o);
            return false;
        }
        return true;
    }
}
