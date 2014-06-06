package main

import "os"
import "fmt"
import "time"

type DailyLogger struct {
    tag string
    date string
    file *os.File
    ch chan string
}

func NewDailyLogger(tag string) *DailyLogger {
    err := os.Mkdir(ROOT + "/" + tag, 0777)
    if err != nil && !os.IsExist(err) {
        panic("mkdir")
    }

    logger := new(DailyLogger)
    logger.date = Date()
    logger.tag = tag
    logger.ch = make(chan string)

    logger.OpenFile()

    return logger
}

func Date() string {
    year, month, day := time.Now().Date()
    return fmt.Sprintf("%d%02d%02d", year, month, day)
}

func (logger *DailyLogger) CurrentFilePath() string {
    path := fmt.Sprintf("%s/%s/%s/log.1", ROOT, logger.tag, logger.date)
    return path
}

func (logger *DailyLogger) OpenFile() {
    dir := fmt.Sprintf("%s/%s/%s", ROOT, logger.tag, logger.date)
    err := os.Mkdir(dir, 0777)
    if err != nil && !os.IsExist(err) {
        panic("mkdir")
    }

    path := logger.CurrentFilePath()
    file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        panic("open file")
    }
    logger.file = file
}

func (logger *DailyLogger) Move() {
    err := logger.file.Close()
    if err != nil {
        panic("close file")
    }

    logger.OpenFile()
}


func (logger *DailyLogger) WriteLine(line string) {
    date := Date()
    if date != logger.date {
        logger.date = date
        logger.Move()
    }
    
    _, err := logger.file.WriteString(line)
    if err != nil {
        panic("disk write")
    }
}

func (logger *DailyLogger) Run() {
    ch := logger.ch
    for {
        line := <- ch
        logger.WriteLine(line)
    }
}
