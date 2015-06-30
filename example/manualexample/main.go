package main

import (
	"bytes"
	"fmt"
	"os"

	godotenv "github.com/joho/godotenv"
	"github.com/rockneurotiko/go-tgbot"
)

var availableCommands = map[string]string{
	"/start": "Start the bot with you!",
	"/help":  "Get help!!",
}

func buildHelpMessage() string {
	var buffer bytes.Buffer
	for cmd, htext := range availableCommands {
		str := fmt.Sprintf("%s - %s\n", cmd, htext)
		buffer.WriteString(str)
	}
	return buffer.String()
}

func handleMessageText(text string, message tgbot.Message) string {
	tosend := ""
	if text == "/help" {
		tosend = buildHelpMessage()
	}
	return tosend
}

// MessageHandler will be the custom handler
func MessageHandler(Incoming <-chan tgbot.MessageWithUpdateID, bot *tgbot.TgBot) {
	for {
		input := <-Incoming

		if input.Msg.Text != nil {
			text := handleMessageText(*input.Msg.Text, input.Msg)
			if text != "" {
				nmsg, err := bot.SimpleSendMessage(input.Msg, text)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(nmsg.String())
			}
		}
	}
}

func main() {
	godotenv.Load("secrets.env")
	// Add a file secrets.env, with the key like:
	// TELEGRAM_KEY=yourtoken
	token := os.Getenv("TELEGRAM_KEY")
	bot := tgbot.NewTgBot(token)
	ch := make(chan tgbot.MessageWithUpdateID)
	bot.AddMainListener(ch)
	go MessageHandler(ch, bot)
	bot.Start()
}
