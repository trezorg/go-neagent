package main

import (
	"net/http"
)

const (
	defaultConfigFile = "neagent.conf"
	defaultTimeout    = 3600
	telegramLink      = "https://api.telegram.org/bot%s/sendMessage?chat_id=%s"
)

var (
	agentHeaders = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.116 Safari/537.36"},
		"Accept":     []string{"text/html"},
	}
)
