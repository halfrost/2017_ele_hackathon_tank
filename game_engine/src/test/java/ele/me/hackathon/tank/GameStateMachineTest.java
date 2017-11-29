package ele.me.hackathon.tank;

import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;

import java.io.IOException;
import java.util.*;

import static org.junit.Assert.*;

/**
 * Created by lanjiangang on 01/11/2017.
 */
public class GameStateMachineTest {
    private static GameMap map = null;
    private GameStateMachine stateMachine = null;
    private Tank tankA;
    private Tank tankB;
    GameOptions options;

    @BeforeClass
    public static void beforeClass() throws IOException {
        map = MapFactory.getMap();
    }

    @Before
    public void setup() {
        tankA = new Tank(1, new Position(1, 1), Direction.DOWN, 1, 2, 1);
        tankB = new Tank(2, new Position(1, 5), Direction.DOWN, 1, 2, 2);

        Map tanks = new LinkedHashMap<Integer, Tank>();
        tanks.put(1, tankA);
        tanks.put(2, tankB);

        options = new GameOptions(1, 1, 2, 2, 1, 1, 100, 2000);
        stateMachine = new GameStateMachine(tanks, map);
        stateMachine.setOptions(options);

        Map<String, Player> players = new LinkedHashMap<>();
        players.put("playerA", new Player("playerA", Arrays.asList(new Integer[] { 1 })));
        players.put("playerB", new Player("playerB", Arrays.asList(new Integer[] { 2 })));
        stateMachine.setPlayers(players);
    }

    @Test
    public void testShellMoveToTank() throws InvalidOrder {
        tankA.moveTo(new Position(2, 2));
        //add a shell which will end at position of tankA
        Shell shellA = new Shell(1, new Position(2, 4), Direction.LEFT, 2);
        stateMachine.getShells().add(shellA);

        stateMachine.newOrders(new LinkedList<TankOrder>());
        assertTrue(tankA.isDestroyed());
        assertTrue(shellA.isDestroyed());

    }

    @Test
    public void testShellPassThroughTank() throws InvalidOrder {
        tankB.moveTo(new Position(5, 5));
        //add a shell which will pass through tankB
        Shell shellB = new Shell(2, new Position(6, 5), Direction.UP, 2);
        stateMachine.getShells().add(shellB);

        stateMachine.newOrders(new LinkedList<TankOrder>());

        assertEquals("tankB has 1 hp left", 1, tankB.getHp());
        assertTrue(shellB.isDestroyed());
    }

    @Test
    public void testShellsEndOnSameTank() throws InvalidOrder {
        tankA.moveTo(new Position(2, 2));
        //add 2 shells which will hit tank at same time
        Shell shellA = new Shell(1, new Position(2, 4), Direction.LEFT, 2);
        Shell shellB = new Shell(1, new Position(4, 2), Direction.UP, 2);
        stateMachine.getShells().add(shellA);
        stateMachine.getShells().add(shellB);

        stateMachine.newOrders(new LinkedList<TankOrder>());

        assertTrue(tankA.isDestroyed());
        assertEquals("both shells are destroyed even if tankA only has one HP", 0, stateMachine.getShells().size());
    }

    @Test
    public void testEvaluateShellMovementStepByStep() {
        tankA.moveTo(new Position(2, 2));
        //add a shell which will hit tankA first
        Shell shellA = new Shell(1, new Position(2, 3), Direction.LEFT, 2);
        Shell shellB = new Shell(1, new Position(2, 4), Direction.LEFT, 2);
        stateMachine.getShells().add(shellA);
        stateMachine.getShells().add(shellB);

        stateMachine.newOrders(new LinkedList<TankOrder>());

        assertTrue(tankA.isDestroyed());

        assertTrue(shellA.isDestroyed());
        assertFalse("shellB shall still exist", shellB.isDestroyed());

    }

