## 简明教程 & FAQ

<details>
<summary>从新项目的生成到发布上线</summary>

首先得安装好 `nex` 和相关依赖，见 [README](./README.md)。

1. 生成项目

   ```
   > nex bootstrap -appID arch.note -serviceName Note
   > # 或者使用已存在的 thrift 文件
   > nex bootstrap -appID arch.note -serviceName Note --thriftFile xxx.thrift
   > cd arch.note
   ```
   项目结构
   ```
   > tree
   .
   ├── Makefile                                // make dep 用于生成依赖, make server client 用于编译服务端和客户端相关代码
   ├── README.md                               // READ IT!!!
   ├── app.yaml                                // 项目主配置文件
   ├── arch.note_build.yml                     // 给 eless 部署用
   ├── cmds                                    // 客户端和服务端入口
   │   ├── client                              // 主要用于验证服务端功能
   │   │   └── main.go
   │   └── server
   │       └── main.go
   ├── handler                                 // 服务端相关代码
   │   ├── auto-endpoints.go
   │   ├── auto-handler.go
   │   └── note.go                             // 自行实现 thrift service
   ├── models                                  // 一个简单的 db 使用示例
   │   └── note
   │       └── models.go
   ├── rpc                                     // 远程调用相关代码
   │   └── note
   │       └── auto-client.go
   ├── services                                // thrift 生成的代码
   │   └── note
   │       ├── GoUnusedProtection__.go
   │       ├── Note-consts.go
   │       └── Note.go
   └── thriftfs                                // 用于存放 thrift 文件和声明依赖的服务
       └── Note.thrift
           └── deps.json
   ```

2. 编译运行

   解决依赖问题

   ```
   > make dep
   ```

   编译并运行：
   ```
   > make server client
   > ./bin/server -dev -local-huskar
   > ./bin/client -dev -local-huskar  # 另起一个 terminal
   ```
   更改 thrift 文件后一定要重新生成相关代码
   ```
   > vim thriftfs/Note.thrift
   > nex regen
   ```

3. nex 的配置
  
   Nex 大部分配置是从 huskar 中读取的. 为了方便本地开发, 提供了本地配置的方式. 下面的教程中
   为了简化, 都是以本地 huskar 配置的方式给各个组件添加配置.
   - config.json: 对应 huskar 上的 config, key/value 均为 string.
   - toggle.json: 对应 huskar 上的 switch, key 为 string, value 为数字, 标识开关程度.
   - service.json: 对应 huskar 上的 service, 用以模拟 huskar 的服务节点列表功能. 很少会用到.

3. 使用 MySQL

   nex 是按需初始化资源的，这里需要先将 app.yaml 里面的 `mysql: false` 改成 `mysql: true`，
   编辑 config.json ，`DB_SETTINGS` 的详细格式见https://godoc.elenet.me/github.com/eleme/nex/db#hdr-Database_configs：
   ```
   > pwd
   path/to/arch.note
   > cat config.json
   {
     "DB_SETTINGS": "{\"name\": \"master\": \"eleme:eleme@tcp(localhost:3306)/note\", \"slave\": \"eleme:eleme@tcp(localhost:3306)/note\"}",
   }
   ```
   直观的使用示例：
   ```go
   dbm := nex.GetDBManager()
   db := dbm.GetDBMaster("note")
   db.DoDBThings()
   ```
   注意不要在 `init` 函数等 package 初始化就执行的作用域里面做获取`GetDBManager`的操作，因为这时候资源可能还没初始化，
   更好的方式是自己封装一层来做 lazy init，或者在 `handler.NewNoteService()` 里面初始化：
   ```go
   type NoteDB struct {
       master *db.DB
       field FieldType
   }
   func NewNoteDB() *NoteDB {
       dbm := nex.GetDBManager()
       return &NoteDB{
           master: dbm.GetDBMaster("note"),
       }
   }
   func (noteDB *NoteDB) QueryAllNotes() ([]*Users, error) { return nil, nil }
   // 单例
   var once sync.Once
   var noteDB *NoteDB
   func GetNoteDB() *NoteDB {
       once.Do(func(){
           noteDB = NewNoteDB()
       })
       return noteDB
   }
   ```

