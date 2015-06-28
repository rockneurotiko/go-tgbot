package main

import (
	"bytes"
	"fmt"
	"time"

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

func helloHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	msgr := fmt.Sprintf("Hi %s! <3", msg.From.FirstName)
	return &msgr
}

func tellmeHand(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	msgtext := ""
	if len(vals) > 1 {
		msgtext = vals[1]
	}
	return &msgtext
}

func multiregexHelpHand(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	if len(vals) > 1 {
		for k, v := range availableCommands {
			if k[1:] == vals[1] {
				res := v
				return &res
			}
		}
	}
	res := buildHelpMessage()
	return &res
}

func testGoroutineHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.SimpleSendMessage(msg, "Starting")
	time.Sleep(5000 * time.Millisecond)
	r := "Ending"
	return &r
}

func main() {
	bot := tgbot.NewTgBot(token).
		MultiRegexFn([]string{`^/help (\w+)$`, `^/help$`}, multiregexHelpHand).
		SimpleCommandFn(`^/sleep$`, testGoroutineHand).
		SimpleCommandFn(`^/keyboard$`, cmdKeyboard).
		SimpleCommandFn(`^/hidekeyboard$`, hideKeyboard).
		SimpleCommandFn(`^/forwardme$`, forwardHand).
		CommandFn(`^/hardecho (.+)`, hardEcho).
		SimpleRegexFn(`^Hello!$`, helloHand).
		RegexFn(`^Tell me (.+)$`, tellmeHand)

	bot.SimpleStart()
}
