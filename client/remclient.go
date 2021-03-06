package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	mq "github.com/abkhan/mqwrap"
	log "github.com/sirupsen/logrus"
)

type CmdReq struct {
	Who string            `json:"iAm"`
	Msg string            `json:"msg"`
	Cmd string            `json:"cmd"`
	Env map[string]string `json:"env"`
}

type RespMsg struct {
	Who  string `json:"iAm"`
	Cmd  string `json:"cmd"`
	Resp string `json:"msg"`
}

const (
	//useTopic = "amq.topic"
	useTopic = "scope.polling"
	ownName  = "remoteClientTest"
)

// Version of the service
var (
	BuildTag            = "0.1.1"       // Build version to be provided by build script
	BuildDate           = "today, haha" // Build date to be provided by build script
	startTime time.Time = time.Now()
	amqs      *mq.MQWrap
	target    = "hi"
	name      string
	msg       string
	env       map[string]string
)

func main() {

	flag.Parse()

	// read name from user
	fmt.Println("Please provide your name.")
	name = readline()

	fmt.Println("Where you want to connect?")
	printTargets()
	target = readline()

	reader(ownName, useTopic)

	// Starting alarm Service
	amqs = mq.NewMQSender(ownName, useTopic)
	amqs.ExchangeName = useTopic

	log.Infof("RemCmd as Client")
	log.Infof(">> Build Version %s, on date %s", BuildTag, BuildDate)

	//this sleep is required to avoid search timeout issue
	time.Sleep(5 * time.Second)

	env = make(map[string]string)
	for {
		printOptions()
		opt := readline()
		switch {
		case opt == "1":
			fmt.Println("Write the Env var like A=Z")
			option := readline()
			oparts := strings.Split(option, "=")
			if len(oparts) >= 2 {
				env[oparts[0]] = oparts[1]
			} else {
				log.Warnf("Env var should be Name=Value")
			}
		case opt == "2":
			env = make(map[string]string)
		case opt == "3":
			fmt.Println("Write the command to execute.")
			c := readline()
			sendCmd(c)
		case opt == "4":
			os.Exit(1)
		default:
			fmt.Println("Wrong option.\n")
		}
	}
}

func sendCmd(cmd string) {

	dqr := CmdReq{
		Who: name,
		Msg: "A cmd",
		Cmd: cmd,
		Env: env,
	}

	// send
	e := amqs.SendToRabbit(dqr, "remcmd-formill-masjid-pi", useTopic, ownName)
	if e != nil {
		log.Errorf("PublishError: %s", e.Error())
	} else {
		fmt.Println("Sent")
	}

}

func printOptions() {
	fmt.Println("\n-----------------")
	fmt.Println("#1: Add Environment Varaible")
	fmt.Println("#2: Clear Env Vars")
	fmt.Println("#3: Command")
	fmt.Println("#4: Quit")
}

func printTargets() {
	fmt.Println("\n--Targets-----------------")
	fmt.Println("Like mac/dev/etc")
	fmt.Println("\n--------------------------")
}

func readline() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\n-> ")
	text, _ := reader.ReadString('\n')
	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	return text
}

// n: name, also
// t: topic
func reader(n, t string) {

	mqr := mq.NewMQReceiver(n)
	mqr.ExchangeName = useTopic
	if err := mqr.AddHandler(n, []string{n}, false, retInfo, cmdRespFunc); err != nil {
		log.Fatalf("remcmd service ..  failed to consume: %+v", err)
	}
	log.Infof("%s Service now consuming on %s", n, n)
}

func retInfo() interface{} {
	return &RespMsg{}
}

func cmdRespFunc(md mq.MessageDelivery) (interface{}, error) {

	msg := md.Message
	r := RespMsg{}

	if mb, k := msg.([]byte); !k {
		return nil, fmt.Errorf("RespMessage: [] byte assertion Error")
	} else {
		err := json.Unmarshal(mb, &r)
		if err != nil {
			log.Infof("msg> %+v", r)
			return nil, fmt.Errorf("RespMessage: Unmarshall Error")
		}
	}

	log.Infof("Response:\n%s\n", r)

	return nil, nil
}
