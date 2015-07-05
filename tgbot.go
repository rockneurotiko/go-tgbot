package tgbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/oleiade/reflections"
)

const (
	baseURL = "https://api.telegram.org/bot%s/%s"
	timeout = 60
)

// DefaultOptionsBot ...
type DefaultOptionsBot struct {
	DisableWebURL              *bool
	Selective                  *bool
	OneTimeKeyboard            *bool
	CleanInitialUsername       bool
	AllowWithoutSlashInMention bool
}

// New just for the lazy guys
func New(token string) *TgBot {
	return NewTgBot(token)
}

// NewTgBot creates a new bot <3
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

// TgBot basic bot struct
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

// AddUsernameCommand ...
func (bot TgBot) AddUsernameCommand(expr string) string {
	strs := strings.Split(expr, " ")
	opts := fmt.Sprintf(`(?:@%s)?`, bot.Username)
	if len(strs) == 1 {
		capt := strs[0]
		lastc := capt[len(capt)-1]
		if lastc == '$' {
			strs[0] = capt[:len(capt)-1] + opts + "$"
		} else {
			strs[0] = strs[0] + opts
		}
	} else {
		strs[0] = strs[0] + opts
	}
	newexpr := strings.Join(strs, " ")
	return newexpr
}

// AddToConditionalFuncs ...
func (bot *TgBot) AddToConditionalFuncs(cf ConditionCallStructure) {
	if !bot.BuildingChain {
		bot.TestConditionalFuncs = append(bot.TestConditionalFuncs, cf)
	} else {
		if len(bot.ChainConditionals) > 0 {
			bot.ChainConditionals[len(bot.ChainConditionals)-1].AddToConditionalFuncs(cf)
		}
	}
}

// StartChain ...
func (bot *TgBot) StartChain() *TgBot {
	bot.ChainConditionals = append(bot.ChainConditionals, NewChainStructure())
	bot.BuildingChain = true
	return bot
}

