package ele.me.hackathon.tank;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;

/**
 * Created by lanjiangang on 27/10/2017.
 */
public class GameMap {
    int size = 0;
    int[][] pixels;

    public GameMap(int size, int[][] pixels) {
        this.size = size;
        this.pixels = pixels;
    }

    public boolean isBarrier(Position pos) {
        return pixels[pos.getX()][pos.getY()] == 1;
    }

    public boolean isVisible(Position pos) {
        return pixels[pos.getX()][pos.getY()] != 2;
    }

    public static GameMap load(InputStream in) throws IOException {
        BufferedReader reader = new BufferedReader(new InputStreamReader(in));
        int size = readSize(reader);
        int[][] pixels = readMap(reader, size);

        return new GameMap(size, pixels);
    }

    private static int readSize(BufferedReader reader) throws IOException {
        for(String line = reader.readLine(); line != null; line = reader.readLine()){
            if(line.startsWith("//"))
                continue;

            if(line.startsWith("size")){
                return Integer.parseInt(line.split(":")[1].trim());
            }
        }
        throw new InvalidMapFile("no size found!");
    }

    private static int[][] readMap(BufferedReader reader, int size) throws IOException {
        int[][] map = new int[size][size];
        for (int i = 0; i < size; i++){
            map[i] = readLine(reader, size);
        }
        return map;
    }

    private static int[] readLine(BufferedReader reader, int size) throws IOException {
        for(String line = reader.readLine(); line != null; line = reader.readLine()){
            if(line.startsWith("//")) {
                continue;
            }

            String[] row = line.split(" ");
            if(row.length != size) {
                throw new InvalidMapFile("expect map size is " + size + ", but read a line sized " + row.length + ", the line is: " + line);
            }

            int[] nums = new int[size];
            for (int i = 0; i < size; i++){
                nums[i] = Integer.parseInt(row[i]);
            }
            return nums;
        }
        throw new InvalidMapFile("No enough lien read");
    }

    public int size() {
        return size;
    }

    public int[][] getPixels() {
        return pixels;
    }

}
