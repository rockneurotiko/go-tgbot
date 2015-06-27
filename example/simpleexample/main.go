package main

import (
	"bytes"
	"fmt"

	"github.com/rockneurotiko/go-tgbot"
)

const (
	token = ""
)

var availableCommands = map[string]string{
	"/help":         "Get help!!",
	"/start":        "Start the bot with you!",
	"/keyboard":     "Send you a keyboard",
	"/hidekeyboard": "Hide the keyboard",
	"/hardecho":     "Echo with force reply",
}

func buildHelpMessage() string {
	var buffer bytes.Buffer
	for cmd, htext := range availableCommands {
		str := fmt.Sprintf("%s - %s\n", cmd, htext)
		buffer.WriteString(str)
	}
	return buffer.String()
}

func helpHandler(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	res := buildHelpMessage()
	return &res
}

func hideKeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	rkm := tgbot.ReplyKeyboardHide{HideKeyboard: true, Selective: false}
	bot.SendMessageWithKeyboardHide(msg.Chat.ID, "Hiden it!", nil, nil, rkm)
	return nil
}

func cmdKeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{"I", "<3"}, {"You"}}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: false,
		Selective:       false}
	bot.SendMessageWithKeyboard(msg.Chat.ID, "Enjoy the keyboard", nil, nil, rkm)
	return nil
}

func hardEcho(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	msgtext := ""
	if len(vals) > 1 {
		msgtext = vals[1]
	}
	rkm := tgbot.ForceReply{Force: true, Selective: false}
	bot.SendMessageWithForceReply(msg.Chat.ID, msgtext, nil, nil, rkm)

	return nil
}

func forwardHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.ForwardMessage(msg.Chat.ID, msg.Chat.ID, msg.ID)
	return nil
}

func main() {
	bot := tgbot.NewTgBot(token)
	bot.SimpleCommandFn(`^/help$`, helpHandler)
	bot.SimpleCommandFn(`^/keyboard$`, cmdKeyboard)
	bot.SimpleCommandFn(`^/hidekeyboard$`, hideKeyboard)
	bot.SimpleCommandFn(`^/forwardme$`, forwardHand)
	bot.CommandFn(`^/hardecho (.+)`, hardEcho)
	bot.SimpleStart()
}
