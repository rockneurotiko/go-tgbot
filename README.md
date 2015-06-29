# go-tgbot
Telegram API bot wrapper for Go Language! &lt;3

This is a beauty way to build telegram bots with Go language.

Almost all methods have been added, and all features will be available soon. If you want a feature that hasn't been added yet or something is broken, open an issue and let's build it!

## Disclaimer

This is the first time I write Go code, so, if you see something that I'm doing bad, please, tell me, I love to learn :-)

You can talk to me in [telegram](https://telegram.me/rock_neurotiko)

## Example

`Show me the code!`

```go
package main

import (
	"fmt"

	"github.com/rockneurotiko/go-tgbot"
)

func echoHandler(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	newmsg := fmt.Sprintf("[Echoed]: %s", vals[1])
	return &newmsg
}

func main() {
	bot := tgbot.NewTgBot("token").
		CommandFn(`echo (.+)`, echoHandler)
	bot.SimpleStart()
}
```

You can see the code in [the echo example](https://github.com/rockneurotiko/go-tgbot/blob/master/example/echoexample/main.go)

## Installation

As all the go libraries, you can install it with the `go` tool:

```
go get -u github.com/rockneurotiko/go-tgbot
```

## Receiving messages!

First, you need to create a new TgBot:
```go
bot := tgbot.NewTgBot("token")
```

The `"token"` is the token that [@botfather](https://telegram.me/botfather) gives you when you create a new bot.

After that, you can add your functions that will be executed, right now only support text messages, but in a short time you will can add custom conditions of messages, and file callbacks. This functions are called in a goroutine, so don't worry to "hang" the bot :-)

Currently, there are two function signatures that you can use:

- The simplest one, this should be used for messages that don't have parameters, like `/help`
  ```go
  func(TgBot, Message, string) *string
  ```
  The parameters are:
  - TgBot: the instance of the bot, so you can call functions inside that one.
  - Message: The Message struct that represents it, so you can get any param.
  - string: The message in a string :)

- The complex one, this should be used when you gave complex regular expressions.
  ```go
  func(TgBot, Message, []string, map[string]string) *string
  ```
  The parameters are:
  - TgBot: the instance of the bot, so you can call functions inside that one.
  - Message: The Message struct that represents it, so you can get any param.
  - []string: This will be the captures groups in the regular expression, easy to get them ^^
  - map[string]string: This are the named capture groups, much more ealy to get them!!


With this two kinds of function, you can build your function calls with commands in the `TgBot` instance, all the functions that start with `Simple` uses the simplest function call.

Before of explain them, you have to know what a command is. A command is what Telegram API understand as a command, that ones that [@botfather](https://telegram.me/botfather) let you define with `/setcommands` function.

That commands always look like `/<command>`, but they can have other parameters `/<command> <param1> <param2> ...`, we'll see that later.

The curious thing is that the commands can be called as `/<command>` or `/<command>@username`, this is useful when you are in a group and you want to specify the bot to send that command. If you use this functions, you don't have to worry about adding or handling the @username, the library will handle it magically for you &lt;3

Also, more magic is that you don't need to write the `/` command, neither the safe-command characters for the expression, that are the starting `^` and the leading `$`, so, if you say that want the commasd `help`, the library will understand you and make `^/help(?:@username)?$` :)

So, let's stop talking and let's see the functions that you can use, in the [simpleexample file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/simpleexample/main.go) you can see an example for every of this ^^

- `SimpleCommandFn(string, func(TgBot, Message, string) *string)`:
  This is the basic one, when you want a command without arguments, like the basic `/start` or `/help`, you just have to create a `simple` function that we saw before, and add it to a command.
  For example, this code will add a simple command handler (`helpHandler`) to the `/help` command, and it will be called properly with a `/help` command and `/help@username` ^^
  ```go
  bot := tgbot.NewTgBot("token").
    SimpleCommandFn(`^/help$`, helpHandler)
  bot.SimpleStart()
  ```

- `CommandFn(string, func(TgBot, Message, []string, map[string]string) *string)`:
  Sometimes, you don't want a simple command, you want something more interesting that take parameters, this function will help you.
  This snippet code will add a command handlerp (not a simple one) (`echoHandler`) to the `/echo` command that will take a parameter, it will be called properly if `/echo@username` is used :smile:
  ```go
  bot := tgbot.NewTgBot("token").
    CommandFn(`^/echo (.+)`, echoHandler)
  bot.SimpleStart()
  ```
  The `vals []string` parameter in the echoHandler will be of size 2, the first one is the complete text, and the second one is what the capture group `(.+)` handles.

- `MultiCommandFn([]string, func(TgBot, Message, []string, map[string]string) *string)`:
  Other times, you want multiple commands, for example, you want a `/help` and `/help <command>`, of course, you can build a regexp that matches that, or build it with two different functions, but there can be more complicated, and this is a beauty way to build this.
  You just give him a list of regular expressions that will try to execute in order, but will only execute one of them.
  ```go
  bot := tgbot.NewTgBot("token").
    MultiCommandFn([]string{`^/help (\w+)$`, `^/help$`}, helpHand).
  bot.SimpleStart()
  ```

That's all the functions for working with commands! You have many choices to build, and many times they are interchangeable, so you can use the one you prefer =D

After see the commands, there are some messages that are not commands, but you still want to do something in that messages! So, the regex functions come to save you!

And why to use different functions? Two reasons, first one, you can know what's a `real` command and what not, just for looking in the call, second one, because the command functions do magic to handle the `@username` thing, won't work properly in custom regexp ;-)

Actually, this functions work exactly the same as their command-like function, just that don't add the `@username` magic :)

- `SimpleRegexFn(string, func(TgBot, Message, string) *string)`:
  ```go
  bot := tgbot.NewTgBot("token").
    SimpleRegexFn(`^Hello!$`, helloHand)
  bot.SimpleStart()
  ```

- `RegexFn([]string, func(TgBot, Message, []string, map[string]string) *string)`:
  ```go
  bot := tgbot.NewTgBot("token").
    RegexFn(`^Repet this: (.+)$`, repeatHand)
  bot.SimpleStart()
  ```

- `MultiRegexFn([]string, func(TgBot, Message, []string, map[string]string) *string)`:
  (Sorry for this bad example, but no one comed to my mind, if you have some good example, please tell me!)
  ```go
  bot := tgbot.NewTgBot("token").
    MultiRegexFn([]string{`^First regex$`, `^Second regex (.*)`}, multiregHand)
  bot.SimpleStart()
  ```


Now, let's see the callback functions that are not text.

- `AnyMsgFn(func(TgBot, Message))`
  This functions will be called in every message, be careful!

## Doing actions!

So, you know how to get your functions call when something arrives, but how can you answer? That's really important! If the bot can't answer, then you don't have a bot! So, let's talk about the action functions availables in TgBot

The basic action is send a text, and this is really simple here, did you see that the functions had a `*string` parameter at the end? Well, that's because the string that you return, will be sended to the chat that the message comes from. Easy! Why a pointer? to allow returning `nil`, but I don't like this, if you have a better idea, please tell me!

For example, this function will send to the sender (person or group) the same message:
```go
func example(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
    return &text
}
```

You can do other actions too! Did you see that the first parameter is the TgBot instance? That's to allow you doing actions!

All the actions have a "pure" function that just sends the query, you can call them directry, they are called like `ActionNameQuery`, for example, `SendMessageQuery` or `ForwardMessageQuery`, but it's better to use the custom functions:

- `SendMessage` functions:

    - `SimpleSendMessage(msg Message, text string) (Message, error)`: Simplified call with the message and a string, and it will send that string to the sender.

    - `SendMessage(chatid int, text string, disable_web_preview *bool, reply_to_message_id *int, reply_markup *ReplyMarkupInt) ResultWithMessage`: Send a message with all parameters, the chat id (you can acces with msg.Chat.ID), the string to send, and two pointers (because are optional, so if you don't want them, just pass `nil`), disable\_web\_preview, reply\_to\_message\_id, and reply\_markup, that is an interface, and the structs you can use are: `ReplyKeyboardMarkup`, `ReplyKeyboardHide` and `ForceReply`, but for this, better use the following functions.

  - `SendMessageWithKeyboard(chatid int, text string, disable_web_preview *bool, reply_to_message_id *int, reply_markup ReplyKeyboardMarkup) ResultWithMessage`: This makes easier to send a keyboard, just pass the struct :)

  - `SendMessageWithKeyboardHide(chatid int, text string, disable_web_preview *bool, reply_to_message_id *int, reply_markup ReplyKeyboardHide) ResultWithMessage`: This makes easier to send a the hide keyboard, just pass the struct :)

  - `SendMessageWithForceReply(chatid int, text string, disable_web_preview *bool, reply_to_message_id *int, reply_markup ForceReply) ResultWithMessage`: This makes easier to send a force reply, just pass the struct :)

  - `SendMessageQuery(payload QuerySendMessage) ResultWithMessage`: Try not to use this :)
