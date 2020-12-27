package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	as "github.com/abkhan/mqwrap"
	log "github.com/sirupsen/logrus"
)

type CmdReq struct {
	Who string            `json:"iAm"`
	Msg string            `json:"msg"`
	Cmd string            `json:"cmd"`
	Env map[string]string `json:"env"`
}

const useExchange = "scope.polling"

func (p *CmdReq) run() string {
	cmd := exec.Command(p.Cmd)
	cmdp := strings.Split(p.Cmd, " ")
	if len(cmdp) > 1 {
		cmd = exec.Command(cmdp[0], cmdp[1:]...)
	}
	for en, ev := range p.Env {
		cmd.Env = append(os.Environ(), en+"="+ev)
	}
	//cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "cmd exec error: " + err.Error()
	}
	return out.String()
}

// Version of the service
var (
	BuildTag            = "0.1.1" // Build version to be provided by build script
	BuildDate           = "na"    // Build date to be provided by build script
	startTime time.Time = time.Now()
	amqs      *as.MQWrap
	mh        as.MessageHandler = remCmdServFunc
	commands  map[string]CmdReq
)

func main() {

	flag.Parse()
	commands = make(map[string]CmdReq)
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("error getting hostname: %+v", err)
	}

	log.Infof("Run RemCmd Server on %s", host)
	log.Infof(">> Build Version %s, on date %s", BuildTag, BuildDate)

	// Starting remote command Service
	sname := "remcmd-" + host
	amqs = as.NewMQReceiver(sname)
	amqs.ExchangeName = useExchange

	//this sleep is required to avoid search timeout issue
	time.Sleep(3 * time.Second)
	log.Info("Creating handler")

	if err := amqs.AddHandler(sname, []string{sname}, false, retInfo, mh); err != nil {
		log.Fatalf("remcmd service ..  failed to consume: %+v", err)
	}
	log.Infof("%s Service now consuming on %s", sname, sname)

	// stay alive
	select {}
}

//
func retInfo() interface{} {
	return &CmdReq{}
}

func remCmdServFunc(md as.MessageDelivery) (interface{}, error) {

	msg := md.Message
	c := CmdReq{}

	if mb, k := msg.([]byte); !k {
		return nil, fmt.Errorf("ReqMessage: [] byte assertion Error")
	} else {
		err := json.Unmarshal(mb, &c)
		if err != nil {
			log.Infof("msg> %+v", c)
			return nil, fmt.Errorf("ReqMessage: Unmarshall Error")
		}
	}
	log.Infof("Message>ReplyTo>%s", md.Delivery.ReplyTo)

	if c.Who == "Don" {
		log.Infof("Done Called")
		return commands, nil
	}
	commands[c.Who] = c
	log.Infof("Running Cmd: %+v", c)
	return c.run(), nil
}
