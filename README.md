# Telegram Bot API library! [![GoDoc](https://godoc.org/github.com/rockneurotiko/go-tgbot?status.png)](http://godoc.org/github.com/rockneurotiko/go-tgbot)
Telegram API bot wrapper for Go (golang) Language! &lt;3

This is a beauty way to build telegram bots with golang.

Almost all methods have been added, and all features will be available soon. If you want a feature that hasn't been added yet or something is broken, open an issue and let's build it!

Also, if you develop a bot with this, I would love to hear about it! ^^

## Disclaimer

This is the first time I write Go code, so, if you see something that I'm doing bad, please, tell me, I love to learn :-)

Also, some people had tell me that the way this library have to handle the functions is "too javascript". I am trying to be as much Go-like as I know, and the way I build the library chains are based in `mux` and `gorequest`, if you know a most Go-like way of build this, please, tell me! (I hate JS, and have heard that my library looks like JS made me sad xD)

You can talk to me in [telegram](https://telegram.me/rockneurotiko)

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

As all the go libraries, you can install it with the `go get` tool:

```
go get -u github.com/rockneurotiko/go-tgbot
```

## Receiving messages!

First, you need to create a new TgBot:
```go
bot := tgbot.NewTgBot("token")
```

The `"token"` is the token that [@botfather](https://telegram.me/botfather) gives you when you create a new bot.

After that, you can add your functions that will be executed, see bellow to see the the different ways and conditions for your functions, choose the proper one. This functions are called in a goroutine, so don't worry to "hang" the bot :-)

### Call in text messages (The typical)

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
  - map[string]string: This are the named capture groups, much more easy to get them!!


With this two kinds of functions, you can build your function calls with commands in the `TgBot` instance, all the functions that start with `Simple` uses the simplest function call.

Before of explain them, you have to know what a command is. A command is what Telegram API understand as a command, that ones that [@botfather](https://telegram.me/botfather) let you define with `/setcommands` function.

That commands always look like `/<command>`, but they can have other parameters `/<command> <param1> <param2> ...`, we'll see that later.

The curious thing is that the commands can be called as `/<command>` or `/<command>@username`, this is useful when you are in a group and you want to specify the bot to send that command. If you use this functions, you don't have to worry about adding or handling the @username, the library will handle it magically for you &lt;3

Also, more magic is that you don't need to write the `/` command, neither the safe-command characters for the expression, that are the starting `^` and the leading `$`, so, if you say you want the command `help`, the library will understand you and make `^/help(?:@username)?$` :)

So, let's stop talking and let's see the functions that you can use, in the [simpleexample file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/simpleexample/main.go) you can see an example for every one ^^

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
  This snippet code will add a command handler (not a simple one) (`echoHandler`) to the `/echo` command that will take a parameter, it will be called properly if `/echo@username` is used :smile:
  ```go
  bot := tgbot.NewTgBot("token").
    CommandFn(`^/echo (.+)`, echoHandler)
  bot.SimpleStart()
  ```
  The `vals []string` parameter in the echoHandler will be of size 2, the first one is the complete text, and the second one is what the capture group `(.+)` handles.

- `MultiCommandFn([]string, func(TgBot, Message, []string, map[string]string) *string)`:
  Other times, you want multiple commands, for example, you want a `/help` and `/help <command>`, of course, you can build a regexp that matches that, or build it with two different functions, but there can be more complicated situations, and this is a beauty way to build this.
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


### Call in file messages.

- `ImageFn(func(TgBot, Message, []PhotoSize, string))`
  Function to be called when an image is received, the two extra parameters are an array of PhotoSize, and the ID of the file.

### Other miscellaneous calls.

- `AnyMsgFn(func(TgBot, Message))`
  This functions will be called in every message, be careful!

- `CustomFn(func(TgBot, Message) bool, func(TgBot, Message))`
  With this callbacks you can add your custom conditions, first it will execute the first function, if the return value is true, execute the second one.


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

### Message actions

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

### File actions

- `SendPhoto` functions, wherever you see the `path`, can be a file path or a file id :smile:, it's handled automatically:

  - `SimpleSendPhoto(msg Message, path string) (Message, error)`: Simplified call with the path and a string, and it will send that string to the sender.

  - `SendPhoto(chatid int, path string, caption *string, reply_to_message_id *int, reply_markup *ReplyMarkupInt) ResultWithMessage`: Like the SendMessage, but sending the photo, use this for full control over the parameters.

  - `SendPhotoWithKeyboard(chatid int, path string, caption *string, reply_to_message_id *int, reply_markup ReplyKeyboardMarkup) ResultWithMessage`: This makes easier to send a keyboard, just pass the struct instead of a pointer to an interface :)

  - `SendPhotoWithKeyboardHide(chatid int, path string, caption *string, reply_to_message_id *int, reply_markup ReplyKeyboardHide) ResultWithMessage`: This makes easier to send a the hide keyboard, just pass the struct instead of a pointer to an interface  :)

  - `SendPhotoWithForceReply(chatid int, path string, caption *string, reply_to_message_id *int, reply_markup ForceReply) ResultWithMessage`: This makes easier to send a force reply, just pass the struct instead of a pointer to an interface  :)

  - `SendPhotoQuery(payload interface{}) ResultWithMessage`: Try not to use this :) (btw, the interface{} is checked agains SendPhotoIDQuery and SendPhotoPathQuery)


