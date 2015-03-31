package main

import (
    "bufio"
    "strings"
    "net"
    "fmt"
    "os"
    "strconv"
    "github.com/jimlawless/cfg"
    log "github.com/golang/glog"
    "sync"
    "flag"
)

var ROOT = "/tmp/"
var PORT = 24000
var BIND_ADDR = ""

var day_tags = map[string]*DailyLogger{}
var lock sync.Mutex

func handle_client(conn net.Conn) {
    reader := bufio.NewReader(conn)

    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            break
        }

        index := strings.IndexByte(line, ':')
        if index == -1 {
            continue
        }
        tag := line[0:index]

        day_logger, day_present := day_tags[tag]

        if !day_present {
            lock.Lock()
            day_tags[tag] = NewDailyLogger(tag)
            go day_tags[tag].Run()
            lock.Unlock()
            day_logger = day_tags[tag]
        }

        l := line[index+1:]

        log.Infof("tag:%s line:%s", tag, l)
        if day_logger != nil {
            day_logger.ch <- l
        }
    }
}

func read_cfg(filename string) {
    app_cfg := make(map[string]string)
    err := cfg.Load(filename, app_cfg)
    if err != nil {
        log.Fatal(err)
    }
    root, present := app_cfg["root"]
    if !present {
        fmt.Println("need config root directory")
        os.Exit(1)
    }
    ROOT = root

    port, present := app_cfg["port"]
    if !present {
        fmt.Println("need config listen port")
        os.Exit(1)
    }
    nport, err := strconv.Atoi(port)
    if err != nil {
        fmt.Println("need config listen port")
        os.Exit(1)
    }
    PORT = nport
    fmt.Printf("root:%s port:%d\n", ROOT, PORT)

    if _, present = app_cfg["bind_addr"]; present {
        BIND_ADDR = app_cfg["bind_addr"]
    }
    fmt.Printf("root:%s bind addr:%s port:%d\n", ROOT, BIND_ADDR, PORT)

}

func main() {
    flag.Parse()

    var filename string
    if len(flag.Args()) == 0 {
        filename = "log.cfg"
        if _, err := os.Stat(filename); err != nil {
            filename = ""
        }
    } else {
        filename = flag.Args()[0]
    }

    if filename != "" {
        read_cfg(filename)
    }

    addr := fmt.Sprintf("%s:%d", BIND_ADDR, PORT)
    Serve(addr, handle_client)
}