4. 使用 Redis

   同上，将 app.yaml 里面的 `redis: false` 改为 `redis: true`，
   编辑 config.json，`REDIS_SETTINGS` 的详细格式见https://godoc.elenet.me/github.com/eleme/nex/client/redis#NewPoolManager：
   ```
   {
     ...,
     "REDIS_SETTINGS": "{\"note\": \"url\": \"localhost:6379\"}",
   }
   ```
   直观的使用示例，最好是自己封装一层然后 lazy init：
   ```go
   rps := nex.GetRedisPools()
   pc := rps.GetPooledClient()
   pc.Get("key")
   ```

5. 服务依赖和远程调用

   首先需要在 thriftfs/deps.json 里面声明服务依赖：
   ```
   [
      {
        "Name": "Note",                      // 这个是自己服务的客户端声明,一般不用更改
        "AppID": "arch.note",
        "ThriftFile": "Note.thrift",
        "Addr": ":8010",
        "PoolOptions": {
          "MaxCap": 10,
          "MaxActive": 5,
          "IdleTimeout": 30,
          "ConnectTimeout": 500,
          "RWTimeout": 500
        }
      },
      {
        "Name": "JService",
        "AppID": "me.ele.arch.jservice",
        "ThriftFile": "JService.thrift",     // thrift 文件放在当前目录下
        "IFace": "me.ele.arch.jservice.DemoService",  // Java 服务必须指定 interface
        "Addr": "localhost:8010",          // 可选，调试用
        "PoolOptions": {                   // 可选是否自定义以下配置
          "MaxCap": 10,                    // 连接池最大可扩充大小
          "MaxActive": 5,                  // 连接池实际大小
          "IdleTimeout": 30,               // 连接最长闲置时间，秒
          "ConnectTimeout": 500,           // 建立连接的最长时间，毫秒
          "RWTimeout": 500                 // 连接读写的超时时间，毫秒
        }
      }
   ]
   ```
   接下来生成代码
   ```
   > pwd
   path/to/arch.note
   > nex regen
   ```
   我们会看到 rpc 和 services 目录下多了一些文件夹，对应生成的相关代码，文件夹都是以 `Name` 来命名的，如果
   在 deps.json 中定义了 `Addr`, 则会使用定义的地址进行连接：
   ```go
   client, err := noterpc.GetThriftNoteServiceClient()
   b, err := client.Ping(context.TODO())
   ```

   如果没有定义 `Addr`, 而要使用基于 Huskar 的连接池，需要在 Huskar 上面配置 key 为 "CLUSTER:{Name}" 的配置来声明依赖服务的 cluster，
   这里我们在本地 config.json 里面来模拟：

   ```
   {
     "CLUSTER:Note": "Common"
   }
   ```
   然后需要创建配置文件 service.json 用于模拟服务发现：
   ```
   > pwd
   path/to/arch.note
   > cat service.json
   [
     {
       "key": "127.0.0.1_8010",          // 这个保证唯一就行
       "application": "arch.note",       // 依赖服务的 app id
       "cluster": "Common",              // 依赖服务的 cluster
       "state": "up",                    // 状态
       "ip": "127.0.0.1",                // IP
       "port": {"main": 8010}            // 端口，Java 服务需要指明心跳端口：{"main": 5353, "back": 5354}
     }
   ]
   ```
   config.json 里面的 cluster 对应 service.json 里面的 cluster，这样我们就可以在
   本地使用基于 Huskar 的连接池了：

   ```go
   client, err := calculatorrpc.GetThriftNoteServiceClientFromHuskar()
   b, err := client.Ping(context.TODO())
   ```

   注意运行时一定要使用 `-dev -local-huskar`. 生产环境只会使用 huskar 的配置.

6. Go CI

   见：https://github.com/eleme/goci/blob/master/README.md

7. 部署

   我们已经生成了 Eless 部署用的配置文件，eless 的使用见：http://wiki.ele.to:8090/pages/viewpage.action?pageId=39204511 。

   前面我们使用的是本地模拟的 Huskar，线上会直连 Huskar，这是在部署的时候从 eless 拿的相关配置，
   还记得我们用了 -dev 这个选项，就是因为本地没有一个由 eless 生成的配置文件 eless_env.yaml，
   在线上可自行查看相关内容，在 /data/arch.note 或者 /srv/arch.note 目录下就是 eless 上传的相关文件。

   日志由 syslog 管理，在 /data/log/app/arch.note 目录下，应用自己的日志会被分发到 app...log 里面，
   SOA 客户端的日志在 soa_client...log 里面，服务端日志在 nex...log 里面。

   进程的管理使用的是 systemd，相关使用姿势自行查阅。

