package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Apps []Application

type Application struct {
	Names []string `json:"names"`
	Exec  string   `json:"exec"`
}

type AppArguments struct {
	Number     string   `json:"number"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	PrevData   string   `json:"prev_data"`
	GlobalData string   `json:"global_data"`
}

type AppReturn struct {
	ToDisplay  string `json:"to_display"`
	NewData    string `json:"new_data"`
	GlobalData string `json:"global_data"`
	Exit       bool   `json:"exit"`
}

func (a *Apps) findExec(name string) string {
	name = strings.ToLower(name)
	for _, app := range *a {
		for _, appName := range app.Names {
			if appName == name {
				return app.Exec
			}
		}
	}
	return ""
}

func (a *Apps) Exec(number PhoneNumber, name string, args []string, prevData string, globalData string) (*AppReturn, error) {
	exeName := a.findExec(name)
	if exeName == "" {
		return nil, fmt.Errorf("failed to find command with name %v", name)
	}

	appArgsDat := AppArguments{
		Number:     string(number),
		Command:    strings.ToLower(name),
		Args:       args,
		PrevData:   prevData,
		GlobalData: globalData,
	}

	appArgsBytes, err := json.Marshal(appArgsDat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments to json: %v", err)
	}

	appArgsStr := string(appArgsBytes)

	fmt.Println(appArgsStr)
	cmd := exec.Command(exeName, appArgsStr)
	var outBuff bytes.Buffer
	cmd.Stdout = &outBuff
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error during execution of %v: %v", exeName, err)
	}
	res := &AppReturn{}
	err = json.Unmarshal(outBuff.Bytes(), res)
	if err != nil {
		return nil, fmt.Errorf("command %v returned an invalid response: %v", exeName, err)
	}
	return res, nil
}
