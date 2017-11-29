package ele.me.hackathon.tank;

import ele.me.hackathon.tank.player.PlayerServer;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.InputStreamEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.transport.TSocket;
import org.apache.thrift.transport.TTransportException;

import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.stream.Collectors;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class GameEngine {
    private String mapFile;
    private GameMap map;
    private GameStateMachine stateMachine;
    private String playerAAddres;
    private String playerBAddres;
    private boolean flagGenerated = false;
    private int noOfFlagGenerated = 0;

    private Map<String, PlayerServer.Client> clients = new ConcurrentHashMap<>();
    private Map<String, Player> players;
    private GameOptions gameOptions;
    private Environment env = new Environment();

    GameResult result = new GameResult();

    public static class Environment {
        public String get(String name) {
            return System.getenv(name);
        }
    }

    public static void main(String[] args) throws TTransportException, InterruptedException {
        System.out.println("Game starts at " + LocalDateTime.now());
        GameEngine engine = new GameEngine();
        engine.parseArgs(args);
        engine.startGame();
        System.out.println("Game ends at " + LocalDateTime.now());
        //do not exit the main thread cause the judge process needs me to alive
        Thread.sleep(1000 * 60 * 10);
    }

    public GameEngine() {
    }

    void parseArgs(String[] args) {
        mapFile = args[0];
        int noOfTanks = Integer.parseInt(args[1]);
        if (noOfTanks > 5) {
            System.out.println("Max no of tanks is 5");
            System.exit(-1);
        }
        int tankSpeed = Integer.parseInt(args[2]);
        int shellSpeed = Integer.parseInt(args[3]);
        int tankHP = Integer.parseInt(args[4]);
        int tankScore = Integer.parseInt(args[5]);
        int flagScore = Integer.parseInt(args[6]);
        int maxRound = Integer.parseInt(args[7]);
        int roundTimeout = Integer.parseInt(args[8]);
        playerAAddres = args[9];
        playerBAddres = args[10];
        this.gameOptions = new GameOptions(noOfTanks, tankSpeed, shellSpeed, tankHP, tankScore, flagScore, maxRound, roundTimeout);
        System.out.println("Parameters parsed. " + this.gameOptions);
    }

    private void startGame() throws TTransportException {
        initGameStateMachine();
        connectToPlayers();
        play();
    }

    private void initGameStateMachine() {
        map = loadMap(mapFile);
        printMap(map);
        Map<Integer, Tank> tanks = generateTanks();
        this.players = assignTankToPlayers(tanks);

        stateMachine = new GameStateMachine(tanks, map);
        stateMachine.setOptions(gameOptions);
        stateMachine.setPlayers(players);
    }

    private void printMap(GameMap map) {
        StringBuilder sb = new StringBuilder("const MAP_1 = '");
        for(int[] line : map.getPixels()) {
            for(int p : line) {
                sb.append(p).append(" ");
            }
            sb.deleteCharAt(sb.length() - 1);
            sb.append("\n");
        }
        sb.deleteCharAt(sb.length() - 1);
        sb.append("'");
        System.out.println(sb.toString());
    }

    private Map<String, Player> assignTankToPlayers(Map<Integer, Tank> tanks) {
        Map<String, Player> players = new LinkedHashMap<>();

        players.put(playerAAddres, new Player(playerAAddres,
                tanks.keySet().stream().filter(id -> id <= getGameOptions().getNoOfTanks()).collect(Collectors.toCollection(LinkedList::new))));
        players.put(playerBAddres,
                new Player(playerBAddres, tanks.keySet().stream().filter(id -> id > getGameOptions().getNoOfTanks()).collect(Collectors.toCollection(LinkedList::new))));
        return players;
    }

    protected Map<Integer, Tank> generateTanks() {
        Position[] burnPoss = new Position[] { new Position(1, 1), new Position(1, 2), new Position(2, 1), new Position(2, 2), new Position(1, 3) };
        Map<Integer, Tank> tanks = new LinkedHashMap<>();
        int mapSize = map.size();
        for (int i = 0; i < gameOptions.getNoOfTanks(); i++) {
            Position pos = burnPoss[i];
            tanks.put(i + 1,
                    new Tank(i + 1, pos, Direction.DOWN, getGameOptions().getTankSpeed(), getGameOptions().getShellSpeed(), getGameOptions().getTankHP()));
            tanks.put(i + 1 + gameOptions.getNoOfTanks(),
                    new Tank(i + 1 + gameOptions.getNoOfTanks(), new Position(mapSize - pos.getX() - 1, mapSize - pos.getY() - 1), Direction.UP,
                            getGameOptions().getTankSpeed(), getGameOptions().getShellSpeed(), getGameOptions().getTankHP()));
        }
        return tanks;
    }

    private void play() {
        List<PlayerInteract> actors = Arrays.asList(new String[] { playerAAddres, playerBAddres }).stream().map(name -> buildPlayerInteract(name, gameOptions))
                .collect(Collectors.toList());
        Map<String, LinkedBlockingQueue<List<TankOrder>>> tankOrderQueues = actors.stream()
                .collect(Collectors.toMap(PlayerInteract::getAddress, act -> act.getCommandQueue()));
        Map<String, LinkedBlockingQueue<GameState>> stateQueues = actors.stream()
                .collect(Collectors.toMap(PlayerInteract::getAddress, act -> act.getStatusQueue()));

        actors.forEach(act -> act.start());

        //print the init state
        stateMachine.printReplayLog();

        //send a singal tp upload map and tank list
        stateQueues.values().forEach(q -> q.offer(new GameState("fakeState")));
        int round = 0;
        for (; round < getGameOptions().getMaxRound(); round++) {
            System.out.println("Round " + round);
            //clear the command queue to prevent previous dirty command left in the queue
            tankOrderQueues.values().forEach(q -> q.clear());

            Map<String, GameState> latestState = stateMachine.reportState();
            latestState.entrySet().forEach(k -> stateQueues.get(k.getKey()).offer(k.getValue()));

            List<TankOrder> orders = new LinkedList<>();
            tankOrderQueues.values().forEach(q -> {
                try {
                    orders.addAll(q.take());
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }
            });

            stateMachine.newOrders(orders);

            if (stateMachine.gameOvered()) {
                break;
            }

            checkGenerateFlag(round);
        }

        //print the final state
        stateMachine.printReplayLog();

        calculateResult(round);
        reportResult();
    }

    void calculateResult(int round) {
        Map<String, Integer> scores;
        if (round < getGameOptions().getMaxRound()) {
            scores = stateMachine.countScore(getGameOptions().getTankScore(), 0);
        } else {
            scores = stateMachine.countScore(getGameOptions().getTankScore(), getGameOptions().getFlagScore());
        }

        if (scores.get(playerAAddres) > scores.get(playerBAddres)) {
            result.setResult("win");
            result.setWin(playerAAddres);
        } else if (scores.get(playerAAddres) == scores.get(playerBAddres)) {
            result.setResult("draw");
        } else {
            result.setResult("win");
            result.setWin(playerBAddres);
        }
        result.setState(playerAAddres + ": " + scores.get(playerAAddres) + "," + playerBAddres + ": " + scores.get(playerBAddres));

        System.out.println("Game result: " + Util.toJson(result));
    }

    void reportResult() {
        String resUrl = env.get("WAR_CALLBACK_URL");
        System.out.println("WAR_CALLBACK_URL=" + resUrl);

        HttpPost post = new HttpPost(resUrl);
        post.setEntity(new InputStreamEntity(new ByteArrayInputStream(Util.toJson(result).getBytes())));
        CloseableHttpClient httpclient = HttpClients.createDefault();
        try {
            CloseableHttpResponse response = httpclient.execute(post);
            System.out.println(response.toString());
        } catch (IOException e) {
            System.out.println("Failed to send!");
            e.printStackTrace();
        }
        System.out.println("Result sent.");
    }

    protected void checkGenerateFlag(int round) {
        if (flagGenerated == false) {
            //generate if has past half rounds and no tank is lost
            if (round > (gameOptions.getMaxRound() / 2 - 1) && stateMachine.getTankList().size() == 2 * gameOptions.getNoOfTanks()) {
                System.out.println("Start to generate flag.");
                flagGenerated = true;
                stateMachine.generateFlag();
                noOfFlagGenerated++;
            }
        } else {
            //after first time, generate the flag repeatly but no more than one player's number of tanks
            if ((round - gameOptions.getMaxRound() / 2) % (gameOptions.getMaxRound() / 2 / gameOptions.getNoOfTanks() + 1) == 0) {
                stateMachine.generateFlag();
                noOfFlagGenerated++;
            }
        }

    }

    private PlayerInteract buildPlayerInteract(String name, GameOptions gameOptions) {
        return new PlayerInteract(name, clients.get(name), map, players.get(name).getTanks(), this.gameOptions);
    }

    private class PlayerConnector implements Runnable {
        private String name;

        public PlayerConnector(String name) {
            this.name = name;
        }

        @Override
        public void run() {
            for (int i = 0; i < 300; i++) {
                try {
                    clients.put(this.name, buildPlayerConnection(this.name));
                    break;
                } catch (TTransportException e) {
                    System.out.println("Failed to connect to " + this.name);
                    e.printStackTrace();
                }
                try {
                    Thread.sleep(1000);
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }
            }

        }
    }

    private void connectToPlayers() {
        
        Thread[] threads = new Thread[2];
        threads[0] = new Thread(new PlayerConnector(playerAAddres));
        threads[1] = new Thread(new PlayerConnector(playerBAddres));

        threads[0].start();
        threads[1].start();
        try {
            threads[0].join(60*1000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        try {
            threads[1].join(60*1000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        if (this.clients.size() < 2) {

            if (this.clients.size() == 0) {
                result.setResult("draw");
                result.setReason("Failed to connect to both players.");
            } else {
                result.setResult("win");
                String name = this.clients.keySet().stream().findFirst().get();
                result.setWin(name);
                result.setReason("Only connected to " + name);
            }

            System.out.println(Util.toJson(result));
            reportResult();
            //do not exit the main thread cause the judge process needs me to alive
            try {
                Thread.sleep(1000 * 60 * 10);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }

            System.exit(-1);
        }

        System.out.println("Succeed to connect to both player.");
    }

    private PlayerServer.Client buildPlayerConnection(String addr) throws TTransportException {
        System.out.println("Connecting to " + addr);
        String host = addr.split(":")[0];
        int port = Integer.parseInt(addr.split(":")[1]);
        TSocket transport = new TSocket(host, port);
        transport.open();
        transport.setTimeout(getGameOptions().getRoundTimeout());
        TProtocol protocol = new TBinaryProtocol(transport);
        PlayerServer.Client client = new PlayerServer.Client(protocol);
        System.out.println("Succeed to connect to  " + addr);
        return client;
    }

    private GameMap loadMap(String fileName) {
        try {
            return GameMap.load(new FileInputStream(new File(fileName)));
        } catch (IOException e) {
            throw new RuntimeException("failed to load map file : " + fileName);
        }
    }

    public void setStateMachine(GameStateMachine stateMachine) {
        this.stateMachine = stateMachine;
    }

    public GameOptions getGameOptions() {
        return gameOptions;
    }

    public int getNoOfFlagGenerated() {
        return noOfFlagGenerated;
    }

    public void setMap(GameMap map) {
        this.map = map;
    }

    public void setEnv(Environment env) {
        this.env = env;
    }
}
