1. 首先，安装 docker，官方给的安装连接： https://docs.docker.com/v17.06/docker-for-mac/install/，具体参考文档：http://wiki.ele.to:8090/pages/viewpage.action?pageId=63669236；

2. 如果没有在本地登陆过，需要先执行如下命令按提示输入账号密码进行登录：

```
docker login docker-hack.ele.me
```

2. 将 docker 中要用的 server 程序编译好并放到 server 目录下，名字也叫 server 就行；

3. cd 到 TankBattle 仓库根目录，执行如下命令编译 docker 镜像（注意将命令中的「版本号」替换为你要的）：

```
docker build -t docker-hack.ele.me/dezhi.yu/meituanpurchaser:版本号 .
```

4. 执行如下命令将本地编译的 docker 镜像推送到远端：

docker push docker-hack.ele.me/dezhi.yu/meituanpurchaser:版本号

5. 这时候去 docker hub 应该能看到刚才推送上去的 docker，在这个页面：https://docker-hack.ele.me/Repos/TagList?repoId=226

然后需要提交的 docker 地址就是 `docker-hack.ele.me/dezhi.yu/meituanpurchaser:版本号`，直接提交就行。

6. 其他命令

查看本地所有 docker 镜像：docker image ls
删除本地所有 docker 镜像：docker rmi $(docker images -q) -f
