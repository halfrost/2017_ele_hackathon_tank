package ele.me.hackathon.tank;

import java.util.LinkedList;
import java.util.List;

/**
 * Created by lanjiangang on 07/11/2017.
 */
public class GameState {
    private String playerName;
    private List<Tank> tanks = new LinkedList<>();
    private List<Shell> shells = new LinkedList<>();
    private int yourFlagNo = 0;
    private int enmeyFlagNo = 0;
    private Position flagPos = null;

    public GameState(String playerName) {
        this.playerName = playerName;
    }

    public List<Tank> getTanks() {
        return tanks;
    }

    public List<Shell> getShells() {
        return shells;
    }

    public String getPlayerName() {
        return playerName;
    }

    public int getYourFlagNo() {
        return yourFlagNo;
    }

    public void setYourFlagNo(int yourFlagNo) {
        this.yourFlagNo = yourFlagNo;
    }

    public int getEnmeyFlagNo() {
        return enmeyFlagNo;
    }

    public void setEnmeyFlagNo(int enmeyFlagNo) {
        this.enmeyFlagNo = enmeyFlagNo;
    }

    public Position getFlagPos() {
        return flagPos;
    }

    public void setFlagPos(Position flagPos) {
        this.flagPos = flagPos;
    }

    @Override
    public String toString() {
        return "GameState{" +
                "playerName='" + playerName + '\'' +
                ", tanks=" + tanks +
                ", shells=" + shells +
                ", yourFlagNo=" + yourFlagNo +
                ", enmeyFlagNo=" + enmeyFlagNo +
                ", flagPos=" + flagPos +
                '}';
    }
}
