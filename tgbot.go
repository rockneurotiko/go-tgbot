package tgbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/oleiade/reflections"
	"github.com/rockneurotiko/gorequest"
)

const (
	baseURL = "https://api.telegram.org/bot%s/%s"
	timeout = 20
)

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
}

// AddUsernameExpr ...
func (bot TgBot) AddUsernameExpr(expr string) string {
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

func convertToCommand(reg string) string {
	if !strings.HasSuffix(reg, "$") {
		reg = reg + "$"
	}
	if !strings.HasPrefix(reg, "^/") {
		if strings.HasPrefix(reg, "/") {
			reg = "^" + reg
		} else {
			reg = "^/" + reg
		}
	}
	return reg
}

// AddToConditionalFuncs ...
func (bot *TgBot) AddToConditionalFuncs(cf ConditionCallStructure) {
	bot.TestConditionalFuncs = append(bot.TestConditionalFuncs, cf)
}

// CommandFn Add a command function callback
func (bot *TgBot) CommandFn(path string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.AddUsernameExpr(path)
	r := regexp.MustCompile(path)

	bot.AddToConditionalFuncs(TextConditionalCall{RegexCommand{r, f}})
	return bot
}

// SimpleCommandFn Add a simple command function callback
func (bot *TgBot) SimpleCommandFn(path string, f func(TgBot, Message, string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.AddUsernameExpr(path)
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
		p = bot.AddUsernameExpr(p)
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

// FindStringSubmatchMap ...
func FindStringSubmatchMap(r *regexp.Regexp, s string) map[string]string {
	captures := make(map[string]string)
	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}
	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}
	return captures
}

// ProcessAllMsg ...
func (bot TgBot) ProcessAllMsg(msg Message) {
	for _, v := range bot.TestConditionalFuncs {
		if v.canCall(bot, msg) {
			go v.call(bot, msg)
		}
	}
}

// MessageHandler ...
func (bot TgBot) MessageHandler(Incoming <-chan MessageWithUpdateID) {
	for {
		input := <-Incoming
		go bot.ProcessAllMsg(input.Msg) // go this or not?
	}
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

// SimpleStart Start with the default listener and callbacks
func (bot TgBot) SimpleStart() {
	ch := make(chan MessageWithUpdateID)
	bot.AddMainListener(ch)
	go bot.MessageHandler(ch)
	bot.Start()
}

// Start ...
func (bot TgBot) Start() {
	if bot.ID == 0 {
		fmt.Println("No ID, maybe the token is bad.")
		return
	}

	if bot.MainListener == nil {
		fmt.Println("No listener!")
		return
	}

	i := 0
	for {
		i = i + 1
		fmt.Println(i)
		updatesList, err := bot.GetUpdates()
		if err != nil {
			fmt.Println(err)
			continue
		}
		bot.ProcessMessages(updatesList)
	}
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
		errormsg := fmt.Sprintf("Some error happened, maybe your token is bad:\nError code: %d\nDescription: %s\nToken: %s", *data.ErrorCode, *data.Description, bot.Token)
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
		return []MessageWithUpdateID{}, errors.New("Some error happened in your petition, check your token.")
	}
	return data.Result, nil
}

// SimpleSendMessage send a simple text message
func (bot TgBot) SimpleSendMessage(msg Message, text string) (res Message, err error) {
	ressm := bot.SendMessage(msg.Chat.ID, text, nil, nil, nil)

	if ressm.Ok && ressm.Result != nil {
		res = *ressm.Result
		err = nil
	} else {
		res = Message{}
		err = fmt.Errorf("Error in petition.\nError code: %d\nDescription: %s", *ressm.ErrorCode, *ressm.Description)
	}

	return
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
	return bot.GenericSendPostData(url, payload)
}

// ForwardMessage full function wrapper for forwardMessage
func (bot TgBot) ForwardMessage(cid int, fid int, mid int) ResultWithMessage {
	payload := ForwardMessageQuery{cid, fid, mid}
	return bot.ForwardMessageQuery(payload)
}

// LooksLikePath ...
func LooksLikePath(p string) bool {
	p = filepath.Clean(p)
	if len(strings.Split(p, ".")) > 1 {
		// The IDS don't have dots :P
		// But let's check if exist, anyway
		_, err := os.Stat(p)
		return err == nil
	}
	return false
}

// ForwardMessageQuery  full forwardMessage call
func (bot TgBot) ForwardMessageQuery(payload ForwardMessageQuery) ResultWithMessage {
	url := bot.buildPath("forwardMessage")
	return bot.GenericSendPostData(url, payload)
}

// SendPhotoWithKeyboard ...
func (bot TgBot) SendPhotoWithKeyboard(cid int, photo string, caption *string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithForceReply ...
func (bot TgBot) SendPhotoWithForceReply(cid int, photo string, caption *string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithKeyboardHide ...
func (bot TgBot) SendPhotoWithKeyboardHide(cid int, photo string, caption *string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SimpleSendPhoto ...
func (bot TgBot) SimpleSendPhoto(msg Message, photo string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendPhotoIDQuery{cid, photo, nil, nil, nil}
	if LooksLikePath(photo) {
		payload = SendPhotoPathQuery{cid, photo, nil, nil, nil}
	}
	ressm := bot.SendPhotoQuery(payload)

	if ressm.Ok && ressm.Result != nil {
		res = *ressm.Result
		err = nil
	} else {
		res = Message{}
		err = fmt.Errorf("Error in petition.\nError code: %d\nDescription: %s", *ressm.ErrorCode, *ressm.Description)
	}
	return
}

// SendPhoto ...
func (bot TgBot) SendPhoto(cid int, photo string, caption *string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendPhotoIDQuery{cid, photo, caption, rmi, rm}
	if LooksLikePath(photo) {
		payload = SendPhotoPathQuery{cid, photo, caption, rmi, rm}
	}
	return bot.SendPhotoQuery(payload)
}

// SendPhotoQuery ...
func (bot TgBot) SendPhotoQuery(payload interface{}) ResultWithMessage {
	url := bot.buildPath("sendPhoto")
	switch val := payload.(type) {
	case SendPhotoIDQuery:
		return bot.GenericSendPostData(url, val)
	case SendPhotoPathQuery:
		path := val.Photo
		params := ConvertInterfaceMap(val, []string{"Photo"})
		return bot.UploadFileWithResult(url, params, "photo", path)
	}
	errc := 400
	errs := "Wrong Query!"
	return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
}

// GenericSendPostData ...
func (bot TgBot) GenericSendPostData(url string, payload interface{}) ResultWithMessage {
	body, error := postPetition(url, payload)
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
func (bot TgBot) UploadFileWithResult(url string, params map[string]string, fieldname string, filename string) ResultWithMessage {
	res, err := bot.UploadFile(url, params, fieldname, filename)
	if err != nil {
		errc := 500
		errs := err.Error()
		res = ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return res
}

// UploadFile ...
func (bot TgBot) UploadFile(url string, params map[string]string, fieldname string, filename string) (ResultWithMessage, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	filename = filepath.Clean(filename)
	f, err := os.Open(filename)
	if err != nil {
		return ResultWithMessage{}, err
	}

	fw, err := w.CreateFormFile(fieldname, filename)
	if err != nil {
		return ResultWithMessage{}, err
	}

	if _, err = io.Copy(fw, f); err != nil {
		return ResultWithMessage{}, err
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

// IsZeroOfUnderlyingType ...
func IsZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

// IsInList ...
func IsInList(v string, l []string) bool {
	sort.Strings(l)
	i := sort.SearchStrings(l, v)
	return i < len(l) && l[i] == v
}

// ConvertInterfaceMap ...
func ConvertInterfaceMap(p interface{}, except []string) map[string]string {
	nint := map[string]string{}
	var structItems map[string]interface{}

	structItems, _ = reflections.Items(p)
	for v, v2 := range structItems {
		if IsZeroOfUnderlyingType(v2) || IsInList(v, except) {
			continue
		}
		v = strings.ToLower(strings.Join(camelcase.Split(v), "_"))
		switch val := v2.(type) {
		case interface{}:
			sv, _ := json.Marshal(val)
			nint[v] = string(sv)
		default:
			nint[v] = fmt.Sprintf("%+v", v2)
		}
	}
	return nint
}

// buildPath build the path
func (bot TgBot) buildPath(action string) string {
	return fmt.Sprintf(bot.BaseRequestURL, action)
}

// AddMainListener ...
func (bot *TgBot) AddMainListener(list chan MessageWithUpdateID) {
	bot.MainListener = list
}

// postPetition ...
func postPetition(url string, payload interface{}) (string, error) {
	request := gorequest.New().Post(url).
		Send(payload)
	request.TargetType = "form"

	_, body, err := request.End()
	if err != nil {
		return "", errors.New("Some error happened")
	}
	return body, nil
}

// getPetition ...
func getPetition(url string, queries []string) (string, error) {
	req := gorequest.New().Get(url)

	for _, q := range queries {
		req.Query(q)
	}
	_, body, errq := req.End()
	if errq != nil {
		return "", errors.New("There were some error trying to do the petition")
	}
	return body, nil
}
