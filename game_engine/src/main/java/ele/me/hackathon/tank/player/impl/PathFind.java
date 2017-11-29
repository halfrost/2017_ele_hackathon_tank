package ele.me.hackathon.tank.player.impl;

import ele.me.hackathon.tank.Direction;
import ele.me.hackathon.tank.GameMap;
import ele.me.hackathon.tank.GameState;
import ele.me.hackathon.tank.Position;

import java.util.LinkedList;
import java.util.List;

/**
 * Created by lanjiangang on 17/11/2017.
 */
public class PathFind {
    private GameMap map;

    public PathFind(GameMap map) {
        this.map = map;
    }

    public Position[] find(GameState state, Position src, Position dest) {
        List<Direction> suggestedDirs = suggestDir(src, dest);
        return null;
    }

    private List<Direction> suggestDir(Position src, Position dest) {
        List<Direction> dirs = new LinkedList<>();
        int deltaX = src.getX() - dest.getX();
        int deltaY = src.getY() - dest.getY();

        if (Math.abs(deltaX) > Math.abs(deltaY)) {

        }
        return null;
    }
}
