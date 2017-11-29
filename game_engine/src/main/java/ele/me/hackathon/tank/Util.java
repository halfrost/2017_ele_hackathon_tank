package ele.me.hackathon.tank;

import org.codehaus.jackson.map.ObjectMapper;
import org.codehaus.jackson.map.ObjectWriter;

import java.io.IOException;

/**
 * Created by lanjiangang on 15/11/2017.
 */
public class Util {
    public  static String toJson(Object o) {
        ObjectWriter ow = new ObjectMapper().writer().withDefaultPrettyPrinter();
        String json = "";
        try {
            json = ow.writeValueAsString(o);
        } catch (IOException e) {
            e.printStackTrace();
        }
        return json;
    }
}