## Full examples!

You can found a full example with all the functions/calls in the [simpleexample/main.go file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/simpleexample/main.go).

If you want to handle the messages by yourself, you can too, you will have to add a channel to the listener and start it.

See the [manualexample/main.go file](https://github.com/rockneurotiko/go-tgbot/blob/master/example/manualexample/main.go) to see an example of manual handling :smile:

## What is done and what left!

You are welcome to help in building this project :smile: &lt;3

- [x] Callback functions
  - [x] Simple command
  - [x] Command with parameters
  - [x] Multiple commands
  - [x] Simple regular expression
  - [x] Normal regular expression
  - [x] Multiple regular expressions
  - [x] Any message
  - [x] On custom function
  - [x] On image
  - [x] On audio
  - [x] On document
  - [x] On video
  - [x] On location
  - [x] On message replied
  - [x] On message forwarded
  - [x] On any group event
  - [x] On new chat participant
  - [x] On left chat participant
  - [x] On new chat title
  - [x] On new chat photo
  - [x] On delete chat photo
  - [x] On group chat created

- [x] Action functions
  - [x] Get me
  - [x] Send message
    - [x] easy with keyboard
    - [x] easy with force reply
    - [x] easy with keyboard hide
  - [x] Forward message
  - [x] getUpdates
    - [x] This is done automatically when you use the `SimpleStart()` or `Start()`, you shouldn't touch this ;-)
  - [x] setWebhook  (!! in testing !!)
    - [x] This is done automatically when you use `ServerStart()`, you can use other ways.
  - [x] Send photo
   - [x] From id
   - [x] From file
   - [x] easy with keyboard
   - [x] easy with force reply
   - [x] easy with keyboard hide
  - [x] Send audio
   - [x] From id
   - [x] From file
   - [x] easy with keyboard
   - [x] easy with force reply
   - [x] easy with keyboard hide
  - [x] Send document
   - [x] From id
   - [x] From file
   - [x] easy with keyboard
   - [x] easy with force reply
   - [x] easy with keyboard hide
  - [x] Send sticker
   - [x] From id
   - [x] From file
   - [x] easy with keyboard
   - [x] easy with force reply
   - [x] easy with keyboard hide
  - [x] Send video
   - [x] From id
   - [x] From file
   - [x] easy with keyboard
   - [x] easy with force reply
   - [x] easy with keyboard hide
  - [x] Send location
  - [x] Send chat action
  - [x] Get user profile photos

- [ ] Other nice things!
  - [x] Default options for messages configured before start.
    - [x] Disable webpage preview
    - [x] Selective the reply_markup
    - [x] One time keyboard
    - [x] Clean initial @username in message
    - [x] Add slash in message if don't exist and @username had been used
  - [ ] Easy to work with authorized users
  - [x] Easy to work with "flow" messages

- [ ] Complete documentation xD
  - [ ] Audio doc
  - [ ] Document doc
  - [ ] Sticker doc
  - [ ] Video doc
  - [ ] Location doc
  - [ ] ChatAction doc
  - [ ] Awesome chain doc
  - [ ] GetUserProfilePhotos
  - [ ] Webhook
  - [ ] Chain messages
  - [ ] Default options
  - [ ] Call from ReplyFn

- [ ] Tests


[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/rockneurotiko/go-tgbot/trend.png)](https://bitdeli.com/free "Bitdeli Badge")