// CancelChainCommand ...
func (bot *TgBot) CancelChainCommand(path string, f func(TgBot, Message, string) *string) *TgBot {
	if !bot.BuildingChain {
		return bot
	}
	if len(bot.ChainConditionals) > 0 {

		path = convertToCommand(path)
		path = bot.AddUsernameCommand(path)
		r := regexp.MustCompile(path)
		newf := SimpleCommandFuncStruct{f}
		bot.ChainConditionals[len(bot.ChainConditionals)-1].
			SetCancelCond(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	}
	return bot
}

// LoopChain ...
func (bot *TgBot) LoopChain() *TgBot {
	if !bot.BuildingChain {
		return bot
	}
	if len(bot.ChainConditionals) > 0 {
		bot.ChainConditionals[len(bot.ChainConditionals)-1].SetLoop(true)
	}
	return bot
}

// EndChain ...
func (bot *TgBot) EndChain() *TgBot {
	bot.BuildingChain = false
	return bot
}

// DefaultDisableWebpagePreview ...
func (bot *TgBot) DefaultDisableWebpagePreview(b bool) *TgBot {
	bot.DefaultOptions.DisableWebURL = &b
	return bot
}

// DefaultReplyMessage ...
// func (bot *TgBot) DefaultReplyMessage(b bool) *TgBot {
// 	bot.DefaultOptions.ReplyToMessageID = &b
// 	return bot
// }

// DefaultSelective ...
func (bot *TgBot) DefaultSelective(b bool) *TgBot {
	bot.DefaultOptions.Selective = &b
	return bot
}

// DefaultOneTimeKeyboard ...
func (bot *TgBot) DefaultOneTimeKeyboard(b bool) *TgBot {
	bot.DefaultOptions.OneTimeKeyboard = &b
	return bot
}

// DefaultCleanInitialUsername ...
func (bot *TgBot) DefaultCleanInitialUsername(b bool) *TgBot {
	bot.DefaultOptions.CleanInitialUsername = b
	return bot
}

// DefaultAllowWithoutSlashInMention ...
func (bot *TgBot) DefaultAllowWithoutSlashInMention(b bool) *TgBot {
	bot.DefaultOptions.AllowWithoutSlashInMention = b
	return bot
}

// CommandFn Add a command function callback
func (bot *TgBot) CommandFn(path string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.AddUsernameCommand(path)
	r := regexp.MustCompile(path)

	bot.AddToConditionalFuncs(TextConditionalCall{RegexCommand{r, f}})
	return bot
}

// SimpleCommandFn Add a simple command function callback
func (bot *TgBot) SimpleCommandFn(path string, f func(TgBot, Message, string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.AddUsernameCommand(path)
	r := regexp.MustCompile(path)
	newf := SimpleCommandFuncStruct{f}

	bot.AddToConditionalFuncs(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	return bot
}

// MultiCommandFn ...
func (bot *TgBot) MultiCommandFn(paths []string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	rc := []*regexp.Regexp{}
	for _, p := range paths {
		p = convertToCommand(p)
		p = bot.AddUsernameCommand(p)
		r := regexp.MustCompile(p)
		rc = append(rc, r)
	}

	bot.AddToConditionalFuncs(TextConditionalCall{MultiRegexCommand{rc, f}})
	return bot
}

// RegexFn ...
func (bot *TgBot) RegexFn(path string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	r := regexp.MustCompile(path)

	bot.AddToConditionalFuncs(TextConditionalCall{RegexCommand{r, f}})
	return bot
}

// SimpleRegexFn ...
func (bot *TgBot) SimpleRegexFn(path string, f func(TgBot, Message, string) *string) *TgBot {
	r := regexp.MustCompile(path)
	newf := SimpleCommandFuncStruct{f}

	bot.AddToConditionalFuncs(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	return bot
}

// MultiRegexFn ...
func (bot *TgBot) MultiRegexFn(paths []string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	rc := []*regexp.Regexp{}
	for _, p := range paths {
		r := regexp.MustCompile(p)
		rc = append(rc, r)
	}

	bot.AddToConditionalFuncs(TextConditionalCall{MultiRegexCommand{rc, f}})
	return bot
}

// ImageFn ...
func (bot *TgBot) ImageFn(f func(TgBot, Message, []PhotoSize, string)) *TgBot {
	bot.AddToConditionalFuncs(ImageConditionalCall{f})
	return bot
}

// AudioFn ...
func (bot *TgBot) AudioFn(f func(TgBot, Message, Audio, string)) *TgBot {
	bot.AddToConditionalFuncs(AudioConditionalCall{f})
	return bot
}

// DocumentFn ...
func (bot *TgBot) DocumentFn(f func(TgBot, Message, Document, string)) *TgBot {
	bot.AddToConditionalFuncs(DocumentConditionalCall{f})
	return bot
}

// StickerFn ...
func (bot *TgBot) StickerFn(f func(TgBot, Message, Sticker, string)) *TgBot {
	bot.AddToConditionalFuncs(StickerConditionalCall{f})
	return bot
}

// VideoFn ...
func (bot *TgBot) VideoFn(f func(TgBot, Message, Video, string)) *TgBot {
	bot.AddToConditionalFuncs(VideoConditionalCall{f})
	return bot
}

// LocationFn ...
func (bot *TgBot) LocationFn(f func(TgBot, Message, float64, float64)) *TgBot {
	bot.AddToConditionalFuncs(LocationConditionalCall{f})
	return bot
}

// ReplyFn ...
func (bot *TgBot) ReplyFn(f func(TgBot, Message, Message)) *TgBot {
	bot.AddToConditionalFuncs(RepliedConditionalCall{f})
	return bot
}

// ForwardFn ...
func (bot *TgBot) ForwardFn(f func(TgBot, Message, User, int)) *TgBot {
	bot.AddToConditionalFuncs(ForwardConditionalCall{f})
	return bot
}

// GroupFn ..
func (bot *TgBot) GroupFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.AddToConditionalFuncs(GroupConditionalCall{f})
	return bot
}

// NewParticipantFn ...
func (bot *TgBot) NewParticipantFn(f func(TgBot, Message, int, User)) *TgBot {
	bot.AddToConditionalFuncs(NewParticipantConditionalCall{f})
	return bot
}

// LeftParticipantFn ...
func (bot *TgBot) LeftParticipantFn(f func(TgBot, Message, int, User)) *TgBot {
	bot.AddToConditionalFuncs(LeftParticipantConditionalCall{f})
	return bot
}

// NewTitleChatFn ...
func (bot *TgBot) NewTitleChatFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.AddToConditionalFuncs(NewTitleConditionalCall{f})
	return bot
}

// NewPhotoChatFn ...
func (bot *TgBot) NewPhotoChatFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.AddToConditionalFuncs(NewPhotoConditionalCall{f})
	return bot
}

// DeleteChatPhotoFn ...
func (bot *TgBot) DeleteChatPhotoFn(f func(TgBot, Message, int)) *TgBot {
	bot.AddToConditionalFuncs(DeleteChatPhotoConditionalCall{f})
	return bot
}

// GroupChatCreatedFn ...
func (bot *TgBot) GroupChatCreatedFn(f func(TgBot, Message, int)) *TgBot {
	bot.AddToConditionalFuncs(GroupChatCreatedConditionalCall{f})
	return bot
}

// AnyMsgFn ...
func (bot *TgBot) AnyMsgFn(f func(TgBot, Message)) *TgBot {
	bot.AddToConditionalFuncs(CustomCall{AlwaysReturnTrue, f})
	return bot
}

// CustomFn ...
func (bot *TgBot) CustomFn(cond func(TgBot, Message) bool, f func(TgBot, Message)) *TgBot {
	bot.AddToConditionalFuncs(CustomCall{cond, f})
	return bot
}

// ProcessMessages ...
func (bot *TgBot) ProcessMessages(messages []MessageWithUpdateID) {
	for _, msg := range messages {
		if msg.UpdateID > bot.LastUpdateID {
			bot.LastUpdateID = msg.UpdateID
		}
		bot.MainListener <- msg
	}
}

// ProcessMessage ...
func (bot *TgBot) ProcessMessage(msg MessageWithUpdateID) {
	if msg.UpdateID > bot.LastUpdateID {
		bot.LastUpdateID = msg.UpdateID
	}
	bot.MainListener <- msg
}

// CleanMessage ...
func (bot TgBot) CleanMessage(msg Message) Message {
	if bot.DefaultOptions.CleanInitialUsername {
		if msg.Text != nil {
			text := *msg.Text
			username := fmt.Sprintf("@%s", bot.Username)
			if strings.HasPrefix(text, username) {
				text = strings.TrimSpace(strings.Replace(text, username, "", 1)) // Replace one time
				if bot.DefaultOptions.AllowWithoutSlashInMention && !strings.HasSuffix(text, "/") {
					text = "/" + text
				}
				msg.Text = &text
			}
		}
	}

	return msg
}

// ProcessAllMsg ...
func (bot TgBot) ProcessAllMsg(msg Message) {
	msg = bot.CleanMessage(msg)

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

// MessagesHandler ...
func (bot *TgBot) MessagesHandler(Incoming <-chan MessageWithUpdateID) {
	for {
		input := <-Incoming
		go bot.ProcessAllMsg(input.Msg) // go this or not?
	}
}

// MessageHandler ...
func (bot *TgBot) MessageHandler(Incoming <-chan Message) {
	for {
		input := <-Incoming
		go bot.ProcessAllMsg(input) // go this or not?
	}
}

// SimpleStart Start with the default listener and callbacks
func (bot *TgBot) SimpleStart() {
	ch := make(chan MessageWithUpdateID)
	bot.AddMainListener(ch)
	go bot.MessagesHandler(ch)
	bot.Start()
}

// StartWithMessagesChannel ...
func (bot *TgBot) StartWithMessagesChannel(ch chan MessageWithUpdateID) {
	go bot.MessagesHandler(ch)
	bot.Start()
}

// StartWithChannel ...
func (bot *TgBot) StartWithChannel(ch chan Message) {
	go bot.MessageHandler(ch)
	bot.Start()
}

// ServerStart ...
func (bot *TgBot) ServerStart(uri string, pathl string) {
	tokendiv := strings.Split(bot.Token, ":")
	if len(tokendiv) != 2 {
		return
	}
	pathl = path.Join(pathl, fmt.Sprintf("%s%s", tokendiv[0], tokendiv[1]))
	fmt.Println(pathl)
	// fmt.Println(pathl)
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

	ch := bot.GetMessageChannel()
	m := martini.Classic()
	m.Post(pathl, binding.Json(MessageWithUpdateID{}), func(params martini.Params, msg MessageWithUpdateID) {
		// fmt.Println(msg)
		if msg.UpdateID > 0 && msg.Msg.ID > 0 {
			ch <- msg.Msg
		}
	})

	m.Run()
}

// GetMessageChannel ...
func (bot *TgBot) GetMessageChannel() chan Message {
	ch := make(chan Message)
	go bot.MessageHandler(ch)
	return ch
}

// Start ...
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
	i := 0
	for {
		i = i + 1
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

// SetWebhook ...
func (bot TgBot) SetWebhook(args ...string) (ResultSetWebhook, error) {
	pet := SetWebhookQuery{}
	if len(args) >= 1 {
		urlq := args[0]
		pet = SetWebhookQuery{&urlq}
	}
	req := bot.SetWebhookQuery(pet)
	if !req.Ok {
		return req, errors.New(req.Description)
	}
	return req, nil
}

// SetWebhookQuery ...
func (bot TgBot) SetWebhookQuery(q SetWebhookQuery) ResultSetWebhook {
	url := bot.buildPath("setWebhook")
	body, error := postPetition(url, q, nil)

	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultSetWebhook{false, err, nil, &errc}
	}
	var result ResultSetWebhook
	json.Unmarshal([]byte(body), &result)
	return result
}

// GetUserProfilePhotos args will use only the two first parameters, the first one will be the limit of images to get, and the second will be the offset photo id.
func (bot TgBot) GetUserProfilePhotos(uid int, args ...int) UserProfilePhotos {
	pet := ResultWithUserProfilePhotos{}
	getq := GetUserProfilePhotosQuery{uid, nil, nil}
	if len(args) == 1 {
		v1 := args[0]
		getq = GetUserProfilePhotosQuery{uid, nil, &v1}
	} else if len(args) >= 2 {
		v1 := args[0]
		v2 := args[1]
		getq = GetUserProfilePhotosQuery{uid, &v2, &v1}
	}

	pet = bot.GetUserProfilePhotosQuery(getq)

	if !pet.Ok || pet.Result == nil {
		return UserProfilePhotos{}
	}
	return *pet.Result
}

// GetUserProfilePhotosQuery ...
func (bot TgBot) GetUserProfilePhotosQuery(quer GetUserProfilePhotosQuery) ResultWithUserProfilePhotos {
	url := bot.buildPath("getUserProfilePhotos")
	body, error := postPetition(url, quer, nil)

	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultWithUserProfilePhotos{ResultBase{false, &errc, &err}, nil}
	}
	var result ResultWithUserProfilePhotos
	json.Unmarshal([]byte(body), &result)
	return result
}

// GetMe Call getMe path
func (bot TgBot) GetMe() (User, error) {
	body, err := getPetition(bot.buildPath("getMe"), nil)

	if err != nil {
		return User{}, err
	}

	var data ResultGetUser
	dec := json.NewDecoder(strings.NewReader(body))
	dec.Decode(&data)

	if !data.Ok {
		errc := 403
		desc := ""
		if data.ErrorCode != nil {
			errc = *data.ErrorCode
		}
		if data.Description != nil {
			desc = *data.Description
		}

		errormsg := fmt.Sprintf("Some error happened, maybe your token is bad:\nError code: %d\nDescription: %s\nToken: %s", errc, desc, bot.Token)
		return User{}, errors.New(errormsg)
	}
	return data.Result, nil
}

// GetUpdates call getUpdates
func (bot TgBot) GetUpdates() ([]MessageWithUpdateID, error) {
	timeoutreq := fmt.Sprintf("timeout=%d", timeout)
	lastid := fmt.Sprintf("offset=%d", bot.LastUpdateID+1)

	body, err := getPetition(bot.buildPath("getUpdates"), []string{timeoutreq, lastid})

	if err != nil {
		return []MessageWithUpdateID{}, err
	}

	var data ResultGetUpdates
	json.Unmarshal([]byte(body), &data)

	if !data.Ok {
		return []MessageWithUpdateID{}, errors.New("Some error happened in your petition, check your token or remove the webhook.")
	}
	return data.Result, nil
}

// SimpleSendMessage send a simple text message
func (bot TgBot) SimpleSendMessage(msg Message, text string) (res Message, err error) {
	ressm := bot.SendMessage(msg.Chat.ID, text, nil, nil, nil)
	return SplitResultInMessageError(ressm)
}

// SendMessageWithKeyboard ...
func (bot TgBot) SendMessageWithKeyboard(cid int, text string, dwp *bool, rtmid *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, dwp, rtmid, &rkm)
}

