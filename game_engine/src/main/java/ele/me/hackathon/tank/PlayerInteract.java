package ele.me.hackathon.tank;

import ele.me.hackathon.tank.player.Args;
import ele.me.hackathon.tank.player.Order;
import ele.me.hackathon.tank.player.PlayerServer;
import org.apache.thrift.TException;

import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.stream.Collectors;

/**
 * Created by lanjiangang on 08/11/2017.
 */
public class PlayerInteract {
    private final PlayerServer.Client client;
    private final List<Integer> tanks;
    private final GameMap map;

    private final String address;
    private final GameOptions gameOptions;

    LinkedBlockingQueue<List<TankOrder>> commandQueue = new LinkedBlockingQueue<>();
    LinkedBlockingQueue<GameState> statusQueue = new LinkedBlockingQueue<>();

    Thread t;

    public void start() {
        t.setDaemon(true);
        t.start();
    }

    public PlayerInteract(String addr, PlayerServer.Client client, GameMap map, List<Integer> tanks, GameOptions gameOptions) {
        this.address = addr;
        this.client = client;
        this.map = map;
        this.tanks = tanks;
        this.gameOptions = gameOptions;
        t = new Thread(new Runnable() {
            @Override
            public void run() {

                try {
                    //wait an signal to go
                    GameState state = statusQueue.take();

                    client.uploadMap(convertMap(map));

                    client.uploadParamters(convertGameOptions(gameOptions));

                    client.assignTanks(tanks);
                } catch (TException e) {
                    e.printStackTrace();
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }

                for (; ; ) {

                    try {
                        GameState state = statusQueue.take();
                        System.out.println("Send state to " + getAddress() + " : " + PlayerInteract.toString(convert(state)));
                        //System.out.println("Send state to " + getAddress() + " : " + Util.toJson(state));
                        client.latestState(convert(state));
                    } catch (Exception e) {
                        e.printStackTrace();
                    }

                    try {
                        List<Order> orders = client.getNewOrders();
                        System.out.println("Recv orders from " + getAddress() + " : " + PlayerInteract.toString(orders));
                        List<TankOrder> tankOrders = convertOrders(orders);
                        if (!verifyPlayerInput(tankOrders)) {
                            tankOrders = new LinkedList<>();
                        }
                        commandQueue.offer(tankOrders);
                    } catch (Exception e) {
                        e.printStackTrace();
                        commandQueue.offer(new LinkedList<>());
                    }

                }
            }
        });
    }

    protected boolean verifyPlayerInput(List<TankOrder> tankOrders) {
        if (tankOrders.stream().filter(o -> !tanks.contains(o.getTankId())).count() > 0) {
            tankOrders.stream().filter(o -> !tanks.contains(o.getTankId())).forEach(o -> System.out.println(address + " try to control enemy's tank:" + o));
            return false;
        }
        if (tankOrders.stream().map(TankOrder::getTankId).collect(Collectors.toSet()).size() < tankOrders.size()) {
            System.out.println(address + " has duplicate orders");
            return false;
        }

        if (tankOrders.stream().filter(o -> !TankOrder.isValid(o)).count() > 0) {
            System.out.println(address + " has invalid orders");
            return false;
        }
        return true;
    }

    private Args convertGameOptions(GameOptions gameOptions) {
        Args args = new Args();
        args.setTankSpeed(gameOptions.getTankSpeed());
        args.setShellSpeed(gameOptions.getShellSpeed());
        args.setTankHP(gameOptions.getTankHP());
        args.setFlagScore(gameOptions.getFlagScore());
        args.setTankScore(gameOptions.getTankScore());
        args.setMaxRound(gameOptions.getMaxRound());
        args.setRoundTimeoutInMs(gameOptions.getRoundTimeout());
        return args;
    }

    protected ele.me.hackathon.tank.player.GameState convert(GameState state) {
        ele.me.hackathon.tank.player.GameState res = new ele.me.hackathon.tank.player.GameState();
        res.setTanks(convertTanks(state.getTanks()));
        res.setShells(convertShells(state.getShells()));
        res.setYourFlagNo(state.getYourFlagNo());
        res.setEnemyFlagNo(state.getEnmeyFlagNo());
        if (state.getFlagPos() != null) {
            res.setFlagPos(convertPosition(state.getFlagPos()));
        }
        return res;
    }