</details>


<details>
<summary>app.yaml 配置文件</summary>

```
app_name: arch.note                   // app id
service: Note                         // thrift service name
addr: 0.0.0.0:8010                    // 服务绑定的地址，默认 :8010
graceful_timeout: 3                   // 关闭服务的最长缓冲时间，默认 3 秒
client_timeout: 20                    // 客户端超时时间，默认 20 分钟
max_requests_in_progress: 5000        // 在处理中的最大请求数，默认未开启
huskar:
  dial_timeout: 10                    // 连接超时
  wait_timeout: 10                    // 连接超时
  retry_delay                         // 重试间隔
plugins:
  redis: false                        // 是否初始化 redis
  db: false                           // 是否初始化 db
  statsd: true                        // 是否使用 statsd
  etrace: true                        // 是否使用 etrace
  sam: false                          // 是否使用 sam，如果安装了 Samaritan(GoProxy)，强烈推荐打开
```

</details>


<details>
<summary>有问题，寻求解决？</summary>

**文档**：https://godoc.elenet.me/github.com/eleme/nex

使用 `Slack`，最好是直接面对面沟通，在 `Slack` 上千万不要这样：
```
U: 有个问题
.10m.30m..
F: 什么问题？（追问）
U: 这里貌似有问题 <code>
.10m.30m..
F: 什么问题？（没有上下文，继续追问）
.ba.la.la.
```

**建议先理好思路，说明在什么环境下怎么使用什么时候出现了什么问题，最好能区分是自己应用的问题还是框架的问题，越详细越好。**

**非常非常欢迎有建设性的意见和改进方案！**

</details>


<details>
<summary>编译时出现接口签名不一致相关的报错？</summary>

一般情况下都是应该应用的依赖与 `nex` 的 `vendor` 里面的依赖不一致导致的，
解决方案：

1. 自动：
    ```
    > # Make sure your version is recent enough to support it
    > nex update
    ```

2. 手动

    ```
    > cd `go list -f '{{.Dir}}' github.com/eleme/nex`
    ```
    更新代码（可选）：
    ```
    > git pull
    > make install
    ```
    注意，这里比较蛋碎的是，你的 `GOPATH` 里面与 `nex` 相关的库都会
    变成和 `nex` 一致的版本，可能会 `break` 你其他不依赖 `nex` 的项目，
    当前还没找到很好的解决方案，一个变通方案是使用环境隔离的方式，如`Vagrant`：
    ```
    > godep restore
    ```

最后
```
> cd path/to/your/app
> rm -rf Godeps/ vendor
> make dep # godep save ./...
```
</details>


<details>
<summary>怎样对接口进行降级？</summary>

在 `Huskar` 的 `Switch` 配置页里面配置以接口名为键的开关。

注意接口名一定要与 `thrift` 文件里面定义的接口名一致。

降级会返回 `api is down` 的错误消息给客户端。

| 值       | 意义           |
|----------|----------------|
| 0        | 完全降级       |
| 100      | 不降级         |
| (0, 100) | 按比例随机降级 |

</details>

<details>
<summary>怎样对接口进行超时控制？</summary>

在 `Huskar` 的 `Config` 配置页中配置键为 `HARD_TIMEOUT:{APIName}`，值为整数的配置项，单位为 Millisecond (Second*1000)。

注意 `{APIName}` 为接口名，一定要与 `thrift` 文件里面定义的接口名一致。

若超时之后请求还未处理完成，服务端会返回超时信息给客户端，时间上可能会有偏差，因为 `GoRoutine` 的调度也是需要时间的。

</details>


<details>
<summary>只想用 `nex` 生成的客户端代码，但是会有一些奇怪的问题？</summary>