- `ForwardMessage` functions:

  - `ForwardMessage(chatid int, fromid int, messageid int) ResultWithMessage`: Will forward to `chatid` saying that comes from `fromid` the message `messageid`

  - `ForwardMessageQuery(payload ForwardMessageQuery) ResultWithMessage`: Try to don't use this :)


## Full example!

This is a full example using all we saw until now, you can found this example in the [simpleexample/main.go file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/simpleexample/main.go).

```go
package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/rockneurotiko/go-tgbot"
)

const (
	token = "awesometoken"
)

var availableCommands = map[string]string{
	"/start":          "Start the bot!",
	"/help":           "Get help!!",
	"/helpbotfather":  "Get the help formatted to botfather",
	"/help <command>": "Get the help of one command",
	"/keyboard":       "Send you a keyboard",
	"/hidekeyboard":   "Hide the keyboard",
	"/hardecho":       "Echo with force reply",
	"/forwardme":      "Forward that message to you",
	"/sleep":          "Sleep for 5 seconds, without blocking, awesome goroutines",
	"/showmecommands": "Returns you a keyboard with the simplest commands",
}

func buildHelpMessage(complete bool) string {
	var buffer bytes.Buffer
	for cmd, htext := range availableCommands {
		str := ""
		if complete {
			str = fmt.Sprintf("%s - %s\n", cmd, htext)
		} else if len(strings.Split(cmd, " ")) == 1 {
			str = fmt.Sprintf("%s - %s\n", cmd[1:], htext)
		}
		buffer.WriteString(str)
	}
	return buffer.String()
}

func hideKeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	rkm := tgbot.ReplyKeyboardHide{HideKeyboard: true, Selective: false}
	bot.SendMessageWithKeyboardHide(msg.Chat.ID, "Hiden it!", nil, nil, rkm)
	return nil
}

func cmdKeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{"I", "<3"}, {"You"}}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: false,
		Selective:       false}
	bot.SendMessageWithKeyboard(msg.Chat.ID, "Enjoy the keyboard", nil, nil, rkm)
	return nil
}

func hardEcho(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	msgtext := ""
	if len(vals) > 1 {
		msgtext = vals[1]
	}
	rkm := tgbot.ForceReply{Force: true, Selective: false}
	bot.SendMessageWithForceReply(msg.Chat.ID, msgtext, nil, nil, rkm)
	return nil
}

func forwardHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.ForwardMessage(msg.Chat.ID, msg.Chat.ID, msg.ID)
	return nil
}

func helloHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	msgr := fmt.Sprintf("Hi %s! <3", msg.From.FirstName)
	return &msgr
}

func tellmeHand(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	msgtext := ""
	if len(vals) > 1 {
		msgtext = vals[1]
	}
	return &msgtext
}

func multiregexHelpHand(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	if len(vals) > 1 {
		for k, v := range availableCommands {
			if k[1:] == vals[1] {
				res := v
				return &res
			}
		}
	}
	res := ""
	if vals[0] == "/help" {
		res = buildHelpMessage(true)
	} else if vals[0] == "/helpbotfather" {
		res = buildHelpMessage(false)
	}
	return &res
}

func testGoroutineHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.SimpleSendMessage(msg, "Starting")
	time.Sleep(5000 * time.Millisecond)
	r := "Ending"
	return &r
}

func showMeHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{}}
	for k := range availableCommands {
		if len(strings.Split(k, " ")) == 1 {
			if len(keylayout[len(keylayout)-1]) == 2 {
				keylayout = append(keylayout, []string{k})
			} else {
				keylayout[len(keylayout)-1] = append(keylayout[len(keylayout)-1], k)
			}
		}
	}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       false}
	bot.SendMessageWithKeyboard(msg.Chat.ID, "There you have the commands!", nil, nil, rkm)
	return nil
}

func allMsgHand(bot tgbot.TgBot, msg tgbot.Message) {
	// uncomment this to see it :)
	fmt.Printf("Received message: %+v\n", msg.ID)
	// bot.SimpleSendMessage(msg, "Received message!")
}

func conditionFunc(bot tgbot.TgBot, msg tgbot.Message) bool {
	return msg.Photo != nil
}

func conditionCallFunc(bot tgbot.TgBot, msg tgbot.Message) {
	bot.SimpleSendMessage(msg, "Nice image :)")
}

func main() {
	bot := tgbot.NewTgBot(token).
		SimpleCommandFn(`sleep`, testGoroutineHand).
		SimpleCommandFn(`keyboard`, cmdKeyboard).
		SimpleCommandFn(`hidekeyboard`, hideKeyboard).
		SimpleCommandFn(`forwardme`, forwardHand).
		SimpleCommandFn(`showmecommands`, showMeHand).
		CommandFn(`hardecho (.+)`, hardEcho).
		MultiCommandFn([]string{`help (\w+)`, `help`, `helpbotfather`}, multiregexHelpHand).
		SimpleRegexFn(`^Hello!$`, helloHand).
		RegexFn(`^Tell me (.+)$`, tellmeHand).
		AnyMsgFn(allMsgHand).
		CustomFn(conditionFunc, conditionCallFunc)

	// bot := tgbot.NewTgBot(token)
	// bot.SimpleCommandFn(`^/sleep$`, testGoroutineHand)
	// bot.SimpleCommandFn(`^/keyboard$`, cmdKeyboard)
	// bot.SimpleCommandFn(`^/hidekeyboard$`, hideKeyboard)
	// bot.SimpleCommandFn(`^/forwardme$`, forwardHand)
	// bot.SimpleCommandFn(`^/showmecommands`, showMeHand)
	// bot.CommandFn(`^/hardecho (.+)`, hardEcho)
	// bot.MultiCommandFn([]string{`^/help (\w+)$`, `^/help$`, `^/helpbotfather$`}, multiregexHelpHand)
	// bot.SimpleRegexFn(`^Hello!$`, helloHand)
	// bot.RegexFn(`^Tell me (.+)$`, tellmeHand)
	// bot.AnyMsgFn(allMsgHand)
	// bot.CustomFn(conditionFunc, conditionCallFunc)
	bot.SimpleStart()
}
```

