package ele.me.hackathon.tank;

import org.junit.Test;

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

/**
 * Created by lanjiangang on 25/11/2017.
 */
public class TankOrderTest {
    @Test
    public  void testInvalidOrder(){
        assertTrue(TankOrder.isValid(new TankOrder(1, "move", null))) ;
        assertFalse(TankOrder.isValid(new TankOrder(1, "fire", null))) ;
        assertFalse(TankOrder.isValid(new TankOrder(1, "turnTo", null))) ;
        assertFalse(TankOrder.isValid(new TankOrder(1, "hello", null))) ;
    }

}