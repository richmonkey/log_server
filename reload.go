package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
)

const (
    Graceful = "graceful"
    FD = 3
)

// Test whether an error is equivalent to net.errClosing as returned by
// Accept during a graceful exit.
func IsErrClosing(err error) bool {
    if opErr, ok := err.(*net.OpError); ok {
        err = opErr.Err
    }
    return "use of closed network connection" == err.Error()
}


func WaitSignal(l net.Listener) error {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGTERM, syscall.SIGHUP)
    for {
        sig := <-ch
        log.Println(sig.String())
        switch sig {

            case syscall.SIGTERM:
            return nil
            case syscall.SIGHUP:
            Restart(l)
            return nil
        }
    }
    return nil // It'll never get here.
}

func Restart(l net.Listener) {
    listenTCP, ok := l.(*net.TCPListener)
    if !ok {
        log.Fatal("File descriptor is not a valid TCP socket")
    }
    file, err := listenTCP.File() // this returns a Dup()

    if nil != err {
        log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
    }

    path, err := exec.LookPath(os.Args[0])
    if nil != err {
        log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
    }

    cmd := exec.Command(path, os.Args[1:]...)
    cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", Graceful, 1))
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    // entry i becomes file descriptor 3+i.
    cmd.ExtraFiles = []*os.File{file}

    err = cmd.Start()
    if err != nil {
        log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
    }
}

func Serve(laddr string, handler func(net.Conn)) {
    var l net.Listener
    var err error

    graceful := os.Getenv(Graceful)
    if graceful != "" {
        // entry 0 becomes file descriptor 3.
        log.Printf("main: Listening to existing file descriptor %v.", FD)
        f := os.NewFile(uintptr(FD), "")
        // file listener dup fd
        l, err = net.FileListener(f)
        // close file descriptor 3
        f.Close()
    } else {
        log.Print("main: Listening on a new file descriptor.")
        l, err = net.Listen("tcp", laddr)
    }

    if err != nil {
        log.Fatalf("start fail: %v", err)
    }

    go func() {
        serve(l, handler)
    }()

    WaitSignal(l)
}

func serve(l net.Listener, handle func(net.Conn)) {
    for {
        c, err := l.Accept()
        if nil != err {
            if IsErrClosing(err) {
                break
            }
            log.Fatalln(err)
        }

        go handle(c)
    }
}