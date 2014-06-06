package main

import "os"
import "path/filepath"
import "strconv"
import "log"
import "fmt"

const GB = 1024*1024*1024

type Logger struct {
    tag string
    index int
    file *os.File
    ch chan string
}

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
