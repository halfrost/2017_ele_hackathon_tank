package ele.me.hackathon.tank;

import org.junit.Test;

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

/**
 * Created by lanjiangang on 14/11/2017.
 */
public class DirectionTest {
    @Test
    public void testNegative() {
        assertTrue(Direction.UP.negative(Direction.DOWN));
        assertTrue(Direction.LEFT.negative(Direction.RIGHT));

        assertFalse(Direction.LEFT.negative(Direction.DOWN));
        assertFalse(Direction.LEFT.negative(Direction.UP));
        assertFalse(Direction.RIGHT.negative(Direction.DOWN));
        assertFalse(Direction.RIGHT.negative(Direction.UP));

    }

}