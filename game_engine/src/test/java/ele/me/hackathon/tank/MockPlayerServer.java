package ele.me.hackathon.tank;

import ele.me.hackathon.tank.player.Args;
import ele.me.hackathon.tank.player.GameState;
import ele.me.hackathon.tank.player.Order;
import ele.me.hackathon.tank.player.PlayerServer;
import org.apache.thrift.TException;
import org.apache.thrift.server.TServer;
import org.apache.thrift.server.TSimpleServer;
import org.apache.thrift.transport.TServerSocket;
import org.apache.thrift.transport.TServerTransport;
import org.apache.thrift.transport.TTransportException;

import java.util.LinkedList;
import java.util.List;

/**
 * Created by lanjiangang on 09/11/2017.
 */
public class MockPlayerServer {

    public  static  void main(String[] args) throws TTransportException {
        TServerTransport serverTransport = new TServerSocket(Integer.parseInt(args[0]));
        PlayerServer.Processor processor = new PlayerServer.Processor(new MockServer());
        TServer server = new TSimpleServer(new TServer.Args(serverTransport).processor(processor));
        server.serve();

    }

    private static class MockServer implements PlayerServer.Iface {

        private List<List<Integer>> gamemap;
        private List<Integer> tanks;
        private GameState state;

        @Override
        public void uploadMap(List<List<Integer>> gamemap) throws TException {
            this.gamemap = gamemap;
            System.out.println("Map uploaded!");
            gamemap.forEach(line -> {line.forEach(i -> System.out.print("" + i + " ")); System.out.println();});
        }

        @Override
        public void uploadParamters(Args arguments) throws TException {
            System.out.println("Arguments uploaded!");
            System.out.println(arguments);
        }

        @Override
        public void assignTanks(List<Integer> tanks) throws TException {
            this.tanks = tanks;
            System.out.println("Tanks assigned!");
            tanks.forEach(i -> System.out.print("" + i + ","));
            System.out.println();
        }

        @Override
        public void latestState(GameState state) throws TException {
            this.state = state;
            printState(state);
        }

        private void printState(GameState state) {
            System.out.println("latest game state!");
            System.out.println(PlayerInteract.toString(state));
            //state.getTanks()
        }

        @Override
        public List<Order> getNewOrders() throws TException {
            List<Order> orders = new LinkedList<>();
            //tanks.forEach(i -> orders.add(new Order(i, "fire", Direction.DOWN)));
            return orders;
        }
    }
}
