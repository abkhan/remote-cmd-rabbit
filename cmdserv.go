package main

import (
	"flag"
	"fmt"
	"time"

	as "../../../amqpserv"
	log "github.com/sirupsen/logrus"
)

type PiTest struct {
	Who string `json:"iAm"`
	Msg string `json:"myMsg"`
	Cmd string `json:"myCmd"`
}

// Version of the service
var (
	BuildTag            = "0.1.1" // Build version to be provided by build script
	BuildDate           = "na"    // Build date to be provided by build script
	startTime time.Time = time.Now()
	amqs      *as.Service
	mh        as.MessageHandler = remCmdServFunc
	commands  map[string][]string
)

func main() {

	flag.Parse()
	commands = make(map[string][]string)

	// Starting remote command Service
	sname := "remcmd-" // add uname -n value here
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
	return &PiTest{}
}

func remCmdServFunc(md as.MessageDelivery) (interface{}, error) {

	msg := md.Message

	preq, k := msg.(*PiTest)
	if !k {
		return nil, fmt.Errorf("ReqMessage: Parsing Error")
	}

	if preq.Who == "Don" {
		log.Infof("Done Called")
		return soFar, nil
	}

	log.Infof("%s says %s", preq.Who, preq.Msg)
	soFar[preq.Who] = preq.Msg
	return "Got it, " + preq.Who, nil
}