    private List<ele.me.hackathon.tank.player.Shell> convertShells(List<Shell> shells) {
        return shells.stream().map(s -> convertShell(s)).collect(Collectors.toList());
    }

    private ele.me.hackathon.tank.player.Shell convertShell(Shell s) {
        return new ele.me.hackathon.tank.player.Shell(s.getId(), convertPosition(s.getPos()), convertDir(s.getDir()));
    }

    private List<ele.me.hackathon.tank.player.Tank> convertTanks(List<Tank> tanks) {
        return tanks.stream().map(t -> convertTank(t)).collect(Collectors.toList());
    }

    private ele.me.hackathon.tank.player.Tank convertTank(Tank t) {
        ele.me.hackathon.tank.player.Tank res = new ele.me.hackathon.tank.player.Tank();
        res.setHp(t.getHp());
        res.setId(t.getId());
        res.setPos(convertPosition(t.getPos()));
        res.setDir(convertDir(t.getDir()));
        return res;
    }

    private ele.me.hackathon.tank.player.Direction convertDir(Direction dir) {
        return ele.me.hackathon.tank.player.Direction.findByValue(dir.getValue());
    }

    private ele.me.hackathon.tank.player.Position convertPosition(Position pos) {
        return new ele.me.hackathon.tank.player.Position(pos.getX(), pos.getY());
    }

    private List<TankOrder> convertOrders(List<ele.me.hackathon.tank.player.Order> newOrders) {
        List<TankOrder> res = newOrders.stream().map(o -> convertTankOrder(o)).collect(Collectors.toList());
        return res;
    }

    private TankOrder convertTankOrder(ele.me.hackathon.tank.player.Order o) {
        return new TankOrder(o.getTankId(), o.getOrder(), convertDir(o.getDir()));
    }

    private Direction convertDir(ele.me.hackathon.tank.player.Direction dir) {
        if (dir == null)
            return null;

        return Direction.findByValue(dir.getValue());
    }

    private List<List<Integer>> convertMap(GameMap map) {
        List<List<Integer>> m = new LinkedList<>();
        for (int[] line : map.getPixels()) {
            List<Integer> l = new LinkedList<>();
            for (int p : line) {
                l.add(p);
            }
            m.add(l);
        }
        return m;
    }

    public PlayerServer.Client getClient() {
        return client;
    }

    public LinkedBlockingQueue<List<TankOrder>> getCommandQueue() {
        return commandQueue;
    }

    public LinkedBlockingQueue<GameState> getStatusQueue() {
        return statusQueue;
    }

    public String getAddress() {
        return address;
    }

    public static String toString(ele.me.hackathon.tank.player.GameState state) {
        java.lang.StringBuilder sb = new java.lang.StringBuilder("GameState(");

        sb.append("tanks:");
        if (state.getTanks() == null) {
            sb.append("null");
        } else {
            sb.append("[");
            state.getTanks().forEach(s -> {
                sb.append(s);
                sb.append(",");
            });
            sb.append("]");
        }
        sb.append(",shells:");
        if (state.getShells() == null) {
            sb.append("null");
        } else {
            sb.append("[");
            state.getShells().forEach(s -> {
                sb.append(s);
                sb.append(",");
            });
            sb.append("]");
        }
        sb.append(",yourFlags:");
        sb.append(state.getYourFlagNo());
        sb.append(",enemyFlags:");
        sb.append(state.getEnemyFlagNo());
        if (state.isSetFlagPos()) {
            sb.append(",flagPos:");
            sb.append(state.getFlagPos());
        }
        sb.append(")");
        return sb.toString();
    }

    private static String toString(List<Order> orders) {
        java.lang.StringBuilder sb = new java.lang.StringBuilder("Orders[");
        orders.forEach(o -> {
            sb.append(o);
            sb.append(", ");
        });
        sb.append("]");
        return sb.toString();
    }

}
