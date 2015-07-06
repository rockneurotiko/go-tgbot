package tgbot

import (
	"fmt"

	"net/url"
	"path"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/binding"
)

const (
	baseURL = "https://api.telegram.org/bot%s/%s"
	timeout = 60
)

// New creates an instance of a new bot with the token supplied, if it's invalid this method fail with a panic.
func New(token string) *TgBot {
	return NewTgBot(token)
}

// NewTgBot creates an instance of a new bot with the token supplied, if it's invalid this method fail with a panic.
func NewTgBot(token string) *TgBot {
	url := fmt.Sprintf(baseURL, token, "%s")
	tgbot := &TgBot{
		Token:                token,
		BaseRequestURL:       url,
		MainListener:         nil,
		TestConditionalFuncs: []ConditionCallStructure{},
		ChainConditionals:    []*ChainStructure{},
		BuildingChain:        false,
		DefaultOptions: DefaultOptionsBot{
			CleanInitialUsername:       true,
			AllowWithoutSlashInMention: true,
		},
	}
	user, err := tgbot.GetMe()
	if err != nil {
		panic(err)
	} else {
		tgbot.FirstName = user.FirstName
		tgbot.ID = user.ID
		tgbot.Username = *user.Username
	}
	return tgbot
}

// TgBot basic bot struct that handle all the interaction functions.
type TgBot struct {
	Token                string
	FirstName            string
	ID                   int
	Username             string
	BaseRequestURL       string
	MainListener         chan MessageWithUpdateID
	LastUpdateID         int
	TestConditionalFuncs []ConditionCallStructure
	ChainConditionals    []*ChainStructure
	BuildingChain        bool
	DefaultOptions       DefaultOptionsBot
}

// ProcessAllMsg default message handler that take care of clean the messages, the chains and the action functions.
func (bot TgBot) ProcessAllMsg(msg Message) {
	msg = bot.cleanMessage(msg)

	for _, c := range bot.ChainConditionals {
		if c.canCall(bot, msg) {
			go c.call(bot, msg)
			return
		}
		if c.UserInChain(msg) {
			return
		}
	}

	for _, v := range bot.TestConditionalFuncs {
		if v.canCall(bot, msg) {
			go v.call(bot, msg)
		}
	}
}

// MessagesHandler is the default listener, just listen for a channel and call the default message processor
func (bot *TgBot) MessagesHandler(Incoming <-chan MessageWithUpdateID) {
	for {
		input := <-Incoming
		go bot.ProcessAllMsg(input.Msg) // go this or not?
	}
}

// ProcessMessages will take care about the highest message ID to get updates in the right way. This will call the MainListener channel with a MessageWithUpdateID
func (bot *TgBot) ProcessMessages(messages []MessageWithUpdateID) {
	for _, msg := range messages {
		if msg.UpdateID > bot.LastUpdateID {
			bot.LastUpdateID = msg.UpdateID
		}
		if bot.MainListener != nil {
			bot.MainListener <- msg
		}
	}
}

// AddMainListener add the channel as the main listener, this will be called with the messages received.
func (bot *TgBot) AddMainListener(list chan MessageWithUpdateID) {
	bot.MainListener = list
}

// GetMessageChannel create a channel and start the default messages handler, you can use this to build your own server listener (just send the MessageWithUpdateID to that channel)
func (bot *TgBot) GetMessageChannel() chan MessageWithUpdateID {
	ch := make(chan MessageWithUpdateID)
	go bot.MessagesHandler(ch)
	return ch
}

// StartMainListener will run a start a new channel and start the default message handler, assigning it to the main listener.
func (bot *TgBot) StartMainListener() {
	ch := bot.GetMessageChannel()
	bot.AddMainListener(ch)
}

// SimpleStart will start to get updates with the default listener and callbacks (with long-polling getUpdates way)
func (bot *TgBot) SimpleStart() {
	bot.StartMainListener()
	bot.Start()
}

// StartWithMessagesChannel will start to get updates with your own channel that handle the messages (with long-polling getUpdates way)
func (bot *TgBot) StartWithMessagesChannel(ch chan MessageWithUpdateID) {
	go bot.MessagesHandler(ch)
	bot.Start()
}

// Start will start the main process (that use the MainListener channel), it uses getUpdates with longs-polling way and handle the ID
func (bot *TgBot) Start() {
	if bot.ID == 0 {
		fmt.Println("No ID, maybe the token is bad.")
		return
	}

	if bot.MainListener == nil {
		fmt.Println("No listener!")
		return
	}

	removedhook := false

	// i := 0
	for {
		// i = i + 1
		// fmt.Println(i)
		updatesList, err := bot.GetUpdates()
		if err != nil {
			fmt.Println(err)
			if !removedhook {
				fmt.Println("Removing webhook...")
				bot.SetWebhook("")
				removedhook = true
			}
			continue
		}
		bot.ProcessMessages(updatesList)
	}
}

// ServerStart starts a server that listen for updates, if uri parameter is not empty string, it will try to set the proper webhook
// The default server uses Martini classic, and listen in POST /<pathl>/token (The token without the :)
// To listen it runs the Martini.Run() method, that get $HOST and $PORT from the environment, or uses locashost/3000 if not setted.
func (bot *TgBot) ServerStart(uri string, pathl string) {
	tokendiv := strings.Split(bot.Token, ":")
	if len(tokendiv) != 2 {
		return
	}
	pathl = path.Join(pathl, fmt.Sprintf("%s%s", tokendiv[0], tokendiv[1]))

	if uri != "" {
		puri, err := url.Parse(uri)
		if err != nil {
			fmt.Printf("Bad URL %s", uri)
			return
		}
		nuri, _ := puri.Parse(pathl)
		res, error := bot.SetWebhook(nuri.String())
		if error != nil {
			ec := res.ErrorCode
			fmt.Printf("Error setting the webhook: \nError code: %d\nDescription: %s\n", &ec, res.Description)
			return
		}
	}

	if bot.MainListener == nil {
		bot.StartMainListener()
	}

	m := martini.Classic()
	m.Post(pathl, binding.Json(MessageWithUpdateID{}), func(params martini.Params, msg MessageWithUpdateID) {
		if msg.UpdateID > 0 && msg.Msg.ID > 0 {
			bot.MainListener <- msg
		}
	})

	m.Run()
}

func (bot TgBot) buildPath(action string) string {
	return fmt.Sprintf(bot.BaseRequestURL, action)
}

// Send start a Send petition to the user/chat cid. See Send* structs (SendPhoto, SendVideo, ...)
func (bot *TgBot) Send(cid int) *Send {
	return &Send{cid, bot}
}

// Answer start a Send petition answering the message. See Send* structs (SendPhoto, SendText, ...)
func (bot *TgBot) Answer(msg Message) *Send {
	return &Send{msg.Chat.ID, bot}
}
