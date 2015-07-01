package main

import (
	"bytes"
	"fmt"
	"image"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	hexapic "github.com/blan4/hexapic/core"
	godotenv "github.com/joho/godotenv"
	"github.com/rockneurotiko/go-tgbot"
)

var instagramid = ""

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
	"/sendimage":      "Sends you an image",
	"/sendimagekey":   "Sends you an image with a custom keyboard",
	"/senddocument":   "Sends you a document",
	"/sendsticker":    "Sends you a sticker",
	"/sendvideo":      "Sends you a video",
	"/sendlocation":   "Sends you a location",
	"/sendchataction": "Sends a random chat action",
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
	bot.Answer(msg).Text("Hidden it!").KeyboardHide(rkm).End()
	// bot.SendMessageWithKeyboardHide(msg.Chat.ID, "Hiden it!", nil, nil, rkm)
	return nil
}

func cmdKeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{"I", "<3"}, {"You"}}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: false,
		Selective:       false}
	bot.Answer(msg).Text("Enjoy the keyboard").Keyboard(rkm).End()
	// bot.SendMessageWithKeyboard(msg.Chat.ID, "Enjoy the keyboard", nil, nil, rkm)
	return nil
}

func hardEcho(bot tgbot.TgBot, msg tgbot.Message, vals []string, kvals map[string]string) *string {
	msgtext := ""
	if len(vals) > 1 {
		msgtext = vals[1]
	}
	rkm := tgbot.ForceReply{Force: true, Selective: false}
	bot.Answer(msg).Text(msgtext).ForceReply(rkm).End()
	// bot.SendMessageWithForceReply(msg.Chat.ID, msgtext, nil, nil, rkm)
	return nil
}

func forwardHand(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Forward(msg.Chat.ID, msg.ID).End()
	// bot.ForwardMessage(msg.Chat.ID, msg.Chat.ID, msg.ID)
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
	bot.Answer(msg).Text("Starting").End()
	// bot.SimpleSendMessage(msg, "Starting")
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
	bot.Answer(msg).Text("There you have the commands! http://google.com").Keyboard(rkm).End()
	// bot.SendMessageWithKeyboard(msg.Chat.ID, "There you have the commands!", nil, nil, rkm)
	return nil
}

func allMsgHand(bot tgbot.TgBot, msg tgbot.Message) {
	// uncomment this to see it :)
	fmt.Printf("Received message: %+v\n", msg)
	// bot.SimpleSendMessage(msg, "Received message!")
}

func conditionFunc(bot tgbot.TgBot, msg tgbot.Message) bool {
	return msg.Photo != nil
}

func conditionCallFunc(bot tgbot.TgBot, msg tgbot.Message) {
	fmt.Printf("Text: %+v\n", msg.Text)
	// bot.SimpleSendMessage(msg, "Nice image :)")
}

func imageResend(bot tgbot.TgBot, msg tgbot.Message, photos []tgbot.PhotoSize, id string) {
	caption := "I like this photo <3"
	mid := msg.ID
	bot.Answer(msg).Photo(id).Caption(caption).ReplyToMessage(mid).End()
	// bot.SendPhoto(msg.Chat.ID, id, &caption, &mid, nil)
}

func sendImage(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	// bot.SendPhotoQuery(tgbot.SendPhotoPathQuery{msg.Chat.ID, "test.jpg", nil, nil, nil})
	// bot.SendPhoto(msg.Chat.ID, "test.jpg", nil, nil, nil)
	// bot.SimpleSendPhoto(msg, "example/simpleexample/files/test.jpg")
	bot.Answer(msg).Photo("example/simpleexample/files/test.jpg").End()
	return nil
}

func sendImageWithKey(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	keylayout := [][]string{{"I love it"}, {"Nah..."}}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: false,
		Selective:       false}
	bot.Answer(msg).
		Photo("example/simpleexample/files/test.jpg").
		Keyboard(rkm).
		End()
	// bot.SendPhotoWithKeyboard(msg.Chat.ID, "example/simpleexample/files/test.jpg", nil, nil, rkm)
	return nil
}

func sendAudio(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).
		Audio("example/simpleexample/files/test.mp3").
		End()
	// bot.SimpleSendAudio(msg, "example/simpleexample/files/test.mp3")
	return nil
}

func returnAudio(bot tgbot.TgBot, msg tgbot.Message, audio tgbot.Audio, fid string) {
	bot.Answer(msg).Audio(fid).End()
	// bot.SimpleSendAudio(msg, fid)
}

func sendDocument(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	mid := msg.ID
	bot.Answer(msg).
		Document("example/simpleexample/files/PracticalPrincipledFRP.pdf").
		ReplyToMessage(mid).
		End()
	// bot.SendDocument(msg.Chat.ID, "example/simpleexample/files/PracticalPrincipledFRP.pdf", &mid, nil)
	return nil
}

func returnDocument(bot tgbot.TgBot, msg tgbot.Message, document tgbot.Document, fid string) {
	bot.Answer(msg).
		Document(fid).
		End()
	// bot.SimpleSendDocument(msg, fid)
}

func sendSticker(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Sticker("example/simpleexample/files/sticker.webp").End()
	// bot.SimpleSendSticker(msg, "example/simpleexample/files/sticker.webp")
	return nil
}

