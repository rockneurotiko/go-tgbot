package tgbot

import "regexp"

// CommandFn Add a command function, with capture groups and/or named capture groups.
func (bot *TgBot) CommandFn(path string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.addUsernameCommand(path)
	r := regexp.MustCompile(path)

	bot.addToConditionalFuncs(TextConditionalCall{RegexCommand{r, f}})
	return bot
}

// SimpleCommandFn Add a simple command function.
func (bot *TgBot) SimpleCommandFn(path string, f func(TgBot, Message, string) *string) *TgBot {
	path = convertToCommand(path)
	path = bot.addUsernameCommand(path)
	r := regexp.MustCompile(path)
	newf := SimpleCommandFuncStruct{f}

	bot.addToConditionalFuncs(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	return bot
}

// MultiCommandFn add multiples commands with capture groups. Only one of this will be executed.
func (bot *TgBot) MultiCommandFn(paths []string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	rc := []*regexp.Regexp{}
	for _, p := range paths {
		p = convertToCommand(p)
		p = bot.addUsernameCommand(p)
		r := regexp.MustCompile(p)
		rc = append(rc, r)
	}

	bot.addToConditionalFuncs(TextConditionalCall{MultiRegexCommand{rc, f}})
	return bot
}

// RegexFn add a regular expression function with capture groups and/or named capture groups.
func (bot *TgBot) RegexFn(path string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	r := regexp.MustCompile(path)

	bot.addToConditionalFuncs(TextConditionalCall{RegexCommand{r, f}})
	return bot
}

// SimpleRegexFn add a simple regular expression function.
func (bot *TgBot) SimpleRegexFn(path string, f func(TgBot, Message, string) *string) *TgBot {
	r := regexp.MustCompile(path)
	newf := SimpleCommandFuncStruct{f}

	bot.addToConditionalFuncs(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	return bot
}

// MultiRegexFn add multiples regular expressions with capture groups. Only one will be executed.
func (bot *TgBot) MultiRegexFn(paths []string, f func(TgBot, Message, []string, map[string]string) *string) *TgBot {
	rc := []*regexp.Regexp{}
	for _, p := range paths {
		r := regexp.MustCompile(p)
		rc = append(rc, r)
	}

	bot.addToConditionalFuncs(TextConditionalCall{MultiRegexCommand{rc, f}})
	return bot
}

// ImageFn add a function to be called when an image arrives.
func (bot *TgBot) ImageFn(f func(TgBot, Message, []PhotoSize, string)) *TgBot {
	bot.addToConditionalFuncs(ImageConditionalCall{f})
	return bot
}

// AudioFn  add a function to be called when an audio arrives.
func (bot *TgBot) AudioFn(f func(TgBot, Message, Audio, string)) *TgBot {
	bot.addToConditionalFuncs(AudioConditionalCall{f})
	return bot
}

// VoiceFn  add a function to be called when an audio arrives.
func (bot *TgBot) VoiceFn(f func(TgBot, Message, Voice, string)) *TgBot {
	bot.addToConditionalFuncs(VoiceConditionalCall{f})
	return bot
}

// DocumentFn add a function to be called when a document arrives.
func (bot *TgBot) DocumentFn(f func(TgBot, Message, Document, string)) *TgBot {
	bot.addToConditionalFuncs(DocumentConditionalCall{f})
	return bot
}

// StickerFn add a function to be called when a sticker arrives.
func (bot *TgBot) StickerFn(f func(TgBot, Message, Sticker, string)) *TgBot {
	bot.addToConditionalFuncs(StickerConditionalCall{f})
	return bot
}

// VideoFn add a function to be called when a video arrives.
func (bot *TgBot) VideoFn(f func(TgBot, Message, Video, string)) *TgBot {
	bot.addToConditionalFuncs(VideoConditionalCall{f})
	return bot
}

// LocationFn add a function to be called when a location arrives.
func (bot *TgBot) LocationFn(f func(TgBot, Message, float64, float64)) *TgBot {
	bot.addToConditionalFuncs(LocationConditionalCall{f})
	return bot
}

// ReplyFn add a function to be called when a message replied other is arrives.
func (bot *TgBot) ReplyFn(f func(TgBot, Message, Message)) *TgBot {
	bot.addToConditionalFuncs(RepliedConditionalCall{f})
	return bot
}

// ForwardFn add a function to be called when a message forwarding other arrives.
func (bot *TgBot) ForwardFn(f func(TgBot, Message, User, int)) *TgBot {
	bot.addToConditionalFuncs(ForwardConditionalCall{f})
	return bot
}

// GroupFn add a function to be called in every group message.
func (bot *TgBot) GroupFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.addToConditionalFuncs(GroupConditionalCall{f})
	return bot
}

// NewParticipantFn add a function to be called when new participant is received.
func (bot *TgBot) NewParticipantFn(f func(TgBot, Message, int, User)) *TgBot {
	bot.addToConditionalFuncs(NewParticipantConditionalCall{f})
	return bot
}

// LeftParticipantFn add a function to be called when a participant left.
func (bot *TgBot) LeftParticipantFn(f func(TgBot, Message, int, User)) *TgBot {
	bot.addToConditionalFuncs(LeftParticipantConditionalCall{f})
	return bot
}

// NewTitleChatFn add a function to be called when the title of a group is changed.
func (bot *TgBot) NewTitleChatFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.addToConditionalFuncs(NewTitleConditionalCall{f})
	return bot
}