比如，你想把客户端的代码放在 `rpcclients` 目录下：
```
> pwd
path/to/rpcclients
```
那么你需要在这个目录下提供一个文件夹`thriftfs`，并在其下面创建一个`deps.json`的文件“
```
> tree
.
└── thriftfs
    └── deps.json
```
然后在 `deps.json` 里面声明所需的依赖服务：
```
> cat > thriftfs/deps.json
[
  {
    "Name": "Note",              // 一定要保持唯一
    "AppID": "arch.note",        // 依赖服务的 app_id
    "ThriftFile": "Note.thrift", // 依赖服务的 thrift 文件，放在 thriftfs 下面
  },
  {
    "Name": "JService",
    "AppID": "me.eleme.jservice",
    "IFace": "me.ele.arch.jservice.JService", // java 服务必须指定 interface
    "ThriftFile": "JService.thrift",
  }
]
```
生成代码：
```
> nex bootstrap --onlyClient
> tree
.
├── rpc
│   ├── note
│   │   └── auto-client.go
│   └── jservice
│       └── auto-client.go
├── services
│   ├── note
│       └── thrift_generated_files.go
│   └── jservice
│       └── thrift_generated_files.go
└── thriftfs
    ├── Note.thrift
    ├── JService.thrift
    └── deps.json
```
然后，你只能使用签名为 `NewXXXClient` 和 `NewXXXClientWithHuskar`的方法，不能使用
签名为 `GetXXXClient` 和 `GetXXXClientFromHuskar` 的方法，里面已经内建了连接池：
```go
import (
    "..."
)
trace := etrace.New(cc.NewConfigFrom(map[string]interface{}{
    "app_name":   "xxx.xxx",
    "cluster":    "Cluster",
    "ezone":      "ezone",
    "idc":        "idc",
    "etrace_uri": "http://etrace.uri",
}))
logger, err := log.GetLogger("xxx.xxx")
if err != nil {
    panic(err)
}
configer, err := config.New(structs.Config{
        Endpoint:    "huskar_url",
        Token:       "huskar_token",
        Service:     "arch.note",
        Cluster:     "Common",
        DialTimeout: 1 * time.Second,
        WaitTimeout: 5 * time.Second,
        RetryDelay:  1 * time.Second,
    })
if err != nil {
    panic(err)
}
client := noterpc.NewThriftNoteServiceClientFromHuskar(
    noterpc.ThriftNoteServiceClientOptions{
        AppName:        "xxx.xxx",
        ConnectTimeout: 1 * time.Second,
        RWTimeout:      1 * time.Second,
        IdleTimeout:    30 * time.Second,
        MaxActive:      30,
        MaxCap:         30,
        Logger:         logger,
        Trace:          trace,
        CircuitBreaker: circuitbreaker.New("note", noterpc.CliAppErrTypes),
        HuskarConfiger: configer,
    })
```
更新 `thriftfs` 下面的 `thrift` 文件之后需要重新生成代码：
```
> pwd
path/to/rpcclients
> nex regen
```
</details>

<details>
<summary>如何传递多活相关信息？</summary>

meta 信息必须是 `map[string]string` 类型的。

```go
import (
    "context"
    tracker "github.com/eleme/thrift-tracker"
)

client, err := calculatorrpc.GetThriftNoteServiceClientFromHuskar()
ctx := context.Background()  // or existing context
meta := map[string]string{"routing-key": "xxx"}
ctx = context.WithValue(ctx, tracker.CtxKeyRequestMeta, meta)
err = client.AddNote(ctx, "good note")
```
</details>


<details>
<summary>为啥RPC 调用出现错误，却无法打印错误消息？</summary>

这是由于各种 `thrift` 库的实现不一致导致的，`nex` 提供了一个工具方法：

```go
import "github.com/eleme/nex/utils"

if err := rpcCall(); err != nil {
    fmt.Println(utils.DefaultThriftErrorMessage(err))
}
```

</details>


<details>
<summary>如何进行 profiling？</summary>

nex 已经默认开启了 `pprof` HTTP 服务，端口号为：`4455`

| URL                    | Content          |
|------------------------|------------------|
| `/debug/pprof/`        | pprof.Index      |
| `/debug/pprof/cmdline` | pprof.Cmdline    |
| `/debug/pprof/profile` | pprof.Profile    |
| `/debug/pprof/symbol`  | pprof.Symbol     |
| `/debug/pprof/trace`   | pprof.Trace      |
| `/debug/vars`          | expvar.Handler() |

对于线上 docker 容器里面运行的 nex 应用可以在本地通过 `http://pprof.tools.elenet.me/<container IP>:4455/debug/pprof/`
来访问。

详细使用方式自行查阅官方文档。