func returnSticker(bot tgbot.TgBot, msg tgbot.Message, sticker tgbot.Sticker, fid string) {
	mid := msg.ID
	bot.Answer(msg).
		Sticker(fid).
		ReplyToMessage(mid).
		End()
	// bot.SendSticker(msg.Chat.ID, fid, &mid, nil)
}

func sendVideo(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).
		Video("example/simpleexample/files/video.mp4").
		End()
	// bot.SimpleSendVideo(msg, "example/simpleexample/files/video.mp4")
	return nil
}

func returnVideo(bot tgbot.TgBot, msg tgbot.Message, video tgbot.Video, fid string) {
	mid := msg.ID
	bot.Answer(msg).
		Video(fid).
		ReplyToMessage(mid).
		End()
	// bot.SendVideo(msg.Chat.ID, fid, &mid, nil)
}

func sendLocation(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).
		Location(40.324159, -4.21096).
		End() // Just a random location xD
	// bot.SimpleSendLocation(msg, 40.324159, -4.21096) // Just a random location xD
	return nil
}

func returnLocation(bot tgbot.TgBot, msg tgbot.Message, latitude float64, longitude float64) {
	mid := msg.ID
	bot.Answer(msg).
		Location(latitude, longitude).
		ReplyToMessage(mid).
		End()
	// bot.SendLocation(msg.Chat.ID, latitude, longitude, &mid, nil)
}

func sendAction(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	actions := []tgbot.ChatAction{tgbot.Typing, tgbot.UploadPhoto, tgbot.RecordVideo, tgbot.UploadVideo, tgbot.RecordAudio, tgbot.UploadAudio, tgbot.UploadDocument, tgbot.FindLocation}

	bot.SimpleSendChatAction(msg, actions[rand.Intn(8)])
	return nil
}

func instPic(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	httpClient := http.DefaultClient
	hexapicAPI := hexapic.NewSearchApi(instagramid, httpClient)
	hexapicAPI.Count = 4
	var imgs []image.Image

	bot.Answer(msg).
		Action(tgbot.UploadPhoto).
		End()

	imgs = hexapicAPI.SearchByTag("cat")
	img := hexapic.GenerateCollage(imgs, 2, 2)
	keylayout := [][]string{{"cat", "dog"}, {"nya", "chick"}}
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Selective:       false,
	}
	caption := "Guess the image!"
	bot.Answer(msg).
		Photo(img).
		Caption(caption).
		Keyboard(rkm).
		End()
	// bot.SendPhotoWithKeyboard(msg.Chat.ID, img, &caption, nil, rkm)
	return nil
}

func answer(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	mytext := "Not implemented yet"
	return &mytext
}

func justtest(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	return &text
}

func main() {
	godotenv.Load("secrets.env")
	// Add a file secrets.env, with the key like:
	// TELEGRAM_KEY=yourtoken
	token := os.Getenv("TELEGRAM_KEY")
	instagramid = os.Getenv("INSTAGRAM_CLIENT_ID")

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
		CustomFn(conditionFunc, conditionCallFunc).
		SimpleCommandFn(`sendimage`, sendImage).
		SimpleCommandFn(`sendimagekey`, sendImageWithKey).
		ImageFn(imageResend).
		SimpleCommandFn(`sendaudio`, sendAudio).
		AudioFn(returnAudio).
		SimpleCommandFn(`senddocument`, sendDocument).
		DocumentFn(returnDocument).
		SimpleCommandFn(`sendsticker`, sendSticker).
		StickerFn(returnSticker).
		SimpleCommandFn(`sendvideo`, sendVideo).
		VideoFn(returnVideo).
		SimpleCommandFn(`sendlocation`, sendLocation).
		LocationFn(returnLocation).
		SimpleCommandFn(`sendchataction`, sendAction)

	bot.StartChain().
		SimpleCommandFn(`guessimage`, instPic).
		SimpleRegexFn(`^(cat|dog|nya|chick)$`, answer).
		CancelChainCommand(`cancel`, justtest).
		EndChain()

	bot.DefaultDisableWebpagePreview(true)      // Disable all link preview by default
	bot.DefaultOneTimeKeyboard(true)            // Enable one time keyboard by default
	bot.DefaultSelective(true)                  // Use Seletive by default
	bot.DefaultCleanInitialUsername(true)       // By default is true! (This removes initial @username from messages)
	bot.DefaultAllowWithoutSlashInMention(true) // By default is true! (This adds the / in the messages that have @username, this needs DefaultCleanInitialUsername true, for example: @username test becomes /test)

	// temp := bot.GetUserProfilePhotos(bot.ID, 1)
	// fmt.Println(temp)
	// res, _ := bot.SetWebhook()
	// fmt.Println(res)
	bot.SimpleStart()

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
	// bot.ImageFn(imageResend)
	// bot.SimpleCommandFn(`sendimage`, sendImage)
	// bot.SimpleCommandFn(`sendimagekey`, sendImageWithKey)
	// bot.SimpleCommandFn(`sendaudio`, sendAudio)
	// bot.AudioFn(returnAudio)
	// bot.SimpleCommandFn(`senddocument`, sendDocument)
	// bot.DocumentFn(returnDocument)
	// bot.SimpleCommandFn(`sendsticker`, sendSticker)
	// bot.StickerFn(returnSticker)
	// bot.SimpleCommandFn(`sendvideo`, sendVideo)
	// bot.VideoFn(returnVideo)
	// bot.SimpleCommandFn(`sendlocation`, sendLocation)
	// bot.LocationFn(returnLocation)

	// bot.SimpleStart()

}
