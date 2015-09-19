package tgbot

import (
	"fmt"
	"regexp"
)

// ConditionCallStructure ...
type ConditionCallStructure interface {
	canCall(TgBot, Message) bool
	call(TgBot, Message)
}

// CustomCall ...
type CustomCall struct {
	condition func(TgBot, Message) bool
	f         func(TgBot, Message)
}

// canCall ...
func (cc CustomCall) canCall(bot TgBot, msg Message) bool {
	return cc.condition(bot, msg)
}

// call ...
func (cc CustomCall) call(bot TgBot, msg Message) {
	cc.f(bot, msg)
}

// Custom functions for CustomCall :)

// AlwaysReturnTrue ...
func AlwaysReturnTrue(bot TgBot, msg Message) bool {
	return true
}

// AlwaysReturnFalse ...
func AlwaysReturnFalse(bot TgBot, msg Message) bool {
	return false
}

// NewChainStructure ...
func NewChainStructure() *ChainStructure {
	return &ChainStructure{[]ConditionCallStructure{}, map[int]int{}, false, nil}
}

// ChainStructure ...
type ChainStructure struct {
	chainf     []ConditionCallStructure
	alreadyin  map[int]int // who and what index
	loop       bool
	cancelcond *ConditionCallStructure
}

// AddToConditionalFuncs ...
func (cc *ChainStructure) AddToConditionalFuncs(cf ConditionCallStructure) {
	cc.chainf = append(cc.chainf, cf)
}

// SetLoop ...
func (cc *ChainStructure) SetLoop(b bool) {
	cc.loop = b
}

// SetCancelCond ...
func (cc *ChainStructure) SetCancelCond(c ConditionCallStructure) {
	cc.cancelcond = &c
}

// UserInChain ...
func (cc *ChainStructure) UserInChain(msg Message) bool {
	_, ok := cc.alreadyin[msg.From.ID]
	return ok
}

// canCall ...
func (cc ChainStructure) canCall(bot TgBot, msg Message) bool {
	if len(cc.chainf) < 1 {
		return false
	}
	index, ok := cc.alreadyin[msg.From.ID]
	if !ok {
		res := cc.chainf[0].canCall(bot, msg)
		if res {
			cc.alreadyin[msg.From.ID] = 0
		}
		return res
	}
	// Check for cancelable!
	if cc.cancelcond != nil {
		condcal := *cc.cancelcond
		if condcal.canCall(bot, msg) {
			delete(cc.alreadyin, msg.From.ID)
			condcal.call(bot, msg)
		}
	}
	if index < 0 || index >= len(cc.chainf) {
		// We are more away, so delete us and test the start again :)
		if cc.loop {
			cc.alreadyin[msg.From.ID] = 0
		} else {
			delete(cc.alreadyin, msg.From.ID)
		}
		res := cc.chainf[0].canCall(bot, msg)
		if res {
			cc.alreadyin[msg.From.ID] = 0
		}
		return res
	}
	return cc.chainf[index].canCall(bot, msg)
}

// call ...
func (cc ChainStructure) call(bot TgBot, msg Message) {
	if cc.canCall(bot, msg) {
		index, _ := cc.alreadyin[msg.From.ID]
		cc.alreadyin[msg.From.ID] = index + 1
		cc.chainf[index].call(bot, msg)
	}
}

// ImageConditionalCall ...
type ImageConditionalCall struct {
	f func(TgBot, Message, []PhotoSize, string)
}

// canCall ...
func (icc ImageConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Photo != nil && len(*msg.Photo) > 0
}

// call ...
func (icc ImageConditionalCall) call(bot TgBot, msg Message) {
	if msg.Photo == nil {
		return
	}
	photos := *msg.Photo
	photoid := ""
	maxsize := 0

	for _, p := range photos {
		mult := p.Width * p.Height
		if mult > maxsize {
			photoid = p.FileID
			maxsize = mult
		}
	}

	icc.f(bot, msg, photos, photoid)
}

// AudioConditionalCall ...
type AudioConditionalCall struct {
	f func(TgBot, Message, Audio, string)
}

// canCall ...
func (icc AudioConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Audio != nil
}

// call ...
func (icc AudioConditionalCall) call(bot TgBot, msg Message) {
	if msg.Audio == nil {
		return
	}
	audio := *msg.Audio
	icc.f(bot, msg, audio, audio.FileID)
}

// NoMessageCall ...
type NoMessageCall struct {
	f func(TgBot, Message)
}

