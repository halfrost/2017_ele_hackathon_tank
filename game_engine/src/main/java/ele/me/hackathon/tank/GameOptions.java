package ele.me.hackathon.tank;

/**
 * Created by lanjiangang on 09/11/2017.
 */
public class GameOptions {
    private int noOfTanks;
    private int tankSpeed;
    private int shellSpeed;
    private int tankHP;
    private int tankScore;
    private int flagScore;
    private int maxRound;
    private int roundTimeout;

    public GameOptions(int noOfTanks, int tankSpeed, int shellSpeed, int tankHP, int tankScore, int flagScore, int maxRound, int roundTimeout) {
        this.noOfTanks = noOfTanks;
        this.tankSpeed = tankSpeed;
        this.shellSpeed = shellSpeed;
        this.tankHP = tankHP;
        this.tankScore = tankScore;
        this.flagScore = flagScore;
        this.maxRound = maxRound;
        this.roundTimeout = roundTimeout;
    }

    public int getNoOfTanks() {
        return noOfTanks;
    }

    public int getTankSpeed() {
        return tankSpeed;
    }

    public int getShellSpeed() {
        return shellSpeed;
    }

    public int getTankHP() {
        return tankHP;
    }

    public int getTankScore() {
        return tankScore;
    }

    public int getFlagScore() {
        return flagScore;
    }

    public int getMaxRound() {
        return maxRound;
    }

    public int getRoundTimeout() {
        return roundTimeout;
    }

    @Override
    public String toString() {
        return "GameOptions{" +
                "noOfTanks=" + noOfTanks +
                ", tankSpeed=" + tankSpeed +
                ", shellSpeed=" + shellSpeed +
                ", tankHP=" + tankHP +
                ", tankScore=" + tankScore +
                ", flagScore=" + flagScore +
                ", maxRound=" + maxRound +
                ", roundTimeout=" + roundTimeout +
                '}';
    }

    public void setTankSpeed(int tankSpeed) {
        this.tankSpeed = tankSpeed;
    }

    public void setNoOfTanks(int noOfTanks) {
        this.noOfTanks = noOfTanks;
    }
}