// SendMessageWithForceReply ...
func (bot TgBot) SendMessageWithForceReply(cid int, text string, dwp *bool, rtmid *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, dwp, rtmid, &rkm)
}

// SendMessageWithKeyboardHide ...
func (bot TgBot) SendMessageWithKeyboardHide(cid int, text string, dwp *bool, rtmid *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, dwp, rtmid, &rkm)
}

// SendMessage full function wrapper for sendMessage
func (bot TgBot) SendMessage(cid int, text string, dwp *bool, rtmid *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload := QuerySendMessage{cid, text, dwp, rtmid, rm}
	return bot.SendMessageQuery(payload)
}

// SendMessageQuery full sendMessage call
func (bot TgBot) SendMessageQuery(payload QuerySendMessage) ResultWithMessage {
	url := bot.buildPath("sendMessage")
	HookPayload(&payload, bot.DefaultOptions)
	return bot.GenericSendPostData(url, payload)
}

// ForwardMessage full function wrapper for forwardMessage
func (bot TgBot) ForwardMessage(cid int, fid int, mid int) ResultWithMessage {
	payload := ForwardMessageQuery{cid, fid, mid}
	return bot.ForwardMessageQuery(payload)
}

// ForwardMessageQuery  full forwardMessage call
func (bot TgBot) ForwardMessageQuery(payload ForwardMessageQuery) ResultWithMessage {
	url := bot.buildPath("forwardMessage")
	HookPayload(&payload, bot.DefaultOptions)
	return bot.GenericSendPostData(url, payload)
}

