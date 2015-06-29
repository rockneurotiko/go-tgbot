package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/rockneurotiko/go-tgbot"
)

const (
	token = ""
)

var availableCommands = map[string]string{
	"/start":          "Start the bot!",
	"/help":           "Get help!!",
	"/helpbotfather":  "Get the help formatted to botfather",
	"/help <command>": "Get the help of one command",
	"/keyboard":       "Send you a keyboard",
	"/hidekeyboard":   "Hide the keyboard",
	"/hardecho":       "Echo with force reply",
	"/forwardme":      "Forward that message to you",
	"/sleep":          "Sleep for 5 seconds, without blocking, awesome goroutines",
	"/showmecommands": "Returns you a keyboard with the simplest commands",
}

func buildHelpMessage(complete bool) string {
	var buffer bytes.Buffer
	for cmd, htext := range availableCommands {
		str := ""
		if complete {
			str = fmt.Sprintf("%s - %s\n", cmd, htext)
		} else if len(strings.Split(cmd, " ")) == 1 {
			str = fmt.Sprintf("%s - %s\n", cmd[1:], htext)
		}
		buffer.WriteString(str)
	}
	return buffer.String()
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
	res := ""
	if vals[0] == "/help" {
		res = buildHelpMessage(true)
	} else if vals[0] == "/helpbotfather" {
		res = buildHelpMessage(false)
	}
	return &res
}

func testGoroutineHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.SimpleSendMessage(msg, "Starting")
	time.Sleep(5000 * time.Millisecond)
	r := "Ending"
	return &r
}

func showMeHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{}}
	for k := range availableCommands {
		if len(strings.Split(k, " ")) == 1 {
			if len(keylayout[len(keylayout)-1]) == 2 {
				keylayout = append(keylayout, []string{k})
			} else {
				keylayout[len(keylayout)-1] = append(keylayout[len(keylayout)-1], k)
			}
		}
	}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       false}
	bot.SendMessageWithKeyboard(msg.Chat.ID, "There you have the commands!", nil, nil, rkm)
	return nil
}

func allMsgHand(bot tgbot.TgBot, msg tgbot.Message) {
	// uncomment this to see it :)
	fmt.Printf("Received message: %+v\n", msg.ID)
	// bot.SimpleSendMessage(msg, "Received message!")
}

func conditionFunc(bot tgbot.TgBot, msg tgbot.Message) bool {
	return msg.Photo != nil
}

func conditionCallFunc(bot tgbot.TgBot, msg tgbot.Message) {
	bot.SimpleSendMessage(msg, "Nice image :)")
}

func main() {
	bot := tgbot.NewTgBot(token).
		SimpleCommandFn(`sleep`, testGoroutineHand).
		SimpleCommandFn(`keyboard`, cmdKeyboard).
		SimpleCommandFn(`hidekeyboard`, hideKeyboard).
		SimpleCommandFn(`forwardme`, forwardHand).
		SimpleCommandFn(`showmecommands`, showMeHand).
		CommandFn(`hardecho (.+)`, hardEcho).
		MultiCommandFn([]string{`help (\w+)`, `help`, `helpbotfather`}, multiregexHelpHand).
		SimpleRegexFn(`^Hello!$`, helloHand).
		RegexFn(`^Tell me (.+)$`, tellmeHand).
		AnyMsgFn(allMsgHand).
		CustomFn(conditionFunc, conditionCallFunc)

	// bot := tgbot.NewTgBot(token)
	// bot.SimpleCommandFn(`^/sleep$`, testGoroutineHand)
	// bot.SimpleCommandFn(`^/keyboard$`, cmdKeyboard)
	// bot.SimpleCommandFn(`^/hidekeyboard$`, hideKeyboard)
	// bot.SimpleCommandFn(`^/forwardme$`, forwardHand)
	// bot.SimpleCommandFn(`^/showmecommands`, showMeHand)
	// bot.CommandFn(`^/hardecho (.+)`, hardEcho)
	// bot.MultiCommandFn([]string{`^/help (\w+)$`, `^/help$`, `^/helpbotfather$`}, multiregexHelpHand)
	// bot.SimpleRegexFn(`^Hello!$`, helloHand)
	// bot.RegexFn(`^Tell me (.+)$`, tellmeHand)
	// bot.AnyMsgFn(allMsgHand)
	// bot.CustomFn(conditionFunc, conditionCallFunc)
	bot.SimpleStart()
}
