package main

import "bufio"
import "strings"
import "net"
import "fmt"
import "os"
import "strconv"
import "log"
import "github.com/jimlawless/cfg"

var ROOT = "/tmp/"
var PORT = 24000
var BIND_ADDR = ""

var tags = map[string]*Logger{}
var day_tags = map[string]*DailyLogger{}

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

		day_logger, day_present := day_tags[tag]
		if !present && !day_present {
			continue
		}

		l := line[index+1:]

		log.Printf("tag:%s line:%s", tag, l)
		if logger != nil {
			logger.ch <- l
		}
		if day_logger != nil {
			day_logger.ch <- l
		}
	}
}

func init_tags() {
	tags["application"] = NewLogger("application")
	tags["device"] = NewLogger("device")
	tags["ng_application"] = NewLogger("ng_application")
	tags["service"] = NewLogger("service")
	day_tags["login"] = NewDailyLogger("login")
	day_tags["payment"] = NewDailyLogger("payment")
	day_tags["register"] = NewDailyLogger("register")
	day_tags["level"] = NewDailyLogger("level")
	day_tags["mission"] = NewDailyLogger("mission")
	day_tags["consumption"] = NewDailyLogger("consumption")
	day_tags["coin"] = NewDailyLogger("coin")
}

func read_cfg() {
	app_cfg := make(map[string]string)
	err := cfg.Load("log.cfg", app_cfg)
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
	read_cfg()

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	init_tags()
	for _, t := range tags {
		go t.Run()
	}
	for _, t := range day_tags {
		go t.Run()
	}

	ip := net.ParseIP(BIND_ADDR)
	addr := net.TCPAddr{ip, PORT, ""}
	listen, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println("初始化失败", err.Error())
		return
	}
	for {
		client, err := listen.AcceptTCP()
		if err != nil {
			return
		}
		go handle_client(client)
	}
}
