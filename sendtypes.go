package tgbot

import (
	"errors"
	"io"
	"io/ioutil"
)

// Send general construct to generate send actions
type Send struct {
	ChatID int
	Bot    *TgBot
}

// Text return a SendText instance to chain actions easy
func (s *Send) Text(text string) *SendText {
	return &SendText{s, text, nil, nil, nil, nil}
}

// Forward return a SendForward instance to chain actions easy
func (s *Send) Forward(to int, msg int) *SendForward {
	return &SendForward{s, to, msg}
}

// Photo return a SendPhoto instance to chain actions easy
func (s *Send) Photo(photo interface{}) *SendPhoto {
	// Check here that the interface works?
	return &SendPhoto{s, photo, nil, nil, nil}
}

// Audio return a SendAudio instance to chain actions easy
func (s *Send) Audio(audio string) *SendAudio {
	return &SendAudio{s, audio, nil, nil, nil, nil, nil}
}

// Voice return a SendVoice instance to chain actions easy
func (s *Send) Voice(voice string) *SendVoice {
	return &SendVoice{s, voice, nil, nil, nil}
}

// Document return a SendDocument instance to chain actions easy
func (s *Send) Document(doc interface{}) *SendDocument {
	return &SendDocument{s, doc, nil, nil}
}

// Sticker return a SendSticker instance to chain actions easy
func (s *Send) Sticker(stick interface{}) *SendSticker {
	return &SendSticker{s, stick, nil, nil}
}

// Video return a SendVideo instance to chain actions easy
func (s *Send) Video(vid string) *SendVideo {
	return &SendVideo{s, vid, nil, nil, nil, nil}
}

// Location return a SendLocation instance to chain actions easy
func (s *Send) Location(latitude float64, long float64) *SendLocation {
	return &SendLocation{s, latitude, long, nil, nil}
}

// Action return a SendAction instance to chain actions easy
func (s *Send) Action(action ChatAction) *SendChatAction {
	return &SendChatAction{s, action}
}

// SendText ...
type SendText struct {
	Send                  *Send
	Text                  string
	ParseModeS            *ParseModeT
	DisableWebPagePreview *bool
	ReplyToMessageID      *int
	ReplyMarkup           *ReplyMarkupInt
}

func (sp *SendText) ParseMode(pm ParseModeT) *SendText {
	sp.ParseModeS = &pm
	return sp
}

// DisablePreview ...
func (sp *SendText) DisablePreview(disab bool) *SendText {
	sp.DisableWebPagePreview = &disab
	return sp
}

