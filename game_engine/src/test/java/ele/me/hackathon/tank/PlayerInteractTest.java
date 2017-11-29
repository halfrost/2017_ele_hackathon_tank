package ele.me.hackathon.tank;

import org.junit.Before;
import org.junit.Test;

import java.util.*;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;

/**
 * Created by lanjiangang on 09/11/2017.
 */
public class PlayerInteractTest {
    private PlayerInteract interact;

    @Before
    public void setup() {
        interact = new PlayerInteract("playerA", null, null, Arrays.asList(new Integer[]{1, 2}), null);
    }

    @Test
    public void testVerifyDuplicateOrders() {
        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "move", Direction.UP));
        orders.add(new TankOrder(1, "fire", Direction.UP));

        assertFalse("duplicate orders on same tank is invalid", interact.verifyPlayerInput(orders));
    }

    @Test
    public void testVerifyOrdersOnEnemyTank() {
        List<TankOrder> orders = new LinkedList<>();
        orders.add(new TankOrder(1, "fire", Direction.UP));
        orders.add(new TankOrder(3, "move", Direction.UP));

        assertFalse("order on enemy's tank is invalid", interact.verifyPlayerInput(orders));
    }

    @Test
    public void testConvertGameState() {
        Tank tankA = new Tank(1, new Position(1, 1), Direction.DOWN, 1, 2, 1);
        Tank tankB = new Tank(2, new Position(1, 5), Direction.DOWN, 1, 2, 2);

        GameState state = new GameState("blue");
        state.getTanks().add(tankA);
        state.getTanks().add(tankB);

        ele.me.hackathon.tank.player.GameState res = interact.convert(state);

        assertEquals(2, res.getTanks().size());
        assertTankEquals(tankA, res.getTanks().get(0));
        assertTankEquals(tankB, res.getTanks().get(1));
    }

    private void assertTankEquals(Tank tankA, ele.me.hackathon.tank.player.Tank tank) {
        assertEquals(tankA.getId(), tank.getId());
        assertEquals(tankA.getHp(), tank.getHp());
        assertPositionEquals(tankA.getPos(), tank.getPos());
        assertDirectionEquals(tankA.getDir(), tank.getDir());
    }

    private void assertPositionEquals(Position pos, ele.me.hackathon.tank.player.Position pos1) {
        assertEquals(pos.getX(), pos1.getX());
        assertEquals(pos.getY(), pos1.getY());
    }

    private void assertDirectionEquals(Direction dir, ele.me.hackathon.tank.player.Direction dir1) {
        assertEquals(dir.getValue(), dir1.getValue());
    }

}