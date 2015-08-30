package tgbot

import (
	"fmt"
	"strings"
)

func (bot TgBot) addUsernameCommand(expr string) string {
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

func (bot *TgBot) addToConditionalFuncs(cf ConditionCallStructure) {
	if !bot.BuildingChain {
		bot.TestConditionalFuncs = append(bot.TestConditionalFuncs, cf)
	} else {
		if len(bot.ChainConditionals) > 0 {
			bot.ChainConditionals[len(bot.ChainConditionals)-1].AddToConditionalFuncs(cf)
		}
	}
}

func (bot TgBot) cleanMessage(msg Message) Message {
	if msg.Text != nil {
		if bot.DefaultOptions.CleanInitialUsername {
			text := *msg.Text
			username := fmt.Sprintf("@%s", bot.Username)
			if strings.HasPrefix(text, username) {
				text = strings.TrimSpace(strings.Replace(text, username, "", 1)) // Replace one time
				if bot.DefaultOptions.AllowWithoutSlashInMention &&
					!strings.HasSuffix(text, "/") {
					text = "/" + text
				}
				msg.Text = &text
			}
		}

		if bot.DefaultOptions.LowerText {
			text := strings.ToLower(*msg.Text)
			msg.Text = &text
		}
	}
	return msg
}