// canCall ...
func (self NoMessageCall) canCall(bot TgBot, msg Message) bool {
	return false
}

// call ...
func (self NoMessageCall) call(bot TgBot, msg Message) {
	self.f(bot, msg)
}

// VoiceConditionalCall ...
type VoiceConditionalCall struct {
	f func(TgBot, Message, Voice, string)
}

// canCall ...
func (icc VoiceConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Voice != nil
}

// call ...
func (icc VoiceConditionalCall) call(bot TgBot, msg Message) {
	if msg.Voice == nil {
		return
	}
	voice := *msg.Voice
	icc.f(bot, msg, voice, voice.FileID)
}

// DocumentConditionalCall ...
type DocumentConditionalCall struct {
	f func(TgBot, Message, Document, string)
}

// canCall ...
func (icc DocumentConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Document != nil
}

// call ...
func (icc DocumentConditionalCall) call(bot TgBot, msg Message) {
	if msg.Document == nil {
		return
	}
	document := *msg.Document
	icc.f(bot, msg, document, document.FileID)
}

// StickerConditionalCall ...
type StickerConditionalCall struct {
	f func(TgBot, Message, Sticker, string)
}

// canCall ...
func (icc StickerConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Sticker != nil
}

// call ...
func (icc StickerConditionalCall) call(bot TgBot, msg Message) {
	if msg.Sticker == nil {
		return
	}
	sticker := *msg.Sticker
	icc.f(bot, msg, sticker, sticker.FileID)
}

// VideoConditionalCall ...
type VideoConditionalCall struct {
	f func(TgBot, Message, Video, string)
}

// canCall ...
func (icc VideoConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Video != nil
}

// call ...
func (icc VideoConditionalCall) call(bot TgBot, msg Message) {
	if msg.Video == nil {
		return
	}
	video := *msg.Video
	icc.f(bot, msg, video, video.FileID)
}

// LocationConditionalCall ...
type LocationConditionalCall struct {
	f func(TgBot, Message, float64, float64)
}

// canCall ...
func (icc LocationConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Location != nil
}

// call ...
func (icc LocationConditionalCall) call(bot TgBot, msg Message) {
	if msg.Location == nil {
		return
	}
	location := *msg.Location
	icc.f(bot, msg, location.Latitude, location.Longitude)
}

// RepliedConditionalCall ...
type RepliedConditionalCall struct {
	f func(TgBot, Message, Message)
}

// canCall ...
func (rcc RepliedConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.ReplyToMessage != nil
}

// call ...
func (rcc RepliedConditionalCall) call(bot TgBot, msg Message) {
	if msg.ReplyToMessage == nil {
		return
	}
	newmsg := *msg.ReplyToMessage
	rcc.f(bot, msg, newmsg)
}

// ForwardConditionalCall ...
type ForwardConditionalCall struct {
	f func(TgBot, Message, User, int)
}

// canCall ...
func (fcc ForwardConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.ForwardFrom != nil && msg.ForwardDate != nil
}

// call ...
func (fcc ForwardConditionalCall) call(bot TgBot, msg Message) {
	if !fcc.canCall(bot, msg) {
		return
	}
	from := *msg.ForwardFrom
	dat := *msg.ForwardDate
	fcc.f(bot, msg, from, dat)
}

// GroupConditionalCall ...
type GroupConditionalCall struct {
	f func(TgBot, Message, int, string)
}

// canCall ...
func (gcc GroupConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Chat.ID < 0 && msg.Chat.Title != nil
}

// call ...
func (gcc GroupConditionalCall) call(bot TgBot, msg Message) {
	if !gcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	dat := *msg.Chat.Title
	gcc.f(bot, msg, from, dat)
}

// NewParticipantConditionalCall ...
type NewParticipantConditionalCall struct {
	f func(TgBot, Message, int, User)
}

// canCall ...
func (npcc NewParticipantConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.NewChatParticipant != nil
}

// call ...
func (npcc NewParticipantConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	who := *msg.NewChatParticipant
	npcc.f(bot, msg, from, who)
}

// LeftParticipantConditionalCall ...
type LeftParticipantConditionalCall struct {
	f func(TgBot, Message, int, User)
}

// canCall ...
func (npcc LeftParticipantConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.LeftChatParticipant != nil
}

// call ...
func (npcc LeftParticipantConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	who := *msg.LeftChatParticipant
	npcc.f(bot, msg, from, who)
}

