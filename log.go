package main

import "bufio"
import "strings"
import "net"
import "fmt"
import "path/filepath"
import "os"
import "strconv"
import "log"

const ROOT = "/tmp/"
const GB = 1024*1024*1024

type Logger struct {
    tag string
    index int
    file *os.File
    ch chan string
}

var tags = map[string]*Logger{}

func NewLogger(tag string) *Logger {
    err := os.Mkdir(ROOT + "/" + tag, 0777)
    if err != nil && !os.IsExist(err) {
        panic("mkdir")
    }
    pattern := ROOT + "/" + tag + "/log.*"
    matches, err := filepath.Glob(pattern)
    if err != nil {
        panic("glob")
    }

    max_index := 1

    for i := 0; i < len(matches); i++ {
        path := matches[i]
        index, err := strconv.Atoi(filepath.Ext(path)[1:])
        if err != nil {
            log.Println("invalid path:", path)
            continue
        }
        if index > max_index {
            max_index = index
        }
    }

    logger := new(Logger)
    logger.index = max_index
    logger.tag = tag
    logger.ch = make(chan string)

    path := logger.CurrentFilePath()
    file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        panic("open file:"+err.Error())
    }

    logger.file = file
    return logger
}

func (logger *Logger) FilePath(name string) string {
    return ROOT + "/" + logger.tag + "/" + name
}

func (logger *Logger) CurrentFilePath() string {
    name := fmt.Sprintf("log.%d", logger.index)
    path := logger.FilePath(name)
    return path
}

func (logger *Logger) Move() {
    err := logger.file.Close()
    if err != nil {
        panic("close file")
    }

    logger.index += 1

    path := logger.CurrentFilePath()
    file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        panic("open file")
    }
    logger.file = file
}

func (logger *Logger) WriteLine(line string) {
    pos, err := logger.file.Seek(0, os.SEEK_CUR) 
    if pos >= GB {
        logger.Move()
    }
    
    _, err = logger.file.WriteString(line)
    if err != nil {
        panic("disk write")
    }
}

func (logger *Logger) Run() {
    ch := logger.ch
    for {
        line := <- ch
        logger.WriteLine(line)
    }
}

func handle_client(conn *net.TCPConn) {
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
        logger, present := tags[tag]

        if !present {
            continue
        }
        l := line[index+1:]

        log.Printf("tag:%s line:%s", tag, l)
        logger.ch <- l
    }
}

func init_tags() {
    tags["application"] = NewLogger("application")
}

func main() {
    log.SetFlags(log.Lshortfile|log.LstdFlags)
    
    init_tags()
    for _, t := range tags {
        go t.Run()
    }

    ip := net.ParseIP("0.0.0.0")
    addr := net.TCPAddr{ip, 24000, ""}
    listen, err := net.ListenTCP("tcp", &addr);
    if err != nil {
        fmt.Println("初始化失败", err.Error())
        return
    }
    for {
        client, err := listen.AcceptTCP();
        if err != nil {
            return
        }
        go handle_client(client)
    }
}
