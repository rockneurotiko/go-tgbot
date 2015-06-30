package main

import (
	"fmt"

	"github.com/rockneurotiko/go-tgbot"
)

func echoHandler(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	newmsg := fmt.Sprintf("[Echoed]: %s", vals[1])
	return &newmsg
}

func main() {
	bot := tgbot.NewTgBot("token").
		CommandFn(`echo (.+)`, echoHandler)
	bot.SimpleStart()
}
