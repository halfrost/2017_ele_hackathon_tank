package ele.me.hackathon.tank;

import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNull;

/**
 * Created by lanjiangang on 02/11/2017.
 */
public class TankTest {
    Tank tank;

    @Before
    public void setup() {
        tank = new Tank(0, new Position(1, 1), Direction.DOWN, 1, 2, 1);
    }

    @Test
    public void testTurnTo() {
        tank.turnTo(Direction.LEFT);
        assertEquals(Direction.LEFT, tank.getDir());
    }

    /*
                      (0,0) - (0,1) - (0,2)
                                | UP
                LEFT  (1,0) - (1,1) - (1,2)  RIGHT
                                |  DOWN
                              (2,1)
     */
    @Test
    public void testEvaluateMoveTrack() {
        Position[] track = tank.evaluateMoveTrack();
        assertArrayEquals(new Position[] { new Position(2, 1) }, track);

        tank.turnTo(Direction.RIGHT);
        assertArrayEquals(new Position[] { new Position(1, 2) }, tank.evaluateMoveTrack());

        tank.turnTo(Direction.UP);
        assertArrayEquals(new Position[] { new Position(0, 1) }, tank.evaluateMoveTrack());

        tank.turnTo(Direction.LEFT);
        assertArrayEquals(new Position[] { new Position(1, 0) }, tank.evaluateMoveTrack());
    }


    @Test
    public void testFireAt(){
        Shell shell = tank.fireAt(Direction.RIGHT);
        assertEquals("The shell shall appear at the next position at the fire direction", tank.getPos().moveOneStep(Direction.RIGHT), shell.getPos());
        assertEquals(Direction.RIGHT, shell.getDir());

        //move the shell one step
        shell.moveTo(shell.getPos().moveOneStep(Direction.RIGHT));
        Shell secondShell = tank.fireAt(Direction.LEFT);
        assertNull("The previous shell still exists", secondShell);

        //destroy the shell and fire again, then we shall get a new shell
        shell.destroyed();
        Shell thirdShell = tank.fireAt(Direction.DOWN);
        assertEquals("The shell shall appears at the next position at the fire direction", tank.getPos().moveOneStep(Direction.DOWN), thirdShell.getPos());
        assertEquals(Direction.DOWN, thirdShell.getDir());

    }
}