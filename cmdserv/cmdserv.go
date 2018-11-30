package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	as "github.com/abkhan/amqpserv"
	log "github.com/sirupsen/logrus"
)

type CmdReq struct {
	Who string            `json:"iAm"`
	Msg string            `json:"msg"`
	Cmd string            `json:"cmd"`
	Env map[string]string `json:"env"`
}

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
	amqs      *as.Service
	mh        as.MessageHandler = remCmdServFunc
	commands  map[string]CmdReq
)

func main() {

	flag.Parse()
	commands = make(map[string]CmdReq)

	// Starting remote command Service
	sname := "remcmd-" + "mac" // add uname -n value here
	amqs = as.AmqpService(sname)
	defer amqs.Cleanup()

	log.Infof("Run RemCmd Server")
	log.Infof(">> Build Version %s, on date %s", BuildTag, BuildDate)

	//this sleep is required to avoid search timeout issue
	time.Sleep(3 * time.Second)

	if err := amqs.ConsumeFunc(sname, []string{sname}, retInfo, mh); err != nil {
		log.Fatalf("remcmd service ..  failed to consume: %+v", err)
	}
	log.Infof("%s Service now consuming on %s", sname, sname)

	// setup end processing
	amqs.Wait()

}

//
func retInfo() interface{} {
	return &CmdReq{}
}

func remCmdServFunc(md as.MessageDelivery) (interface{}, error) {

	msg := md.Message

	preq, k := msg.(*CmdReq)
	if !k {
		return nil, fmt.Errorf("ReqMessage: Parsing Error")
	}

	if preq.Who == "Don" {
		log.Infof("Done Called")
		return commands, nil
	}
	commands[preq.Who] = *preq
	log.Infof("Running Cmd: %+v", preq)
	return preq.run(), nil
}
