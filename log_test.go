package main

import "testing" 
import "os"

func Test_WriteLine(t *testing.T) {
    os.RemoveAll(ROOT+"/app")
    logger := NewLogger("app")
    logger.WriteLine("test\n")
    f, _ := os.Open(logger.FilePath("log.1"))
    buf := make([]byte, 5)
    f.Read(buf)
    if string(buf) != "test\n" {
        t.Fail()
    }
}

func touch(path string) {
    f, _ := os.Create(path)
    f.Close()
}

func Test_Move(t *testing.T) {
    os.RemoveAll(ROOT+"/app")
    os.Mkdir(ROOT + "/app", 0777)
    touch(ROOT + "/app/log.1")
    touch(ROOT + "/app/log.2")
    logger := NewLogger("app")
    logger.Move()
    _, err := os.Stat(ROOT+"/app/log.3")
    if err != nil {
        t.Fail()
    }
}
