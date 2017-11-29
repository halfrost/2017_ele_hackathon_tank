package ele.me.hackathon.tank;

import java.util.List;

/**
 * Created by lanjiangang on 01/11/2017.
 */
public class Player {
    private String name;
    private List<Integer> tanks;
    private int noOfFlag = 0;

    public Player(String name, List<Integer> tanks) {
        this.name = name;
        this.tanks = tanks;
    }

    public List<Integer> getTanks() {
        return tanks;
    }

    public void setTanks(List<Integer> tanks) {
        this.tanks = tanks;
        tanks.sort(Integer::compare);
    }

    public int getNoOfFlag() {
        return noOfFlag;
    }

    public void captureFlag(Tank t) {
        System.out.println(name + " captures a flag by tank :" + t);
        this.noOfFlag++;
    }

    public boolean belongTo(Tank t) {
        return tanks.contains(t.getId());
    }

    public String getName() {
        return name;
    }
}