// SendPhotoWithKeyboard ...
func (bot TgBot) SendPhotoWithKeyboard(cid int, photo interface{}, caption *string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithForceReply ...
func (bot TgBot) SendPhotoWithForceReply(cid int, photo interface{}, caption *string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithKeyboardHide ...
func (bot TgBot) SendPhotoWithKeyboardHide(cid int, photo interface{}, caption *string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SimpleSendPhoto ...
func (bot TgBot) SimpleSendPhoto(msg Message, photo interface{}) (res Message, err error) {
	cid := msg.Chat.ID
	ressm := bot.SendPhoto(cid, photo, nil, nil, nil)
	return SplitResultInMessageError(ressm)
}

// SendPhoto ...
func (bot TgBot) SendPhoto(cid int, photo interface{}, caption *string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload, err := bot.ImageInterfaceToType(cid, photo, caption, rmi, rm)
	if err != nil {
		errc := 500
		errs := err.Error()
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return bot.SendPhotoQuery(payload)
}

// ImageInterfaceToType ...
func (bot TgBot) ImageInterfaceToType(cid int, photo interface{}, caption *string, rmi *int, rm *ReplyMarkupInt) (payload interface{}, err error) {
	switch pars := photo.(type) {
	case string:
		payload = SendPhotoIDQuery{cid, pars, caption, rmi, rm}
		if LooksLikePath(pars) {
			payload = SendPhotoPathQuery{cid, pars, caption, rmi, rm}
		}
	case image.Image:
		mp := struct {
			ChatID           int             `json:"chat_id"`
			Photo            image.Image     `json:"photo"`
			Caption          *string         `json:"caption,omitempty"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, caption, rmi, rm}
		HookPayload(&mp, bot.DefaultOptions)
		payload = mp
	default:
		err = errors.New("No struct interface detected")
	}
	return
}

// SendPhotoQuery ...
func (bot TgBot) SendPhotoQuery(payload interface{}) ResultWithMessage {
	return bot.SendGenericQuery("sendPhoto", "Photo", "photo", payload)
}

// SendAudioWithKeyboard ...
func (bot TgBot) SendAudioWithKeyboard(cid int, photo string, caption *string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, photo, caption, rmi, &rkm)
}

// SendAudioWithForceReply ...
func (bot TgBot) SendAudioWithForceReply(cid int, photo string, caption *string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, photo, caption, rmi, &rkm)
}

// SendAudioWithKeyboardHide ...
func (bot TgBot) SendAudioWithKeyboardHide(cid int, photo string, caption *string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, photo, caption, rmi, &rkm)
}

// SimpleSendAudio ...
func (bot TgBot) SimpleSendAudio(msg Message, photo string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendAudioIDQuery{cid, photo, nil, nil, nil}
	if LooksLikePath(photo) {
		payload = SendAudioPathQuery{cid, photo, nil, nil, nil}
	}
	ressm := bot.SendAudioQuery(payload)
	return SplitResultInMessageError(ressm)
}

// SendAudio ...
func (bot TgBot) SendAudio(cid int, photo string, caption *string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendAudioIDQuery{cid, photo, caption, rmi, rm}
	if LooksLikePath(photo) {
		payload = SendAudioPathQuery{cid, photo, caption, rmi, rm}
	}
	return bot.SendAudioQuery(payload)
}

// SendAudioQuery ...
func (bot TgBot) SendAudioQuery(payload interface{}) ResultWithMessage {
	return bot.SendGenericQuery("sendAudio", "Audio", "audio", payload)
}

// SendDocumentWithKeyboard ...
func (bot TgBot) SendDocumentWithKeyboard(cid int, photo string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, photo, rmi, &rkm)
}

// SendDocumentWithForceReply ...
func (bot TgBot) SendDocumentWithForceReply(cid int, photo string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, photo, rmi, &rkm)
}

// SendDocumentWithKeyboardHide ...
func (bot TgBot) SendDocumentWithKeyboardHide(cid int, photo string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, photo, rmi, &rkm)
}

// SimpleSendDocument ...
func (bot TgBot) SimpleSendDocument(msg Message, photo string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendDocumentIDQuery{cid, photo, nil, nil}
	if LooksLikePath(photo) {
		payload = SendDocumentPathQuery{cid, photo, nil, nil}
	}
	ressm := bot.SendDocumentQuery(payload)
	return SplitResultInMessageError(ressm)
}

// SendDocument ...
func (bot TgBot) SendDocument(cid int, photo string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendDocumentIDQuery{cid, photo, rmi, rm}
	if LooksLikePath(photo) {
		payload = SendDocumentPathQuery{cid, photo, rmi, rm}
	}
	return bot.SendDocumentQuery(payload)
}

// SendDocumentQuery ...
func (bot TgBot) SendDocumentQuery(payload interface{}) ResultWithMessage {
	return bot.SendGenericQuery("sendDocument", "Document", "document", payload)
}

// SendStickerWithKeyboard ...
func (bot TgBot) SendStickerWithKeyboard(cid int, photo interface{}, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, photo, rmi, &rkm)
}

// SendStickerWithForceReply ...
func (bot TgBot) SendStickerWithForceReply(cid int, photo interface{}, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, photo, rmi, &rkm)
}

// SendStickerWithKeyboardHide ...
func (bot TgBot) SendStickerWithKeyboardHide(cid int, photo interface{}, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, photo, rmi, &rkm)
}

// SimpleSendSticker ...
func (bot TgBot) SimpleSendSticker(msg Message, sticker interface{}) (res Message, err error) {
	cid := msg.Chat.ID
	ressm := bot.SendSticker(cid, sticker, nil, nil)
	return SplitResultInMessageError(ressm)
}

// SendSticker ...
func (bot TgBot) SendSticker(cid int, sticker interface{}, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload, err := bot.StickerInterfaceToType(cid, sticker, rmi, rm)
	if err != nil {
		errc := 500
		errs := err.Error()
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return bot.SendStickerQuery(payload)
	// var payload interface{} = SendStickerIDQuery{cid, photo, rmi, rm}
	// if LooksLikePath(photo) {
	// 	payload = SendStickerPathQuery{cid, photo, rmi, rm}
	// }
	// return bot.SendStickerQuery(payload)
}

// StickerInterfaceToType ...
func (bot TgBot) StickerInterfaceToType(cid int, sticker interface{}, rmi *int, rm *ReplyMarkupInt) (payload interface{}, err error) {
	switch pars := sticker.(type) {
	case string:
		payload = SendStickerIDQuery{cid, pars, rmi, rm}
		if LooksLikePath(pars) {
			payload = SendStickerPathQuery{cid, pars, rmi, rm}
		}
	case image.Image:
		payload = struct {
			ChatID           int             `json:"chat_id"`
			Photo            image.Image     `json:"photo"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, rmi, rm}
	default:
		err = errors.New("No struct interface detected")
	}
	return
}

// SendStickerQuery ...
func (bot TgBot) SendStickerQuery(payload interface{}) ResultWithMessage {
	return bot.SendGenericQuery("sendSticker", "Sticker", "sticker", payload)
}

// SendVideoWithKeyboard ...
func (bot TgBot) SendVideoWithKeyboard(cid int, photo string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, rmi, &rkm)
}

// SendVideoWithForceReply ...
func (bot TgBot) SendVideoWithForceReply(cid int, photo string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, rmi, &rkm)
}

// SendVideoWithKeyboardHide ...
func (bot TgBot) SendVideoWithKeyboardHide(cid int, photo string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, rmi, &rkm)
}

// SimpleSendVideo ...
func (bot TgBot) SimpleSendVideo(msg Message, photo string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendVideoIDQuery{cid, photo, nil, nil}
	if LooksLikePath(photo) {
		payload = SendVideoPathQuery{cid, photo, nil, nil}
	}
	ressm := bot.SendVideoQuery(payload)
	return SplitResultInMessageError(ressm)
}

// SendVideo ...
func (bot TgBot) SendVideo(cid int, photo string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendVideoIDQuery{cid, photo, rmi, rm}
	if LooksLikePath(photo) {
		payload = SendVideoPathQuery{cid, photo, rmi, rm}
	}
	return bot.SendVideoQuery(payload)
}

// SendVideoQuery ...
func (bot TgBot) SendVideoQuery(payload interface{}) ResultWithMessage {
	return bot.SendGenericQuery("sendVideo", "Video", "video", payload)
}

// SimpleSendLocation send a simple text message
func (bot TgBot) SimpleSendLocation(msg Message, latitude float64, longitude float64) (res Message, err error) {
	ressm := bot.SendLocation(msg.Chat.ID, latitude, longitude, nil, nil)
	return SplitResultInMessageError(ressm)
}

// SendLocationWithKeyboard ...
func (bot TgBot) SendLocationWithKeyboard(cid int, latitude float64, longitude float64, rtmid *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocationWithForceReply ...
func (bot TgBot) SendLocationWithForceReply(cid int, latitude float64, longitude float64, rtmid *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocationWithKeyboardHide ...
func (bot TgBot) SendLocationWithKeyboardHide(cid int, latitude float64, longitude float64, rtmid *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocation full function wrapper for sendLocation
func (bot TgBot) SendLocation(cid int, latitude float64, longitude float64, rtmid *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload := SendLocationQuery{cid, latitude, longitude, rtmid, rm}
	return bot.SendLocationQuery(payload)
}

// SendLocationQuery full sendLocation call
func (bot TgBot) SendLocationQuery(payload SendLocationQuery) ResultWithMessage {
	url := bot.buildPath("sendLocation")
	HookPayload(&payload, bot.DefaultOptions)
	return bot.GenericSendPostData(url, payload)
}

// SimpleSendChatAction ...
func (bot TgBot) SimpleSendChatAction(msg Message, ca ChatAction) {
	bot.SendChatAction(msg.Chat.ID, ca)
}

// SendChatAction ...
func (bot TgBot) SendChatAction(cid int, ca ChatAction) {
	bot.SendChatActionQuery(SendChatActionQuery{cid, ca.String()})
}

// SendChatActionQuery ...
func (bot TgBot) SendChatActionQuery(payload SendChatActionQuery) {
	url := bot.buildPath("sendChatAction")
	HookPayload(&payload, bot.DefaultOptions)
	bot.GenericSendPostData(url, payload)
}

// SendConvertingFile ...
func (bot TgBot) SendConvertingFile(url string, ignore string, file string, val interface{}) ResultWithMessage {
	ipath, err := reflections.GetField(val, ignore)
	if err != nil {
		errc := 400
		errs := "Wrong Query!"
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	fpath := fmt.Sprintf("%+v", ipath)
	params := ConvertInterfaceMap(val, []string{ignore})
	return bot.UploadFileWithResult(url, params, file, fpath)
}

// SendGenericQuery ...
func (bot TgBot) SendGenericQuery(path string, ignore string, file string, payload interface{}) ResultWithMessage {
	url := bot.buildPath(path)
	switch val := payload.(type) {
	// ID
	case SendPhotoIDQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.GenericSendPostData(url, val)
	case SendAudioIDQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.GenericSendPostData(url, val)
	case SendDocumentIDQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.GenericSendPostData(url, val)
	case SendStickerIDQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.GenericSendPostData(url, val)
	case SendVideoIDQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.GenericSendPostData(url, val)
		// Path
	case SendPhotoPathQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.SendConvertingFile(url, ignore, file, val)
	case SendAudioPathQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.SendConvertingFile(url, ignore, file, val)
	case SendDocumentPathQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.SendConvertingFile(url, ignore, file, val)
	case SendStickerPathQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.SendConvertingFile(url, ignore, file, val)
	case SendVideoPathQuery:
		HookPayload(&val, bot.DefaultOptions)
		return bot.SendConvertingFile(url, ignore, file, val)
	default:
		ipath, err := reflections.GetField(val, ignore)
		if err != nil {
			break
		}
		params := ConvertInterfaceMap(val, []string{ignore})
		return bot.UploadFileWithResult(url, params, file, ipath)
	}
	errc := 400
	errs := "Wrong Query!"
	return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
}

// GenericSendPostData ...
func (bot TgBot) GenericSendPostData(url string, payload interface{}) ResultWithMessage {
	// hook the payload :P
	body, error := postPetition(url, payload, nil)
	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultWithMessage{ResultBase{false, &errc, &err}, nil}
	}
	var result ResultWithMessage
	json.Unmarshal([]byte(body), &result)
	return result
}

// UploadFileWithResult ...
func (bot TgBot) UploadFileWithResult(url string, params map[string]string, fieldname string, filename interface{}) ResultWithMessage {
	res, err := bot.UploadFile(url, params, fieldname, filename)
	if err != nil {
		errc := 500
		errs := err.Error()
		res = ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return res
}

// UploadFile ...
func (bot TgBot) UploadFile(url string, params map[string]string, fieldname string, filename interface{}) (ResultWithMessage, error) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer

	switch rfile := filename.(type) {
	case string:
		rfile = filepath.Clean(rfile)
		f, err := os.Open(rfile)
		if err != nil {
			return ResultWithMessage{}, err
		}

		fw, err := w.CreateFormFile(fieldname, rfile)
		if err != nil {
			return ResultWithMessage{}, err
		}

		if _, err = io.Copy(fw, f); err != nil {
			return ResultWithMessage{}, err
		}
	case image.Image:
		var imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
		if fw, err = w.CreateFormFile("photo", "image.jpg"); err != nil {
			return ResultWithMessage{}, err
		}
		if err = jpeg.Encode(fw, rfile, &imageQuality); err != nil {
			return ResultWithMessage{}, err
		}
	}

	for key, val := range params {
		if fw, err = w.CreateFormField(key); err != nil {
			return ResultWithMessage{}, err
		}

		if _, err = fw.Write([]byte(val)); err != nil {
			return ResultWithMessage{}, err
		}
	}

	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return ResultWithMessage{}, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ResultWithMessage{}, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ResultWithMessage{}, err
	}

	var apiResp ResultWithMessage
	json.Unmarshal(bytes, &apiResp)

	return apiResp, nil
}

// buildPath build the path
func (bot TgBot) buildPath(action string) string {
	return fmt.Sprintf(bot.BaseRequestURL, action)
}

// AddMainListener ...
func (bot *TgBot) AddMainListener(list chan MessageWithUpdateID) {
	bot.MainListener = list
}

// Send ...
func (bot *TgBot) Send(cid int) *Send {
	return &Send{cid, bot}
}

// Answer ...
func (bot *TgBot) Answer(msg Message) *Send {
	return &Send{msg.Chat.ID, bot}
}
