package logging

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type TelegramHook struct {
	AppName     string
	c           *http.Client
	apiEndpoint string
}

func NewTelegramHook(appName, userName, authToken, targetID string) (*TelegramHook, error) {
	client := &http.Client{}
	return NewTelegramHookWithClient(appName, userName, authToken, targetID, client)
}

// NewTelegramHook creates a new instance of a hook targeting the
// Telegram API with custom http.Client.
func NewTelegramHookWithClient(appName, username, token, chatID string, client *http.Client) (*TelegramHook, error) {
	// https://api.telegram.org/bot1497194583:AAEmn1sr0WDJSzRD_nZKas-DqOS1BYsgpW8/sendMessage?chat_id=1382735620&text=
	apiEndpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s:%s/sendMessage?chat_id=%s&text=",
		username, token, chatID,
	)
	h := TelegramHook{
		AppName: appName,
		c:       client,

		apiEndpoint: apiEndpoint,
	}

	return &h, nil
}

// createMessage crafts an HTML-formatted message to send to the
// Telegram API.
func (hook *TelegramHook) createMessage(entry *logrus.Entry) string {
	var msg string

	switch entry.Level {
	case logrus.PanicLevel:
		msg = "PANIC"
	case logrus.FatalLevel:
		msg = "FATAL"
	case logrus.ErrorLevel:
		msg = "ERROR"
	}

	msg = strings.Join([]string{msg, hook.AppName}, "@")
	msg = strings.Join([]string{msg, entry.Message}, " - ")
	if len(entry.Data) > 0 {
		msg = strings.Join([]string{msg, "<pre>"}, "\n")
		for k, v := range entry.Data {
			msg = strings.Join([]string{msg, html.EscapeString(fmt.Sprintf("\t%s: %+v", k, v))}, "\n")
		}
		msg = strings.Join([]string{msg, "</pre>"}, "\n")
	}
	return msg
}

// Fire emits a log message to the Telegram API.
func (hook *TelegramHook) Fire(entry *logrus.Entry) error {
	msg := hook.createMessage(entry)
	ep := hook.apiEndpoint + msg

	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		// ABSURDLY large keys, for ABSURDLY dumb devices like raspberry.
		TLSHandshakeTimeout: 60 * time.Second,
	}
	c := &http.Client{
		Transport: t,
	}

	resp, err := c.Get(ep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to send message, %v", err)
		return err
	}

	defer resp.Body.Close()

	return nil
}

// Levels returns the log levels that the hook should be enabled for.
func (hook *TelegramHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
