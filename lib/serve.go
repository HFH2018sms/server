package lib

import (
	"log"
	"time"

	"encoding/json"
	"io/ioutil"

	"strings"
)

type Configuration struct {
	RedisSocket string      `json:"redis_socket"`
	TwiloConfig TwilioCreds `json:"twilio_cfg"`
	AppsConfig  Apps        `json:"apps_config"`
}

func Serve(configFile string) {
	file, err := ioutil.ReadFile(configFile)
	fatalError("failed to read configuration file", err)

	cfg := &Configuration{}
	err = json.Unmarshal(file, cfg)
	fatalError("configuration file invalid", err)

	twilio, err := SetupTwilio(cfg.TwiloConfig)
	fatalError("failed to connect to twilio", err)

	r := RedisConnect(cfg.RedisSocket)
	lmExist, err := r.LastMessageExists()
	fatalError("failed to connect to redis", err)

	if !lmExist {
		err = r.SetLastMessage("none")
		fatalError("Failed to connect to redis", err)
	}
	fatalError("Failed to load twilio configuration", err)

	for true {
		lastMessage, err := r.GetLastMessage()
		fatalError("Failed to get last message", err)
		newMessages, err := twilio.GetNewMessages(lastMessage)
		fatalError("Failed  to get messages from twilio", err)
		if len(newMessages) > 0 {
			err = r.SetLastMessage(newMessages[0].Sid)
			fatalError("Failed to set last message", err)
		}
		for i := len(newMessages) - 1; i >= 0; i-- {
			toDisplay := handleRequest(r, &newMessages[i], &cfg.AppsConfig)
			if toDisplay != "" {
				twilio.SendMessage(newMessages[i].From, toDisplay)
			}
		}
		time.Sleep(time.Second)
	}
}

func handleRequest(r *Redis, t *TwilioMessage, appCfg *Apps) string {
	currentProgram, err := r.GetCurrentProgram(t.From)
	fatalError("Failed to get current program", err)

	fields := strings.Fields(t.Body)
	if len(fields) == 0 {
		return ""
	}

	var ret *AppReturn
	var prog string
	var args []string
	if currentProgram == "none" {
		prog = fields[0]
		args = fields[1:]
	} else {
		prog = currentProgram
		args = fields
	}

	data, err := r.GetPrevData(t.From, prog)
	fatalError("failed to get previous program data", err)

	globalData, err := r.GetGlobalData(prog)
	fatalError("failed to get global program data", err)

	ret, err = appCfg.Exec(t.From, prog, args, data, globalData)
	if err != nil {
		return err.Error()
	}

	err = r.SetCurrentProgram(t.From, prog)
	fatalError("failed to save new program", err)

	err = r.SetData(t.From, prog, ret.NewData)
	fatalError("failed to get previous program data", err)

	err = r.SetGlobalData(prog, ret.GlobalData)
	fatalError("failed to get global program data", err)

	if ret.Exit {
		err = r.SetCurrentProgram(t.From, "none")
		fatalError("failed to save new program", err)
	}

	return ret.ToDisplay
}

func fatalError(str string, err error) {
	if err != nil {
		log.Fatalf("%v: %+v", str, err)
	}
}
