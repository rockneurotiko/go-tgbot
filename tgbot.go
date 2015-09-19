package tgbot

import (
	"fmt"
	"sync/atomic"

	"net/url"
	"path"
	"strings"

	"github.com/botanio/sdk/go"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gorelic"
)

const (
	baseURL = "https://api.telegram.org/bot%s/%s"
	fileURL = "https://api.telegram.org/file/bot%s/%s"
	timeout = 60
)

// New creates an instance of a new bot with the token supplied, if it's invalid this method fail with a panic.
func New(token string) *TgBot {
	return NewTgBot(token)
}

// NewTgBot creates an instance of a new bot with the token supplied, if it's invalid this method fail with a panic.
func NewTgBot(token string) *TgBot {
	bot, err := NewWithError(token)
	if err != nil {
		panic(err)
	}
	return bot
}

// NewWithError creates an instance and return possible error
func NewWithError(token string) (*TgBot, error) {
	url := fmt.Sprintf(baseURL, token, "%s")
	furl := fmt.Sprintf(fileURL, token, "%s")
	tgbot := &TgBot{
		Token:                token,
		BaseRequestURL:       url,
		BaseFileRequestURL:   furl,
		MainListener:         nil,
		RelicCfg:             nil,
		BotanIO:              nil,
		TestConditionalFuncs: make([]ConditionCallStructure, 0),
		NoMessageFuncs:       make([]NoMessageCall, 0),
		ChainConditionals:    make([]*ChainStructure, 0),
		BuildingChain:        false,
		DefaultOptions: DefaultOptionsBot{
			CleanInitialUsername:       true,
			AllowWithoutSlashInMention: true,
		},
	}
	user, err := tgbot.GetMe()
	if err != nil {
		return nil, err
		// panic(err)
	} else {
		tgbot.FirstName = user.FirstName
		tgbot.ID = user.ID
		tgbot.Username = *user.Username
	}
	return tgbot, nil
}

// TgBot basic bot struct that handle all the interaction functions.
type TgBot struct {
	Token                string
	FirstName            string
	ID                   int
	Username             string
	BaseRequestURL       string
	BaseFileRequestURL   string
	RelicCfg             *RelicConfig
	BotanIO              *botan.Botan
	MainListener         chan MessageWithUpdateID
	LastUpdateID         int64
	TestConditionalFuncs []ConditionCallStructure
	NoMessageFuncs       []NoMessageCall
	ChainConditionals    []*ChainStructure
	BuildingChain        bool
	DefaultOptions       DefaultOptionsBot
}

type RelicConfig struct {
	Token string
	Name  string
}

// ProcessAllMsg default message handler that take care of clean the messages, the chains and the action functions.
func (bot TgBot) ProcessAllMsg(msg Message) {
	msg = bot.cleanMessage(msg)
	// Let's try with and without goroutines here
	for _, c := range bot.ChainConditionals {
		if c.canCall(bot, msg) {
			// c.call(bot, msg)
			go c.call(bot, msg)
			return
		}
		if c.UserInChain(msg) {
			return
		}
	}

	// execlater := make([]ConditionCallStructure, 0)
	executed := false
	for _, v := range bot.TestConditionalFuncs {
		// if nm, ok := v.(NoMessageCall); ok {
		// 	execlater = append(execlater, nm)
		// }
		if v.canCall(bot, msg) {
			executed = true
			v.call(bot, msg)
			// go v.call(bot, msg)
		}
	}

	if !executed {
		for _, f := range bot.NoMessageFuncs {
			f.call(bot, msg)
		}
	}
}

// MessagesHandler is the default listener, just listen for a channel and call the default message processor
func (bot TgBot) MessagesHandler(Incoming <-chan MessageWithUpdateID) {
	for {
		input := <-Incoming
		go bot.ProcessAllMsg(input.Msg) // go this or not?
	}
}

// ProcessMessages will take care about the highest message ID to get updates in the right way. This will call the MainListener channel with a MessageWithUpdateID
func (bot *TgBot) ProcessMessages(messages []MessageWithUpdateID) {
	for _, msg := range messages {
		// if int64(msg.UpdateID) > bot.LastUpdateID {
		// 	// Add lock
		// 	bot.LastUpdateID = int64(msg.UpdateID)
		// }
		atomic.CompareAndSwapInt64(&bot.LastUpdateID, bot.LastUpdateID, int64(msg.UpdateID))
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

	for {
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
	bot.ServerStartHostPort(uri, pathl, "", "")
}

func (bot *TgBot) ServerStartHostPort(uri string, pathl string, host string, port string) {
	if bot.DefaultOptions.RecoverPanic {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("There was some panic: %s\n", r)
			}
		}()
	}
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
			bot.HandleBotan(msg.Msg)
			bot.MainListener <- msg
		}
	})

	if bot.RelicCfg != nil {
		gorelic.InitNewrelicAgent(bot.RelicCfg.Token, bot.RelicCfg.Name, false)
		m.Use(gorelic.Handler)
	}
	if host == "" || port == "" {
		m.Run()
	} else {
		m.RunOnAddr(host + ":" + port)
	}
}

func (bot TgBot) HandleBotan(msg Message) {
	if bot.BotanIO != nil {
		id := msg.Chat.ID
		name := "other"
		if msg.Text != nil {
			name = fmt.Sprintf("text:%s", *msg.Text)
		}
		bot.BotanIO.TrackAsync(id, msg, name, func(a botan.Answer, e []error) {})
	}
}

func (bot *TgBot) SetRelicConfig(tok string, name string) *TgBot {
	bot.RelicCfg = &RelicConfig{tok, name}
	return bot
}

func (bot *TgBot) SetBotanToken(tok string) *TgBot {
	t := botan.New(tok)
	bot.BotanIO = &t
	return bot
}

func (bot TgBot) buildPath(action string) string {
	return fmt.Sprintf(bot.BaseRequestURL, action)
}

func (bot TgBot) buildFilePath(path string) string {
	return fmt.Sprintf(bot.BaseFileRequestURL, path)
}

// Send start a Send petition to the user/chat cid. See Send* structs (SendPhoto, SendVideo, ...)
func (bot *TgBot) Send(cid int) *Send {
	return &Send{cid, bot}
}

// Answer start a Send petition answering the message. See Send* structs (SendPhoto, SendText, ...)
func (bot *TgBot) Answer(msg Message) *Send {
	return &Send{msg.Chat.ID, bot}
}

func (bot *TgBot) File(id string) *SendGetFile {
	return &SendGetFile{bot, id}
}
