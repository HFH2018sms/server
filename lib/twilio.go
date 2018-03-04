package lib

import (
	"encoding/json"
	"io/ioutil"

	"fmt"

	"net/http"

	"github.com/sfreiberg/gotwilio"
)

type TwilioCreds struct {
	Sid    string      `json:"sid"`
	Secret string      `json:"secret"`
	Number PhoneNumber `json:"number"`
}

type Twilio struct {
	gotwilio.Twilio
	Number           PhoneNumber
	NewestMessageSid string
}

type PhoneNumber string

type TwilioMessage struct {
	Sid         string      `json:"sid"`
	DateCreated string      `json:"date_created"`
	DateUpdated string      `json:"date_updated"`
	DateSent    string      `json:"date_sent"`
	AccountSid  string      `json:"account_sid"`
	To          PhoneNumber `json:"to"`
	From        PhoneNumber `json:"from"`
	Body        string      `json:"body"`
	Status      string      `json:"status"`
	NumSegments string      `json:"num_segments"`
}

type TwilioMessageList struct {
	Messages    []TwilioMessage `json:"messages"`
	Start       int             `json:"start"`
	End         int             `json:"end"`
	Page        int             `json:"page"`
	PageSize    int             `json:"page_size"`
	Uri         string          `json:"uri"`
	NextPageUri string          `json:"next_page_uri"`
	PrevPageUri string          `json:"previous_page_uri"`
}

func SetupTwilio(creds TwilioCreds) (*Twilio, error) {
	ans := Twilio{}
	ans.Number = creds.Number
	ans.Twilio = *gotwilio.NewTwilioClient(creds.Sid, creds.Secret)

	return &ans, nil
}

func (t *Twilio) GetNewMessages(prevSid string) ([]TwilioMessage, error) {
	currentSid := ""
	uri := "https://api.twilio.com/2010-04-01/Accounts/" + t.AccountSid + "/Messages.json"
	ans := make([]TwilioMessage, 0)
	for currentSid != prevSid && uri != "https://api.twilio.com" {
		client := &http.Client{}
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %v", err)
		}
		req.SetBasicAuth(t.AccountSid, t.AuthToken)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to request: %v", err)
		}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		messageList := &TwilioMessageList{}
		err = json.Unmarshal(bodyBytes, messageList)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall json data: %v", err)
		}
		for _, message := range messageList.Messages {
			currentSid = message.Sid
			if currentSid == prevSid {
				break
			}
			if message.From != t.Number {
				ans = append(ans, message)
			}
		}
		uri = "https://api.twilio.com" + messageList.NextPageUri
	}
	return ans, nil
}

func (t *Twilio) SendMessage(num PhoneNumber, mesg string) (*gotwilio.SmsResponse, error) {
	resp, exception, err := t.SendSMS(string(t.Number), string(num), mesg, "", "")
	if err != nil {
		return nil, fmt.Errorf("sendsms error: %v", err)
	}
	if exception != nil && exception.Code != 200 {
		return nil, fmt.Errorf("%v", exception)
	}
	return resp, nil
}