If you want to handle the message by yourself, you can too, you will have to add a channel to the listener and start it.

See the [manualexample/main.go file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/manualexample/main.go) to see an example of manual handling :smile:

## What is done and what left!

You are welcome to help in building this project :smile: &lt;3

- [ ] Callback functions
  - [x] Simple command
  - [x] Command with parameters
  - [x] Multiple commands
  - [x] Simple regular expression
  - [x] Normal regular expression
  - [x] Multiple regular expressions
  - [x] Any message
  - [x] On custom function
  - [ ] On image
  - [ ] On audio
  - [ ] On document
  - [ ] On video
  - [ ] On location
  - [ ] On message replied
  - [ ] On message forwarded
  - [ ] On any group event
  - [ ] On new chat participant
  - [ ] On left chat participant
  - [ ] On new chat title
  - [ ] On new chat photo
  - [ ] On delete chat photo
  - [ ] On group chat created

- [ ] Action functions
  - [x] Get me
  - [x] Send message
    - [x] easy with keyboard
    - [x] easy with force reply
    - [x] easy with keyboard hide
  - [x] Forward message
  - [x] getUpdates
    - [x] This is done automatically when you use the `SimpleStart()` or `Start()`, you shouldn't touch this ;-)
  - [ ] setWebhook
    - [ ] This is done automatically when you use `ServerStart()`, you shoulnd't touch this
  - [ ] Send photo
   - [ ] From id
   - [ ] From file
   - [ ] easy with keyboard
   - [ ] easy with force reply
   - [ ] easy with keyboard hide
  - [ ] Send audio
   - [ ] From id
   - [ ] From file
   - [ ] easy with keyboard
   - [ ] easy with force reply
   - [ ] easy with keyboard hide
  - [ ] Send document
   - [ ] From id
   - [ ] From file
   - [ ] easy with keyboard
   - [ ] easy with force reply
   - [ ] easy with keyboard hide
  - [ ] Send sticker
   - [ ] From id
   - [ ] From file
   - [ ] easy with keyboard
   - [ ] easy with force reply
   - [ ] easy with keyboard hide
  - [ ] Send video
   - [ ] From id
   - [ ] From file
   - [ ] easy with keyboard
   - [ ] easy with force reply
   - [ ] easy with keyboard hide
  - [ ] Send location
  - [ ] Send chat action
  - [ ] Get user profile photos

- [ ] Other nice things!
 - [ ] Default options for messages configured before start.
   - [ ] Disable webpage preview
   - [ ] Reply to the message
   - [ ] Selective the reply_markup
   - [ ] One time keyboard
  - [ ] Easy to work with authorized users
  - [ ] Easy to work with "flow" messages