</details>

<details>
<summary>不是 thrift 的服务，但是想用有 ETrace 功能的相关组件？</summary>

```go
import (
    "context"

    "github.com/eleme/nex"
    "github.com/eleme/nex/consts/ctxkeys"
    etrace "github.com/eleme/etrace-go"
    tracker "github.com/eleme/thrift-tracker"
)

// use func init here for example..
func init() {
    nex.Init()
}

func CtxWithTraceTransaction(ctx context.Context, appID string) (etrace.Transactioner, context.Context) {
    trace := nex.GetETrace()
    name := fmt.Sprintf("%v.HTTP", appID) // 这个随意，对自己有意义就行
    root, ctx = trace.NewTransaction(ctx, etrace.TypeService, name)
    ctx = context.WithValue(ctx, ctxkeys.EtraceTransactioner, root)
    return root, ctx
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
    // 如果有 rpc id 和 request id 等信息就放在 context 里面，没有的话直接忽略
    // 不知道 context 怎么用，千万不要用 nil 作参数，用 context.TODO()
    ctx := context.Background()
    ctx = context.WithValue(ctx, tracker.CtxKeyRequestID, "request id")
    ctx = context.WithValue(ctx, tracker.CtxKeySequenceID, "rpc id")
    ctx = context.WithValue(ctx, tracker.CtxKeyRequestMeta, map[string]string{"routing-key": "xxx"})

    root, ctx := CtxWithTraceTransaction(ctx, "whatever")

    var err error
    defer func() {
        root.AddTag("URL", r.URL.Path) // 可以加一些有意义的 Tag
        if err != nil {
            root.SetStatus(err.Error())
        } else {
            root.SetStatus(etrace.StatusSuccess)
        }
        root.Commit()
    }()

    // 传递含有 trace transaction 的 context，db 和 redis 组件才会打点
    _ = nex.GetDBManager().GetDBMaster("whatever").QueryRowContext(ctx, "select * from nowhere")
    _, err = nex.GetRedisPools().GetPooledClient("whatever").Get(ctx, "nokey")

    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}
```

另外一种方式是: https://github.com/eleme/etrace-go

</details>

<details>
<summary>如何给 thrift 服务端添加自定义的 middleware？</summary>

在 `package handler` 里面给 `ExtMiddlewareChain` 这个变量赋 `endpoint.Chain` 类型的值即可。

示例：app.id/handler/middlewares.go:

```go
package handler

import (
    "context"
    "fmt"

    "github.com/eleme/nex/endpoint"
)


func init() {
    ExtMiddlewareChain = endpoint.Chain(
        customMiddlewareA("-----"),
        customMiddlewareB,
    )
}

func customMiddlewareA(xx string) endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (response interface{}, err error) {
            fmt.Printf("enter custom middleware AAAAAAAAAAA: %v\n", xx)
            defer fmt.Printf("exit custom middleware AAAAAAAAAAA: %v\n", xx)
            return next(ctx, request)
        }
    }
}

func customMiddlewareB(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        fmt.Println("enter custom middleware BBBBBBBBBBB")
        defer fmt.Println("exit custom middleware BBBBBBBBBBB")
        return next(ctx, request)
    }
}
```

</details>

<details>
<summary>MySQL(DAL) 莫名其妙的报错？</summary>

错误信息：`[DAL]useServerPrepStmts=true is not allowed in DAL, please remove this parameter!`

这是因为使用了 `prepare statement`，`DAL` 不支持。

**一般情况下是不会发生的，由于 MySQL 的 driver 没有很好的提示信息，极有可能是参数个数和占位符个数不一致导致的，请认真仔细的数三遍..**

</details>


<details>
<summary>怎样查看相关打点？</summary>

详见：http://docs.zoo.elenet.me/zeus_core_doc/design/metrics.html#design-arch-metrics

</details>


<details>
<summary>怎样调试 thrift 接口？</summary>
使用任意的 thrift 客户端即可进行调试.


nex 内部提供了一个写 client 的示例(见 bootstrap 出来的 cmds/client/main.go 文件). 你也可以
使用公司提供的其他语言的 SOA 框架进行测试.


如果只是想简单地, 以命令行交互形式的测试, 可以使用 [thriftpy-cli](https://github.com/wooparadog/thriftpy-cli) 工具.
</details>