    @Test
    public void testFireOrders() {
        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "fire", Direction.RIGHT));
        orders.add(new TankOrder(2, "fire", Direction.LEFT));

        stateMachine.newOrders(orders);
        assertEquals(2, stateMachine.getShells().size());

        //no more shells generated if tanks already fired.
        stateMachine.newOrders(orders);
        assertEquals(2, stateMachine.getShells().size());

    }

    @Test
    public void testFireTowardsEachOther() {
        tankA.moveTo(new Position(1, 1));
        tankB.moveTo(new Position(1, 2));

        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(2, "fire", Direction.LEFT));
        orders.add(new TankOrder(1, "fire", Direction.RIGHT));

        stateMachine.newOrders(orders);
        assertEquals(0, stateMachine.getShells().size());
        assertTrue("tankA is destroyed", tankA.isDestroyed());
        assertEquals("tankB shall left one HP", 1, tankB.getHp());
    }

    @Test
    public void testTurnToOrders() throws Exception {

        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "turnTo", Direction.RIGHT));
        orders.add(new TankOrder(2, "turnTo", Direction.LEFT));

        stateMachine.newOrders(orders);

        Map<Integer, Tank> tanks = stateMachine.getTanks();
        assertEquals(new Position(1, 1), tanks.get(1).getPos());
        assertEquals(Direction.RIGHT, tanks.get(1).getDir());

        assertEquals(new Position(1, 5), tanks.get(2).getPos());
        assertEquals(Direction.LEFT, tanks.get(2).getDir());
    }

    @Test
    public void testNormalMoveOrders() throws Exception {

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.RIGHT));
        orders.add(new TankOrder(2, "move", Direction.LEFT));

        stateMachine.newOrders(orders);

        Map<Integer, Tank> tanks = stateMachine.getTanks();
        assertEquals(new Position(1, 1).moveOneStep(tanks.get(1).getDir()), tanks.get(1).getPos());
        assertEquals(new Position(1, 5).moveOneStep(tanks.get(2).getDir()), tanks.get(2).getPos());
    }

    @Test
    public void testMoveToBarrier() throws InvalidOrder {
        tankA.turnTo(Direction.UP);

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.UP));

        stateMachine.newOrders(orders);

        assertEquals(new Position(1, 1), tankA.getPos());
    }

    @Test
    public void testMoveToSamePos() throws InvalidOrder {
        tankA.turnTo(Direction.DOWN);
        tankA.moveTo(new Position(1, 1));

        tankB.turnTo(Direction.LEFT);
        tankB.moveTo(new Position(2, 2));

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.UP));
        orders.add(new TankOrder(2, "move", Direction.UP));

        stateMachine.newOrders(orders);

        assertEquals(new Position(1, 1), tankA.getPos());
        assertEquals(new Position(2, 2), tankB.getPos());
    }

    @Test
    public void testMoveToAnotherTanksPrePos() throws InvalidOrder {
        tankA.turnTo(Direction.DOWN);
        tankA.moveTo(new Position(1, 1));

        tankB.turnTo(Direction.LEFT);
        tankB.moveTo(new Position(1, 2));

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.UP));
        orders.add(new TankOrder(2, "move", Direction.UP));

        stateMachine.newOrders(orders);

        assertEquals(new Position(2, 1), tankA.getPos());
        assertEquals(new Position(1, 1), tankB.getPos());
    }

    @Test
    public void testMoveToAnotherTanksPos() throws InvalidOrder {
        tankA.turnTo(Direction.DOWN);
        tankA.moveTo(new Position(1, 1));

        tankB.turnTo(Direction.LEFT);
        tankB.moveTo(new Position(1, 2));

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(2, "move", Direction.UP));

        stateMachine.newOrders(orders);

        assertEquals(new Position(1, 1), tankA.getPos());
        assertEquals(new Position(1, 2), tankB.getPos());
    }

    @Test
    public void testMoveFace2Face() {
        //move two tanks face to face
        tankA.turnTo(Direction.RIGHT);
        tankA.moveTo(new Position(1, 1));

        tankB.turnTo(Direction.LEFT);
        tankB.moveTo(new Position(1, 2));

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.RIGHT));
        orders.add(new TankOrder(2, "move", Direction.LEFT));

        stateMachine.newOrders(orders);

        assertEquals(new Position(1, 1), tankA.getPos());
        assertEquals(new Position(1, 2), tankB.getPos());

    }

    @Test
    public void testWithdrawMakesOverlap() {
        tankA.moveTo(new Position(1, 1));
        tankA.turnTo(Direction.RIGHT);
        tankB.moveTo(new Position(1, 2));
        tankB.turnTo(Direction.RIGHT);
        Tank tankC = new Tank(3, new Position(2, 3), Direction.UP, 1, 2, 1);
        stateMachine.getTanks().put(3, tankC);

        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.UP));
        orders.add(new TankOrder(2, "move", Direction.UP));
        orders.add(new TankOrder(3, "move", Direction.UP));

        stateMachine.getPlayers().get("playerB").setTanks(Arrays.asList(new Integer[] { 2, 3 }));

        stateMachine.newOrders(orders);

        assertEquals(new Position(1, 1), tankA.getPos());
        assertEquals(new Position(1, 2), tankB.getPos());
        assertEquals(new Position(2, 3), tankC.getPos());
    }

    @Test
    public void testMoveToShell() throws InvalidOrder {
        List<TankOrder> orders = new LinkedList<TankOrder>();
        orders.add(new TankOrder(1, "move", Direction.DOWN));
        orders.add(new TankOrder(2, "move", Direction.DOWN));

        //add shells will move onto the target position of tankA and tankB
        Shell shellA = new Shell(1, new Position(2, 3), Direction.LEFT, 2);
        Shell shellB = new Shell(2, new Position(2, 7), Direction.LEFT, 2);
        stateMachine.getShells().add(shellA);
        stateMachine.getShells().add(shellB);

        stateMachine.newOrders(orders);

        assertTrue("tankA shall be destroyed", tankA.isDestroyed());
        assertFalse("tankB shall servive because it has 2 hp", tankB.isDestroyed());
        assertEquals("tankB has one HP left", 1, tankB.getHp());
        assertEquals(new Position(2, 5), tankB.getPos());
        assertTrue(shellA.isDestroyed());
        assertTrue(shellB.isDestroyed());
    }

    @Test
    public void testPassThroughShell() throws InvalidOrder {
        options.setTankSpeed(2);

        //add shell will move onto the path of tankA and tankB
        Shell shellA = new Shell(1, new Position(2, 3), Direction.LEFT, 2);
        Shell shellB = new Shell(2, new Position(2, 7), Direction.LEFT, 2);
        stateMachine.getShells().add(shellA);
        stateMachine.getShells().add(shellB);

        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "move", Direction.DOWN));
        orders.add(new TankOrder(2, "move", Direction.DOWN));

        stateMachine.newOrders(orders);

        assertTrue("tankA shall be destroyed", tankA.isDestroyed());
        assertFalse("tankB shall servive because it has 2 hp", tankB.isDestroyed());
        assertEquals("tankB has one HP left", 1, tankB.getHp());
        assertEquals(new Position(3, 5), tankB.getPos());
        assertTrue(shellA.isDestroyed());
        assertTrue(shellB.isDestroyed());
    }

    @Test
    public void testGetFlag() throws InvalidOrder {
        Position flagPos = stateMachine.generateFlag();
        tankA.moveTo(flagPos.moveOneStep(Direction.DOWN));
        tankA.turnTo(Direction.UP);

        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "move", Direction.UP));

        stateMachine.newOrders(orders);

        assertEquals("PlayerA got one flag.", 1, stateMachine.getFlagNoByPlayer("playerA"));
        assertEquals("PlayerB got none flag.", 0, stateMachine.getFlagNoByPlayer("playerB"));
    }

    @Test
    public void testGenerateFlag() {
        Position flagPos = new Position(map.size() / 2, map.size() / 2);
        tankA.moveTo(flagPos);
        assertEquals(null, stateMachine.getFlagPos());
        assertEquals(0, stateMachine.getFlagNoByPlayer("playerA"));

        stateMachine.generateFlag();
        assertEquals(1, stateMachine.getFlagNoByPlayer("playerA"));

    }

    @Test
    public void reportState() {
        //move tankA to wood
        tankA.moveTo(new Position(16, 1));

        //add a shell in wood
        Shell shellA = new Shell(1, new Position(16, 4), Direction.LEFT, 2);
        //add a shell visible
        Shell shellB = new Shell(2, new Position(6, 5), Direction.UP, 2);
        stateMachine.getShells().add(shellA);
        stateMachine.getShells().add(shellB);
        stateMachine.generateFlag();

        Map<String, GameState> state = stateMachine.reportState();
        assertEquals(2, state.size());
        assertEquals("playerA shall see 2 tanks", 2, state.get("playerA").getTanks().size());
        assertEquals("playerB shall see its own tank", 1, state.get("playerB").getTanks().size());
        assertEquals(1, state.get("playerA").getShells().size());
        assertEquals(1, state.get("playerB").getShells().size());
        assertEquals(stateMachine.getFlagPos(), state.get("playerB").getFlagPos());
    }

    @Test
    public void testJudgeGameOver() {
        assertFalse(stateMachine.gameOvered());
        tankA.destroyed();
        assertTrue(stateMachine.gameOvered());
        tankB.destroyed();
        assertTrue(stateMachine.gameOvered());
    }

    @Test
    public void testCountScore() {
        Map<String, Integer> scores = stateMachine.countScore(1, 1);
        assertEquals(scores.get("playerA"), scores.get("playerB"));

        tankA.destroyed();
        scores = stateMachine.countScore(1, 1);
        assertEquals(new Integer(0), scores.get("playerA"));
        assertEquals(new Integer(1), scores.get("playerB"));

    }
}