// ReplyToMessage ...
func (sp *SendText) ReplyToMessage(rm int) *SendText {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendText) Keyboard(kb ReplyKeyboardMarkup) *SendText {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendText) KeyboardHide(kb ReplyKeyboardHide) *SendText {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendText) ForceReply(fr ForceReply) *SendText {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendText) End() ResultWithMessage {
	return sp.Send.Bot.SendMessage(sp.Send.ChatID, sp.Text, sp.ParseModeS, sp.DisableWebPagePreview, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendForward ...
type SendForward struct {
	Send *Send
	to   int
	msg  int
}

// End ...
func (sf *SendForward) End() ResultWithMessage {
	return sf.Send.Bot.ForwardMessage(sf.Send.ChatID, sf.to, sf.msg)
}

// SendPhoto ...
type SendPhoto struct {
	Send             *Send
	Photo            interface{}
	CaptionField     *string
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// Caption ...
func (sp *SendPhoto) Caption(caption string) *SendPhoto {
	sp.CaptionField = &caption
	return sp
}

// ReplyToMessage ...
func (sp *SendPhoto) ReplyToMessage(rm int) *SendPhoto {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendPhoto) Keyboard(kb ReplyKeyboardMarkup) *SendPhoto {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendPhoto) KeyboardHide(kb ReplyKeyboardHide) *SendPhoto {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendPhoto) ForceReply(fr ForceReply) *SendPhoto {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendPhoto) End() ResultWithMessage {
	return sp.Send.Bot.SendPhoto(sp.Send.ChatID, sp.Photo, sp.CaptionField, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendAudio ...
type SendAudio struct {
	Send             *Send
	Audio            string
	DurationField    *int
	PerformerField   *string
	TitleField       *string
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// Duration ...
func (sp *SendAudio) Duration(d int) *SendAudio {
	sp.DurationField = &d
	return sp
}

// Performer ...
func (sp *SendAudio) Performer(p string) *SendAudio {
	sp.PerformerField = &p
	return sp
}

// Title ...
func (sp *SendAudio) Title(d string) *SendAudio {
	sp.TitleField = &d
	return sp
}

// ReplyToMessage ...
func (sp *SendAudio) ReplyToMessage(rm int) *SendAudio {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendAudio) Keyboard(kb ReplyKeyboardMarkup) *SendAudio {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendAudio) KeyboardHide(kb ReplyKeyboardHide) *SendAudio {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendAudio) ForceReply(fr ForceReply) *SendAudio {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendAudio) End() ResultWithMessage {
	return sp.Send.Bot.SendAudio(sp.Send.ChatID,
		sp.Audio,
		sp.DurationField,
		sp.PerformerField,
		sp.TitleField,
		sp.ReplyToMessageID,
		sp.ReplyMarkup)
}

// SendVoice ...
type SendVoice struct {
	Send             *Send
	Voice            string
	DurationField    *int
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// Duration ...
func (sp *SendVoice) Duration(d int) *SendVoice {
	sp.DurationField = &d
	return sp
}

// ReplyToMessage ...
func (sp *SendVoice) ReplyToMessage(rm int) *SendVoice {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendVoice) Keyboard(kb ReplyKeyboardMarkup) *SendVoice {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendVoice) KeyboardHide(kb ReplyKeyboardHide) *SendVoice {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendVoice) ForceReply(fr ForceReply) *SendVoice {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendVoice) End() ResultWithMessage {
	return sp.Send.Bot.SendVoice(sp.Send.ChatID, sp.Voice, sp.DurationField, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendDocument ...
type SendDocument struct {
	Send             *Send
	Document         interface{}
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// ReplyToMessage ...
func (sp *SendDocument) ReplyToMessage(rm int) *SendDocument {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendDocument) Keyboard(kb ReplyKeyboardMarkup) *SendDocument {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendDocument) KeyboardHide(kb ReplyKeyboardHide) *SendDocument {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendDocument) ForceReply(fr ForceReply) *SendDocument {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendDocument) End() ResultWithMessage {
	return sp.Send.Bot.SendDocument(sp.Send.ChatID, sp.Document, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendSticker ...
type SendSticker struct {
	Send             *Send
	Sticker          interface{}
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// ReplyToMessage ...
func (sp *SendSticker) ReplyToMessage(rm int) *SendSticker {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendSticker) Keyboard(kb ReplyKeyboardMarkup) *SendSticker {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendSticker) KeyboardHide(kb ReplyKeyboardHide) *SendSticker {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendSticker) ForceReply(fr ForceReply) *SendSticker {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendSticker) End() ResultWithMessage {
	return sp.Send.Bot.SendSticker(sp.Send.ChatID, sp.Sticker, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendVideo ...
type SendVideo struct {
	Send             *Send
	Video            string
	CaptionField     *string
	DurationField    *int
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// ReplyToMessage ...
func (sp *SendVideo) ReplyToMessage(rm int) *SendVideo {
	sp.ReplyToMessageID = &rm
	return sp
}

// Caption ...
func (sp *SendVideo) Caption(cap string) *SendVideo {
	sp.CaptionField = &cap
	return sp
}

// Duration ...
func (sp *SendVideo) Duration(dur int) *SendVideo {
	sp.DurationField = &dur
	return sp
}

// Keyboard ...
func (sp *SendVideo) Keyboard(kb ReplyKeyboardMarkup) *SendVideo {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendVideo) KeyboardHide(kb ReplyKeyboardHide) *SendVideo {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendVideo) ForceReply(fr ForceReply) *SendVideo {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendVideo) End() ResultWithMessage {
	return sp.Send.Bot.SendVideo(sp.Send.ChatID, sp.Video, sp.CaptionField, sp.DurationField, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendLocation ...
type SendLocation struct {
	Send             *Send
	Latitude         float64
	Longitude        float64
	ReplyToMessageID *int
	ReplyMarkup      *ReplyMarkupInt
}

// SetLatitude ...
func (sp *SendLocation) SetLatitude(lat float64) *SendLocation {
	sp.Latitude = lat
	return sp
}

// SetLongitude ...
func (sp *SendLocation) SetLongitude(long float64) *SendLocation {
	sp.Longitude = long
	return sp
}

// ReplyToMessage ...
func (sp *SendLocation) ReplyToMessage(rm int) *SendLocation {
	sp.ReplyToMessageID = &rm
	return sp
}

// Keyboard ...
func (sp *SendLocation) Keyboard(kb ReplyKeyboardMarkup) *SendLocation {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// KeyboardHide ...
func (sp *SendLocation) KeyboardHide(kb ReplyKeyboardHide) *SendLocation {
	var rmi ReplyMarkupInt = kb
	sp.ReplyMarkup = &rmi
	return sp
}

// ForceReply ...
func (sp *SendLocation) ForceReply(fr ForceReply) *SendLocation {
	var rmi ReplyMarkupInt = fr
	sp.ReplyMarkup = &rmi
	return sp
}

// End ...
func (sp SendLocation) End() ResultWithMessage {
	return sp.Send.Bot.SendLocation(sp.Send.ChatID, sp.Latitude, sp.Longitude, sp.ReplyToMessageID, sp.ReplyMarkup)
}

// SendChatAction ...
type SendChatAction struct {
	Send   *Send
	Action ChatAction
}

// SetAction ...
func (sca *SendChatAction) SetAction(act ChatAction) *SendChatAction {
	sca.Action = act
	return sca
}

// End ...
func (sca SendChatAction) End() {
	sca.Send.Bot.SendChatAction(sca.Send.ChatID, sca.Action)
}

// SendGetFile ...
type SendGetFile struct {
	Bot *TgBot
	ID  string
}

func (self SendGetFile) ToPath(path string) error {
	body, err := self.ToReader()
	if err != nil {
		return err
	}
	defer body.Close()
	bodyb, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, bodyb, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (self SendGetFile) ToReader() (io.ReadCloser, error) {
	res := self.Bot.GetFile(self.ID)
	if !res.Ok || res.Result == nil {
		msg := ""
		if res.Description != nil {
			msg = *res.Description
		}
		return nil, errors.New(msg)
	}
	fpath := res.Result.Path
	return self.Bot.DownloadFilePathReader(fpath)
}

func (sgf SendGetFile) End() {
	sgf.Bot.GetFile(sgf.ID)
}
