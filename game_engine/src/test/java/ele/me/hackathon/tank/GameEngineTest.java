package ele.me.hackathon.tank;

import org.apache.http.*;
import org.apache.http.impl.bootstrap.HttpServer;
import org.apache.http.impl.bootstrap.ServerBootstrap;
import org.apache.http.protocol.HttpContext;
import org.apache.http.protocol.HttpRequestHandler;
import org.apache.http.util.EntityUtils;
import org.junit.AfterClass;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;

import java.io.IOException;
import java.util.Arrays;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

/**
 * Created by lanjiangang on 08/11/2017.
 */
public class GameEngineTest {

    private static GameMap map;
    private GameStateMachine stateMachine;
    private GameEngine engine;
    private static HttpServer server;
    private static String postBody = null;
    Map<String, Player> players;
    private static int localport;

    @BeforeClass
    public static void beforeClass() throws Exception {
        map = MapFactory.getMap();
        server = ServerBootstrap.bootstrap().setListenerPort(0).registerHandler("*", new HttpRequestHandler() {
            @Override
            public void handle(HttpRequest request, HttpResponse response, HttpContext context) throws HttpException, IOException {

                HttpEntity entity = null;
                if (request instanceof HttpEntityEnclosingRequest)
                    entity = ((HttpEntityEnclosingRequest) request).getEntity();

                // For some reason, just putting the incoming entity into
                // the response will not work. We have to buffer the message.
                byte[] data;
                if (entity == null) {
                    data = new byte[0];
                } else {
                    data = EntityUtils.toByteArray(entity);
                }

                postBody = new String(data);
                System.out.println(postBody);
            }
        }).create();
        server.start();
        localport = server.getLocalPort();
        System.out.println("Start http server on " + localport);

    }

    @AfterClass
    public static void afterClass() {
        server.shutdown(100, TimeUnit.MILLISECONDS);
    }

    @Before
    public void setup() {
        String[] args = new String[] { "", "4", "1", "2", "2", "1", "1", "100", "2000", "playerA", "playerB" };

        engine = new GameEngine();
        engine.parseArgs(args);
        engine.setMap(map);

        stateMachine = new GameStateMachine(engine.generateTanks(), map);
        stateMachine.setOptions(engine.getGameOptions());

        players = new LinkedHashMap<>();
        players.put("playerA", new Player("playerA", Arrays.asList(new Integer[] { 1 })));
        players.put("playerB", new Player("playerB", Arrays.asList(new Integer[] { 2 })));
        stateMachine.setPlayers(players);

        engine.setStateMachine(stateMachine);
    }

    @Test
    public void testGenerateFlag() {
        for (int i = 0; i < engine.getGameOptions().getMaxRound(); i++) {
            engine.checkGenerateFlag(i);
        }
        assertEquals(engine.getGameOptions().getNoOfTanks(), engine.getNoOfFlagGenerated());
    }

    @Test
    public void testReportRes() {
        engine.setEnv(new GameEngine.Environment() {
            @Override
            public String get(String name) {
                return "http://localhost:" + localport;
            }
        });
        engine.calculateResult(100);
        engine.reportResult();
        assertTrue(postBody.contains("draw"));

        players.get("playerA").captureFlag(null);
        engine.calculateResult(100);
        engine.reportResult();
        assertTrue(postBody.contains("win"));
        assertTrue(postBody.contains("playerA"));
    }

    @Test
    public void testGenerateTanks() {
        for(int i = 1; i < 5; i++) {
            engine.getGameOptions().setNoOfTanks(i);
            Map<Integer, Tank> tanks = engine.generateTanks();
            assertEquals(i*2, tanks.size());
        }
    }
}