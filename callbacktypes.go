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

// Custom functions for CustomCall :)

// AlwaysReturnTrue ...
func AlwaysReturnTrue(bot TgBot, msg Message) bool {
	return true
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
	if len(photos) > 0 {
		photoid = photos[0].FileID
	}
	icc.f(bot, msg, photos, photoid)
}

// canCall ...
func (cc CustomCall) canCall(bot TgBot, msg Message) bool {
	return cc.condition(bot, msg)
}

// call ...
func (cc CustomCall) call(bot TgBot, msg Message) {
	cc.f(bot, msg)
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

// TextConditionalCall ...
type TextConditionalCall struct {
	internal CommandStructure
}

// canCall
func (tcc TextConditionalCall) canCall(bot TgBot, msg Message) bool {
	return msg.Text != nil
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
	kvals := FindStringSubmatchMap(rc.Regex, text)

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
	kvals := FindStringSubmatchMap(regexToUse, text)

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