// NewTitleConditionalCall ...
type NewTitleConditionalCall struct {
	f func(TgBot, Message, int, string)
}

// canCall ...
func (npcc NewTitleConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.NewChatTitle != nil
}

// call ...
func (npcc NewTitleConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	what := *msg.NewChatTitle
	npcc.f(bot, msg, from, what)
}

// NewPhotoConditionalCall ...
type NewPhotoConditionalCall struct {
	f func(TgBot, Message, int, string)
}

// canCall ...
func (npcc NewPhotoConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.NewChatPhoto != nil
}

// call ...
func (npcc NewPhotoConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	what := *msg.NewChatPhoto
	npcc.f(bot, msg, from, what)
}

// DeleteChatPhotoConditionalCall ...
type DeleteChatPhotoConditionalCall struct {
	f func(TgBot, Message, int)
}

// canCall ...
func (npcc DeleteChatPhotoConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.DeleteChatPhoto != nil
}

// call ...
func (npcc DeleteChatPhotoConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	npcc.f(bot, msg, from)
}

// GroupChatCreatedConditionalCall ...
type GroupChatCreatedConditionalCall struct {
	f func(TgBot, Message, int)
}

// canCall ...
func (npcc GroupChatCreatedConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.GroupChatCreated != nil
}

// call ...
func (npcc GroupChatCreatedConditionalCall) call(bot TgBot, msg Message) {
	if !npcc.canCall(bot, msg) {
		return
	}
	from := msg.Chat.ID
	npcc.f(bot, msg, from)
}

// TextConditionalCall ...
type TextConditionalCall struct {
	internal CommandStructure
}

// canCall
func (tcc TextConditionalCall) canCall(bot TgBot, msg Message) bool {
	if msg.Text == nil {
		return false
	}
	text := *msg.Text
	return tcc.internal.canCall(text)
}

// call ...
func (tcc TextConditionalCall) call(bot TgBot, msg Message) {
	if msg.Text == nil {
		return
	}
	text := *msg.Text
	if tcc.internal.canCall(text) {
		tcc.internal.call(bot, msg, text)
	}
}

// CommandStructure ...
type CommandStructure interface {
	canCall(string) bool
	call(TgBot, Message, string)
}

// Simple Regex

// RegexCommand ...
type RegexCommand struct {
	Regex *regexp.Regexp
	f     func(TgBot, Message, []string, map[string]string) *string
}

// canCall ...
func (rc RegexCommand) canCall(text string) bool {
	return rc.Regex.MatchString(text)
}

// call ...
func (rc RegexCommand) call(bot TgBot, msg Message, text string) {
	vals := rc.Regex.FindStringSubmatch(text)
	kvals := findStringSubmatchMap(rc.Regex, text)

	res := rc.f(bot, msg, vals, kvals)
	if res != nil && *res != "" {
		bot.SimpleSendMessage(msg, *res)
	}
}

// Multi Regex

// MultiRegexCommand ...
type MultiRegexCommand struct {
	Regex []*regexp.Regexp
	f     func(TgBot, Message, []string, map[string]string) *string
}

// getRegexMatch
func (rc MultiRegexCommand) getRegexMatch(text string) (bool, *regexp.Regexp) {
	for _, r := range rc.Regex {
		if r.MatchString(text) {
			return true, r
		}
	}
	return false, nil
}

// canCall ...
func (rc MultiRegexCommand) canCall(text string) bool {
	res, _ := rc.getRegexMatch(text)
	return res
}

// call ...
func (rc MultiRegexCommand) call(bot TgBot, msg Message, text string) {
	canC, regexToUse := rc.getRegexMatch(text)
	if !canC {
		fmt.Println("Error")
		return
	}
	vals := regexToUse.FindStringSubmatch(text)
	kvals := findStringSubmatchMap(regexToUse, text)

	res := rc.f(bot, msg, vals, kvals)
	if res != nil && *res != "" {
		bot.SimpleSendMessage(msg, *res)
	}
}

// SimpleCommandFuncStruct struct wrapper for simple command funcs
type SimpleCommandFuncStruct struct {
	f func(TgBot, Message, string) *string
}

// CallSimpleCommandFunc wrapper for simple functions
func (scf SimpleCommandFuncStruct) CallSimpleCommandFunc(bot TgBot, msg Message, m []string, km map[string]string) *string {
	res := ""
	if msg.Text != nil {
		res2 := scf.f(bot, msg, *msg.Text)
		if res2 != nil {
			res = *res2
		}
	}
	return &res
}
