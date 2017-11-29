package ele.me.hackathon.tank; /**
 * Created by lanjiangang on 27/10/2017.
 */

import java.util.Collection;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class GameStateMachine {
    private GameMap map;
    private Map<Integer, Tank> tanks;
    private List<Shell> shells = new LinkedList<>();
    private Position flagPos;
    private Map<String, Player> players;
    private GameOptions options;
    private List<Tank> destroyedTanks = new LinkedList<>();
    private List<Shell> destroyedShells = new LinkedList<>();
    private int replayStep = 1;

    public GameStateMachine(Map<Integer, Tank> tanks, GameMap map) {
        this.tanks = tanks;
        this.map = map;
    }

    public void newOrders(List<TankOrder> orders) {
        destroyedTanks.clear();
        destroyedShells.clear();

        evaluateShellsMovement();
        evaluateFireActions(filtOrder(orders, "fire"));
        evaluateTurnDirectionActions(filtOrder(orders, "turnTo"));
        evaluateMoveActions(filtOrder(orders, "move"));

    }

    public void printReplayLog() {
        System.out.println("ReplayStep " + replayStep++);

        StringBuffer sb = new StringBuffer("ReplayLog: {");
        generateTanksLog(sb);
        generateShellsLog(sb);
        generateFlagLog(sb);
        sb.append("}");
        System.out.println(sb.toString());
    }

    private void generateFlagLog(StringBuffer sb) {
        if (flagPos != null) {
            sb.append(",flag: {");
            sb.append("position:{x:").append(flagPos.getX()).append(",y:").append(flagPos.getY()).append("}}");
        }
    }

    private void generateShellsLog(StringBuffer sb) {
        List<Shell> shells = new LinkedList<>();
        shells.addAll(getShells());
        shells.addAll(destroyedShells);
        shells.sort(Shell::compare);

        sb.append("shells: [");
        for (Shell t : shells) {
            sb.append("{");
            sb.append("id:").append(t.getId());
            sb.append(", dir:").append(t.getDir().name());
            sb.append(", position:{x:").append(t.getPos().getX()).append(",y:").append(t.getPos().getY()).append("},");
            sb.append(", status:").append(t.isDestroyed() ? "destroyed" : "alive");
            sb.append("},");
        }
        if (shells.size() > 0)
            sb.deleteCharAt(sb.length() - 1);
        sb.append("]");
    }

    private void generateTanksLog(StringBuffer sb) {
        List<Tank> tanks = new LinkedList<>();
        tanks.addAll(getTankList());
        tanks.addAll(destroyedTanks);
        tanks.sort(Tank::compare);

        sb.append("tanks: [");
        for (Tank t : tanks) {
            sb.append("{");
            sb.append("id:").append(t.getId());
            sb.append(", dir:").append(t.getDir().name());
            sb.append(", position:{x:").append(t.getPos().getX()).append(",y:").append(t.getPos().getY()).append("}");
            sb.append(", status:").append(t.isDestroyed() ? "destroyed" : "alive");
            sb.append(", owner:").append(getPlayer(t).getName());
            sb.append("},");
        }
        if (tanks.size() > 0)
            sb.deleteCharAt(sb.length() - 1);
        sb.append("],");
    }

    private void evaluateShellsMovement() {
        for (int i = 0; i < options.getShellSpeed(); i++) {
            shells.forEach(s -> s.moveOneStep());
            shells.forEach(shell -> {
                Position pos = shell.getPos();
                if (map.isBarrier(pos)) {
                    shell.destroyed();
                } else {
                    Tank tankAt = getTankAt(pos);
                    if (tankAt != null) {
                        shell.destroyed();
                        tankAt.hit();
                    }
                }
            });
            //evaluate result on each step
            clearDestroyedTargets();
        }
    }

    private void clearDestroyedTargets() {
        List<Tank> newDestroyedTanks = tanks.values().stream().filter(t -> t.isDestroyed()).collect(Collectors.toCollection(() -> new LinkedList<>()));
        destroyedTanks.addAll(newDestroyedTanks);
        newDestroyedTanks.forEach(t -> tanks.remove(t.getId()));

        shells.removeIf(shell -> {
            if (shell.isDestroyed())
                destroyedShells.add(shell);
            return shell.isDestroyed();
        });
        printReplayLog();
    }

    private LinkedList<TankOrder> filtOrder(List<TankOrder> orders, String orderName) {
        return orders.stream().filter(o -> orderName.equals(o.getOrder()) && isValidateOrder(o)).collect(Collectors.toCollection(() -> new LinkedList<>()));
    }

    private void evaluateFireActions(List<TankOrder> orders) {

        List<Shell> newShells = new LinkedList<>();
        //let all tanks fire first so as to simulate all tanks are acting in the SAME time .
        for (TankOrder order : orders) {
            if (!isValidateOrder(order))
                continue;

            Shell shell = tanks.get(order.getTankId()).fireAt(order.getParameter());
            if (shell != null) {
                System.out.println("Tank " + order.getTankId() + " fire a new shell :" + shell);
                newShells.add(shell);
            }
        }
        //then the state machine evaluate new shells's result.
        //thus even if a tank is destroyed by new fired shell, it still has a chance to fire a shell before it dies.
        for (Shell shell : newShells) {
            if (map.isBarrier(shell.getPos())) {
                shell.destroyed();
                continue;
            }
            Tank tankAt = getTankAt(shell.getPos());
            if (tankAt != null) {
                tankAt.hit();
                shell.destroyed();
            }
        }

        newShells.removeIf(Shell::isDestroyed);
        getShells().addAll(newShells);

        clearDestroyedTargets();
    }

    private Tank getTankAt(Position pos) {
        List<Tank> tanksAt = tanks.values().stream().filter(t -> t.getPos().equals(pos)).collect(Collectors.toList());
        if (tanksAt.size() > 1) {
            String msg = "Found more than one tank in same position: ";
            for (Tank t : tanksAt) {
                msg += t;
            }
            throw new InvalidState(msg);
        }
        return tanksAt.size() > 0 ? tanksAt.get(0) : null;
    }

    private void evaluateTurnDirectionActions(List<TankOrder> orders) {
        for (TankOrder order : orders) {
            if (!isValidateOrder(order))
                return;

            tanks.get(order.getTankId()).turnTo(order.getParameter());
        }
    }

    private boolean isValidateOrder(TankOrder order) {
        if (tanks.containsKey(order.getTankId()) && !tanks.get(order.getTankId()).isDestroyed())
            return true;

        return false;
    }

    private void evaluateMoveActions(List<TankOrder> orders) {
        List<Tank> tanksToMove = orders.stream().filter(o -> isValidateOrder(o)).map(o -> getTanks().get(o.getTankId())).collect(Collectors.toList());

        for (int i = 0; i < options.getTankSpeed(); i++) {

            List<Tank> f2fTanks = tanksToMove.stream().filter(t -> hasFace2FaceTank(t, tanksToMove)).collect(Collectors.toList());
            tanksToMove.removeAll(f2fTanks);
            tanksToMove.forEach(t -> t.moveOneStep());
            withdrawUntilNoOverlap(tanksToMove);

            //check flag
            checkFlag();

            //check shells
            tanksToMove.stream().forEach(t -> {
                getShellAt(t.getPos()).forEach(s -> {
                    s.destroyed();
                    t.hit();
                });
            });

            tanksToMove.removeIf(t -> t.isDestroyed());
            clearDestroyedTargets();
        }
    }

     public void checkFlag() {
         getTankList().stream().filter(t -> getFlagPos() != null && flagPos.equals(t.getPos())).forEach(t -> {
             flagPos = null;
             Player p = getPlayer(t);
             p.captureFlag(t);
         });
     }

    private boolean hasFace2FaceTank(Tank t, List<Tank> tanksToMove) {
        Position nextPos = t.getPos().moveOneStep(t.getDir());
        for (Tank tank : tanksToMove) {
            if (nextPos.equals(tank.getPos()) && tank.getDir().negative(t.getDir())) {
                return true;
            }
        }
        return false;
    }

    private void withdrawUntilNoOverlap(List<Tank> allMovedTanks) {
        List<Tank> invalidTanks = null;
        while (!(invalidTanks = findInvalidTanks()).isEmpty()) {
            invalidTanks.stream().filter(t -> allMovedTanks.contains(t)).forEach(t -> {
                t.withdrawOneStep();
                allMovedTanks.remove(t);
            });
        }
    }

    private List<Tank> findInvalidTanks() {
        return getTanks().values().stream().filter(t -> invalidPosition(t.getPos())).collect(Collectors.toList());
    }

    private boolean invalidPosition(Position pos) {
        return map.isBarrier(pos) || existingMoreThanOneTanks(pos);
    }

    private Player getPlayer(Tank t) {
        return getPlayers().values().stream().filter(p -> p.getTanks().contains(t.getId())).findFirst().get();
    }

    private Player getPlayer(int tid) {
        return getPlayers().values().stream().filter(p -> p.getTanks().contains(tid)).findFirst().get();
    }

    private boolean existingMoreThanOneTanks(Position pos) {
        return (getTankList().stream().filter(t -> pos.equals(t.getPos())).count() > 1);
    }

    public Collection<Tank> getTankList() {
        return getTanks().values();
    }

    private List<Shell> getShellAt(Position position) {
        List<Shell> shellList = shells.stream().filter(shell -> !shell.isDestroyed() && shell.getPos().equals(position))
                .collect(Collectors.toCollection(() -> new LinkedList<Shell>()));
        return shellList;
    }

    public int getFlagNoByPlayer(String name) {
        return getPlayers().get(name).getNoOfFlag();
    }

    public Position generateFlag() {
        flagPos = new Position(map.size() / 2, map.size() / 2);
        System.out.println("Generate flag at " + flagPos);
        //check if there is a tank just stays on the same position.
        checkFlag();
        return flagPos;
    }

    public Map<String, GameState> reportState() {
        return getPlayers().keySet().stream().collect(Collectors.toMap(name -> name, name -> generatePlayerState(name)));
    }

    private GameState generatePlayerState(String playerName) {
        GameState playerState = new GameState(playerName);

        //add own tanks
        getPlayers().get(playerName).getTanks().stream().filter(tankId -> tankExisting(tankId)).forEach(tankId -> {
            playerState.getTanks().add(getTanks().get(tankId));
        });

        //add enemy's tanks if they are visible.
        getPlayers().entrySet().stream().filter(e -> !e.getKey().equals(playerName)).forEach(e -> {
            e.getValue().getTanks().stream().filter(tankId -> tankVisible(tankId)).forEach(tankId -> {
                playerState.getTanks().add(getTanks().get(tankId));
            });
        });

        //all shells which are visible
        getShells().stream().filter(s -> map.isVisible(s.getPos())).forEach(s -> playerState.getShells().add(s));

        playerState.setYourFlagNo(getFlagNoByPlayer(playerName));
        playerState.setEnmeyFlagNo(getEnemyFlagNo(playerName));
        playerState.setFlagPos(getFlagPos());

        return playerState;

    }

    private int getEnemyFlagNo(String playerName) {
        return getPlayers().values().stream().filter(p -> !p.getName().equals(playerName)).findFirst().get().getNoOfFlag();
    }

    private boolean tankVisible(Integer tankId) {
        return tankExisting(tankId) && map.isVisible(getTanks().get(tankId).getPos());
    }

    private boolean tankExisting(Integer tankId) {
        return getTanks().containsKey(tankId) && !getTanks().get(tankId).isDestroyed();
    }

    public boolean gameOvered() {
        boolean result = false;
        for (Player p : players.values()) {
            if (p.getTanks().stream().noneMatch(id -> tanks.containsKey(id) && !tanks.get(id).isDestroyed())) {
                System.out.println("Player " + p.getName() + " loses all tanks!");
                result = true;
            }
        }
        return result;
    }

    public Map<String, Integer> countScore(int tankScore, int flagScore) {
        return players.values().stream().collect(Collectors.toMap(p -> p.getName(), p -> {
            long score =
                    p.getTanks().stream().filter(id -> tanks.containsKey(id) && !tanks.get(id).isDestroyed()).count() * tankScore + p.getNoOfFlag() * flagScore;
            return (int) score;
        }));
    }

    public Map<Integer, Tank> getTanks() {
        return tanks;
    }

    protected List<Shell> getShells() {
        return shells;
    }

    public void setPlayers(Map<String, Player> players) {
        this.players = players;
    }

    public Map<String, Player> getPlayers() {
        return players;
    }

    public void setOptions(GameOptions options) {
        this.options = options;
    }

    public long getPlayerTankNo(String playerAAddres) {
        return getPlayers().get(playerAAddres).getTanks().stream().filter(id -> tanks.containsKey(id)).count();
    }

    public Position getFlagPos() {
        return flagPos;
    }
}
