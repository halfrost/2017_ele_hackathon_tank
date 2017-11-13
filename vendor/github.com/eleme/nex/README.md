# nex

Nexus - A simple Thrfit server framework.

![GoCI-test](http://goci.ele.me/na/goci/eleme/nex/badge?type=test)
![GoCI-vet](http://goci.ele.me/na/goci/eleme/nex/badge?type=vet)
![GoCI-lint](http://goci.ele.me/na/goci/eleme/nex/badge?type=lint)
![GoCI-fmt](http://goci.ele.me/na/goci/eleme/nex/badge?type=fmt)

[中文简明教程 & FAQ](./START.zh.md)

## Bootstrap

Execute the following command to bootstap, and ignore the next two chaptors. Supported platforms:

* MacOS

```bash
curl -L https://goo.gl/f3rYM8 | bash
```

## Manual Install
### Git

Because G.F.W. Please use `ssh` instead of `https` to download package. Add
the following config to your `~/.gitconfig`

```
[url "git@github.com:"]
    insteadOf = https://github.com/
```

### Prerequisites

* Install the following packages:
    ```Bash
    $ brew install bison libtool automake wget pkg-config
    ```
* Download and build the modified version of thrift [complier](https://github.com/eleme/thrift/tree/tracker)

    ```Bash
    $ git clone git@github.com:eleme/thrift.git
    $ git checkout tracker
    $ ./bootstrap.sh
    $ # Only golang
    $ ./configure --without-haskell --without-java --without-php --without-nodejs --without-python --without-cpp --without-lua --without-perl --without-ruby --without-erlang --without-rust --without-c_glib
    $ make
    $ cp compiler/cpp/thrift /usr/local/bin/nex-thrift
    ```

### Install nex
#### Install with go get

Execute: `go get -u github.com/eleme/nex/cmd/nex`

#### Install from source

1. Get from github: `go get github.com/eleme/nex`
2. Install from source: ``cd `go list -f '{{.Dir}}' github.com/eleme/nex` && make install``
3. Resolve dependencies: `godep restore  # 'cd' in step 2 is required`, `go get` manually is not recommended.

## Getting started


1. Generate server project template.
    ```
    $ nex bootstrap -appID arch.note -serviceName Note  # or use a existing thrift file with --thriftFile
    ```
    And fill with your own code.

2. Modify `.thrift` file to meets your needs, or add dependency into `thriftfs/deps.json`, then regenerate code.
    ```
    $ cd arch.note
    $ nex regen
    ```
    It will generate a lots of code.
    Please observe the thrift [specifications](https://t.elenet.me/drafts/services.html).

3. Build and run.
    ```
    $ make dep # You may need to save dependency.
    $ make build
    ```
    Thrift server and client: (In separate terminal)
    ```
    $ bin/server
    $ bin/client
    ```

4. Generate RPC clients without server, you must provide a `thriftfs` directory with a file named `deps.json`, the
file content like this:
    ```
    [
      {
        "Name": "Note",               // Unique name
        "AppID": "arch.note",         // application id
        "ThriftFile": "Note.thrift",  // thrift file in "thriftfs"(current directory)
        "Addr": ":8010"               // optional
      },
      ...
    ]
    ```
    Then, run `nex bootstrap --onlyClient`/`nex regen` to generate code.


## Commands
Use `nex -h` or `nex command -h` to see more details.

## Features

* Thrift tracking.
* Metrics(Statsd/ETrace) record.
* RPC(Thrift/HTTP JSON) client with load balance.
* Service register and discovery.
* Database wrapper with ETrace and context support.
* Redis wrapper with ETrace and context support, which support all Corvus commands.
* AMQP wrapper with ETrace and context support.
* Configuration management with Huskar and Eless.
* Circuit breaker.
* API downgrading and timeout controling .
* Deploy automation with syslog and ELK support, etc.