// NewPhotoChatFn add a function to be called when the photo of a chat is changed.
func (bot *TgBot) NewPhotoChatFn(f func(TgBot, Message, int, string)) *TgBot {
	bot.addToConditionalFuncs(NewPhotoConditionalCall{f})
	return bot
}

// DeleteChatPhotoFn add a function to be called when the photo of a chat is deleted.
func (bot *TgBot) DeleteChatPhotoFn(f func(TgBot, Message, int)) *TgBot {
	bot.addToConditionalFuncs(DeleteChatPhotoConditionalCall{f})
	return bot
}

// GroupChatCreatedFn add a function to be called when a group chat is created.
func (bot *TgBot) GroupChatCreatedFn(f func(TgBot, Message, int)) *TgBot {
	bot.addToConditionalFuncs(GroupChatCreatedConditionalCall{f})
	return bot
}

// AnyMsgFn add a function to be called in every message :)
func (bot *TgBot) AnyMsgFn(f func(TgBot, Message)) *TgBot {
	bot.addToConditionalFuncs(CustomCall{AlwaysReturnTrue, f})
	return bot
}

func (bot *TgBot) NotCalledFn(f func(TgBot, Message)) *TgBot {
	bot.NoMessageFuncs = append(bot.NoMessageFuncs, NoMessageCall{f})
	return bot
}

// CustomFn add a function to be called with a custom conditional function.
func (bot *TgBot) CustomFn(cond func(TgBot, Message) bool, f func(TgBot, Message)) *TgBot {
	bot.addToConditionalFuncs(CustomCall{cond, f})
	return bot
}

// StartChain will start a chain process, all the functions you add after this will be part of the same chain.
func (bot *TgBot) StartChain() *TgBot {
	bot.ChainConditionals = append(bot.ChainConditionals, NewChainStructure())
	bot.BuildingChain = true
	return bot
}

// CancelChainCommand add a special command that cancel the current chain
func (bot *TgBot) CancelChainCommand(path string, f func(TgBot, Message, string) *string) *TgBot {
	if !bot.BuildingChain {
		return bot
	}
	if len(bot.ChainConditionals) > 0 {

		path = convertToCommand(path)
		path = bot.addUsernameCommand(path)
		r := regexp.MustCompile(path)
		newf := SimpleCommandFuncStruct{f}
		bot.ChainConditionals[len(bot.ChainConditionals)-1].
			SetCancelCond(TextConditionalCall{RegexCommand{r, newf.CallSimpleCommandFunc}})
	}
	return bot
}

// LoopChain will make the chain start again when the last action is done.
func (bot *TgBot) LoopChain() *TgBot {
	if !bot.BuildingChain {
		return bot
	}
	if len(bot.ChainConditionals) > 0 {
		bot.ChainConditionals[len(bot.ChainConditionals)-1].SetLoop(true)
	}
	return bot
}

// EndChain ends the chain, after this, the functions will be added as always.
func (bot *TgBot) EndChain() *TgBot {
	bot.BuildingChain = false
	return bot
}
