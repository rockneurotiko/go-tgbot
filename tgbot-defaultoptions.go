package tgbot

import "github.com/oleiade/reflections"

// DefaultOptionsBot represents the options that the bot will try to apply automatically
type DefaultOptionsBot struct {
	DisableWebURL              *bool
	Selective                  *bool
	OneTimeKeyboard            *bool
	CleanInitialUsername       bool
	AllowWithoutSlashInMention bool
	LowerText                  bool
	RecoverPanic               bool
}

func (bot *TgBot) SetRecoverPanic(b bool) *TgBot {
	bot.DefaultOptions.RecoverPanic = b
	return bot
}

func (bot *TgBot) SetLowerText(b bool) *TgBot {
	bot.DefaultOptions.LowerText = b
	return bot
}

// DefaultDisableWebpagePreview ...
func (bot *TgBot) DefaultDisableWebpagePreview(b bool) *TgBot {
	bot.DefaultOptions.DisableWebURL = &b
	return bot
}

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

func hookDisableWebpage(payload interface{}, nv *bool) {
	if nv != nil {
		has, _ := reflections.HasField(payload, "DisableWebPagePreview")
		if has {
			value, _ := reflections.GetField(payload, "DisableWebPagePreview")
			bvalue := value.(*bool)
			if bvalue == nil {
				reflections.SetField(payload, "DisableWebPagePreview", nv)
			}
		}
	}
}

func hookReplyToMessageID(payload interface{}, nv *bool) {
	if nv != nil {
		has, _ := reflections.HasField(payload, "ReplyToMessageID")
		if has {
			value, _ := reflections.GetField(payload, "ReplyToMessageID")
			bvalue := value.(*int)
			if bvalue == nil {
				reflections.SetField(payload, "ReplyToMessageID", nv)
			}
		}
	}
}

func hookSelective(payload interface{}, nv *bool) {
	if nv != nil {
		has, _ := reflections.HasField(payload, "Selective")
		if has {
			value, _ := reflections.GetField(payload, "Selective")
			bvalue := value.(*bool)
			if bvalue == nil {
				reflections.SetField(payload, "Selective", nv)
			}
		}
	}
}

func hookOneTimeKeyboard(payload interface{}, nv *bool) {
	if nv != nil {
		has, _ := reflections.HasField(payload, "OneTimeKeyboard")
		if has {
			value, _ := reflections.GetField(payload, "OneTimeKeyboard")
			bvalue := value.(*bool)
			if bvalue == nil {
				reflections.SetField(payload, "OneTimeKeyboard", nv)
			}
		}
	}
}

// HookPayload I hate reflection, sorry for that <3
func hookPayload(payload interface{}, opts DefaultOptionsBot) {
	hookDisableWebpage(payload, opts.DisableWebURL)
	// HookReplyToMessageID(payload, opts.ReplyToMessageID)

	has, _ := reflections.HasField(payload, "ReplyMarkup")

	if has {
		keyint, _ := reflections.GetField(payload, "ReplyMarkup")

		switch val := keyint.(type) {
		case *ForceReply, *ReplyKeyboardHide:
			if val != nil {
				hookSelective(val, opts.Selective)
			}
		case *ReplyKeyboardMarkup:
			if val != nil {
				hookOneTimeKeyboard(val, opts.OneTimeKeyboard)
				hookSelective(val, opts.Selective)
			}
		default:
		}
	}
}
