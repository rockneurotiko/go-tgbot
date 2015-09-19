package tgbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"io"
	"strings"

	"github.com/rockneurotiko/gorequest"
)

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
		errc := 403
		desc := ""
		if data.ErrorCode != nil {
			errc = *data.ErrorCode
		}
		if data.Description != nil {
			desc = *data.Description
		}

		errormsg := fmt.Sprintf("Some error happened, maybe your token is bad:\nError code: %d\nDescription: %s\nToken: %s", errc, desc, bot.Token)
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
		return []MessageWithUpdateID{}, errors.New("Some error happened in your petition, check your token or remove the webhook.")
	}
	return data.Result, nil
}

// SetWebhook call the setWebhook API method with the URL suplied, will return the result or an error (the error will be sended if the webhook can't be setted)
func (bot TgBot) SetWebhook(url string) (ResultSetWebhook, error) {
	// pet := SetWebhookQuery{&url}
	// req := bot.SetWebhookQuery(pet)
	req := bot.SetWebhookQuery(&url, nil)
	if !req.Ok {
		return req, errors.New(req.Description)
	}
	return req, nil
}

// SetWebhookWithCert ...
func (bot TgBot) SetWebhookWithCert(url string, cert string) ResultSetWebhook {
	if !looksLikePath(cert) {
		r := false
		errc := 404
		return ResultSetWebhook{false, "Your certificate is not a valid path", &r, &errc}
	}

	urlsend := bot.buildPath("setWebhook")

	bytes, err := bot.uploadFileNoResult(urlsend, map[string]string{"url": url}, "certificate", cert)
	if err != nil {
		errc := 500
		return ResultSetWebhook{false, "Some error happened", nil, &errc}
	}
	var apiResp ResultSetWebhook
	json.Unmarshal(bytes, &apiResp)

	return apiResp
}

// SetWebhookNoQuery ...
func (bot TgBot) SetWebhookNoQuery(urlw string) ResultSetWebhook {
	q := SetWebhookQuery{&urlw}
	url := bot.buildPath("setWebhook")
	body, error := postPetition(url, q, nil)

	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultSetWebhook{false, err, nil, &errc}
	}
	var result ResultSetWebhook
	json.Unmarshal([]byte(body), &result)
	return result
}

// SetWebhookQuery raw method that uses the struct to send the petition.
func (bot TgBot) SetWebhookQuery(url *string, cert *string) ResultSetWebhook {
	if cert == nil {
		return bot.SetWebhookNoQuery(*url)
	}
	return bot.SetWebhookWithCert(*url, *cert)
}

// GetUserProfilePhotos args will use only the two first parameters, the first one will be the limit of images to get, and the second will be the offset photo id.
func (bot TgBot) GetUserProfilePhotos(uid int, args ...int) UserProfilePhotos {
	pet := ResultWithUserProfilePhotos{}
	getq := GetUserProfilePhotosQuery{uid, nil, nil}
	if len(args) == 1 {
		v1 := args[0]
		getq = GetUserProfilePhotosQuery{uid, nil, &v1}
	} else if len(args) >= 2 {
		v1 := args[0]
		v2 := args[1]
		getq = GetUserProfilePhotosQuery{uid, &v2, &v1}
	}

	pet = bot.GetUserProfilePhotosQuery(getq)

	if !pet.Ok || pet.Result == nil {
		return UserProfilePhotos{}
	}
	return *pet.Result
}

// Send messages

// SimpleSendMessage send a simple text message.
func (bot TgBot) SimpleSendMessage(msg Message, text string) (res Message, err error) {
	ressm := bot.SendMessage(msg.Chat.ID, text, nil, nil, nil, nil)
	return splitResultInMessageError(ressm)
}

// SendMessageWithKeyboard send a message with explicit Keyboard
func (bot TgBot) SendMessageWithKeyboard(cid int, text string, parsemode *ParseModeT, dwp *bool, rtmid *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, parsemode, dwp, rtmid, &rkm)
}

// SendMessageWithForceReply send a message with explicit Force Reply.
func (bot TgBot) SendMessageWithForceReply(cid int, text string, parsemode *ParseModeT, dwp *bool, rtmid *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, parsemode, dwp, rtmid, &rkm)
}

// SendMessageWithKeyboardHide send a message with explicit Keyboard Hide.
func (bot TgBot) SendMessageWithKeyboardHide(cid int, text string, parsemode *ParseModeT, dwp *bool, rtmid *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendMessage(cid, text, parsemode, dwp, rtmid, &rkm)
}

// SendMessage full function wrapper for sendMessage, uses the markup interface
func (bot TgBot) SendMessage(cid int, text string, parsemode *ParseModeT, dwp *bool, rtmid *int, rm *ReplyMarkupInt) ResultWithMessage {
	var pm *string = nil
	if parsemode != nil {
		pmt := parsemode.String()
		pm = &pmt
	}
	payload := QuerySendMessage{cid, text, pm, dwp, rtmid, rm}
	return bot.SendMessageQuery(payload)
}

// SendMessageQuery full sendMessage with the query.
func (bot TgBot) SendMessageQuery(payload QuerySendMessage) ResultWithMessage {
	url := bot.buildPath("sendMessage")
	hookPayload(&payload, bot.DefaultOptions)
	return bot.genericSendPostData(url, payload)
}

// Forward Message!!

// ForwardMessage full function wrapper for forwardMessage
func (bot TgBot) ForwardMessage(cid int, fid int, mid int) ResultWithMessage {
	payload := ForwardMessageQuery{cid, fid, mid}
	return bot.ForwardMessageQuery(payload)
}

// ForwardMessageQuery  full forwardMessage call
func (bot TgBot) ForwardMessageQuery(payload ForwardMessageQuery) ResultWithMessage {
	url := bot.buildPath("forwardMessage")
	hookPayload(&payload, bot.DefaultOptions)
	return bot.genericSendPostData(url, payload)
}

// Send photo!!

// SimpleSendPhoto send just a photo.
func (bot TgBot) SimpleSendPhoto(msg Message, photo interface{}) (res Message, err error) {
	cid := msg.Chat.ID
	ressm := bot.SendPhoto(cid, photo, nil, nil, nil)
	return splitResultInMessageError(ressm)
}

// SendPhotoWithKeyboard send a photo with explicit Keyboard
func (bot TgBot) SendPhotoWithKeyboard(cid int, photo interface{}, caption *string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithForceReply send a photo with explicit Force Reply.
func (bot TgBot) SendPhotoWithForceReply(cid int, photo interface{}, caption *string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhotoWithKeyboardHide send a photo with explicit Keyboard Hide.
func (bot TgBot) SendPhotoWithKeyboardHide(cid int, photo interface{}, caption *string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendPhoto(cid, photo, caption, rmi, &rkm)
}

// SendPhoto full function wrapper for sendPhoto, use the markup interface.
func (bot TgBot) SendPhoto(cid int, photo interface{}, caption *string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload, err := bot.imageInterfaceToType(cid, photo, caption, rmi, rm)
	if err != nil {
		errc := 500
		errs := err.Error()
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return bot.SendPhotoQuery(payload)
}

func (bot TgBot) imageInterfaceToType(cid int, photo interface{}, caption *string, rmi *int, rm *ReplyMarkupInt) (payload interface{}, err error) {
	switch pars := photo.(type) {
	case string:
		payload = SendPhotoIDQuery{cid, pars, caption, rmi, rm}
		if looksLikePath(pars) {
			payload = SendPhotoPathQuery{cid, pars, caption, rmi, rm}
		}
	case image.Image:
		mp := struct {
			ChatID           int             `json:"chat_id"`
			Photo            image.Image     `json:"photo"`
			Caption          *string         `json:"caption,omitempty"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, caption, rmi, rm}
		hookPayload(&mp, bot.DefaultOptions)
		payload = mp
	default:
		err = errors.New("No struct interface detected")
	}
	return
}

// SendPhotoQuery full function that uses the query.
func (bot TgBot) SendPhotoQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendPhoto", "Photo", "photo", payload)
}

// Audio!!

// SimpleSendAudio send just an audio
func (bot TgBot) SimpleSendAudio(msg Message, audio string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendAudioIDQuery{cid, audio, nil, nil, nil, nil, nil}
	if looksLikePath(audio) {
		payload = SendAudioPathQuery{cid, audio, nil, nil, nil, nil, nil}
	}
	ressm := bot.SendAudioQuery(payload)
	return splitResultInMessageError(ressm)
}

// SendAudioWithKeyboard send a audio with explicit Keyboard
func (bot TgBot) SendAudioWithKeyboard(cid int, audio string, duration *int, performer *string, title *string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, audio, duration, performer, title, rmi, &rkm)
}

// SendAudioWithForceReply send a audio with explicit Force Reply.
func (bot TgBot) SendAudioWithForceReply(cid int, audio string, duration *int, performer *string, title *string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, audio, duration, performer, title, rmi, &rkm)
}

// SendAudioWithKeyboardHide send a audio with explicit Keyboard Hide.
func (bot TgBot) SendAudioWithKeyboardHide(cid int, audio string, duration *int, performer *string, title *string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendAudio(cid, audio, duration, performer, title, rmi, &rkm)
}

// SendAudio full function to send an audio. Uses the reply markup interface.
func (bot TgBot) SendAudio(cid int, audio string, duration *int, performer *string, title *string, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendAudioIDQuery{cid, audio, duration, performer, title, rmi, rm}
	if looksLikePath(audio) {
		payload = SendAudioPathQuery{cid, audio, duration, performer, title, rmi, rm}
	}
	return bot.SendAudioQuery(payload)
}

// SendAudioQuery full function using the query.
func (bot TgBot) SendAudioQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendAudio", "Audio", "audio", payload)
}

// Voice!!

// SimpleSendVoice send just an audio
func (bot TgBot) SimpleSendVoice(msg Message, audio string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendVoiceIDQuery{cid, audio, nil, nil, nil}
	if looksLikePath(audio) {
		payload = SendVoicePathQuery{cid, audio, nil, nil, nil}
	}
	ressm := bot.SendVoiceQuery(payload)
	return splitResultInMessageError(ressm)
}

// SendVoiceWithKeyboard send a audio with explicit Keyboard
func (bot TgBot) SendVoiceWithKeyboard(cid int, audio string, duration *int, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVoice(cid, audio, duration, rmi, &rkm)
}

// SendVoiceWithForceReply send a audio with explicit Force Reply.
func (bot TgBot) SendVoiceWithForceReply(cid int, audio string, duration *int, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVoice(cid, audio, duration, rmi, &rkm)
}

// SendVoiceWithKeyboardHide send a audio with explicit Keyboard Hide.
func (bot TgBot) SendVoiceWithKeyboardHide(cid int, audio string, duration *int, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVoice(cid, audio, duration, rmi, &rkm)
}

// SendVoice full function to send an audio. Uses the reply markup interface.
func (bot TgBot) SendVoice(cid int, audio string, duration *int, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendVoiceIDQuery{cid, audio, duration, rmi, rm}
	if looksLikePath(audio) {
		payload = SendVoicePathQuery{cid, audio, nil, rmi, rm}
	}
	return bot.SendVoiceQuery(payload)
}

// SendVoiceQuery full function using the query.
func (bot TgBot) SendVoiceQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendVoice", "Voice", "voice", payload)
}

//Documents!!

// SimpleSendDocument send just a document.
func (bot TgBot) SimpleSendDocument(msg Message, document string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendDocumentIDQuery{cid, document, nil, nil}
	if looksLikePath(document) {
		payload = SendDocumentPathQuery{cid, document, nil, nil}
	}
	ressm := bot.SendDocumentQuery(payload)
	return splitResultInMessageError(ressm)
}

// SendDocumentWithKeyboard send a document with explicit keyboard.
func (bot TgBot) SendDocumentWithKeyboard(cid int, document string, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, document, rmi, &rkm)
}

// SendDocumentWithForceReply send a document with explicit force reply
func (bot TgBot) SendDocumentWithForceReply(cid int, document string, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, document, rmi, &rkm)
}

// SendDocumentWithKeyboardHide send a document with explicit keyboard hide.
func (bot TgBot) SendDocumentWithKeyboardHide(cid int, document string, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendDocument(cid, document, rmi, &rkm)
}

// SendDocument full function to send document, uses the reply markup interface.
func (bot TgBot) SendDocument(cid int, document interface{}, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload, err := bot.documentInterfaceToType(cid, document, rmi, rm)
	if err != nil {
		errc := 500
		errs := err.Error()
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	// var payload interface{} = SendDocumentIDQuery{cid, document, rmi, rm}
	// if looksLikePath(document) {
	// 	payload = SendDocumentPathQuery{cid, document, rmi, rm}
	// }
	return bot.SendDocumentQuery(payload)
}

// func (bot TgBot) SendDocumentImageTest(cid int, payload interface{}) ResultWithMessage {
// 	payload, err := bot.documentInterfaceToType(cid, payload, nil, nil)
// 	if err != nil {
// 		errc := 500
// 		errs := err.Error()
// 		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
// 	}
// 	return bot.SendDocumentQuery(payload)
// }

type ReaderSender struct {
	Read io.Reader
	Name string
}

func (bot TgBot) documentInterfaceToType(cid int, photo interface{}, rmi *int, rm *ReplyMarkupInt) (payload interface{}, err error) {
	switch pars := photo.(type) {
	case string:
		payload = SendDocumentIDQuery{cid, pars, rmi, rm}
		if looksLikePath(pars) {
			payload = SendDocumentPathQuery{cid, pars, rmi, rm}
		}
	case ReaderSender:
		mp := struct {
			ChatID           int             `json:"chat_id"`
			Document         ReaderSender    `json:"document"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, rmi, rm}
		hookPayload(&mp, bot.DefaultOptions)
		payload = mp
	case image.Image:
		{
			mp := struct {
				ChatID           int             `json:"chat_id"`
				Document         image.Image     `json:"document"`
				ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
				ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
			}{cid, pars, rmi, rm}
			hookPayload(&mp, bot.DefaultOptions)
			payload = mp
		}
	case *gif.GIF:
		mp := struct {
			ChatID           int             `json:"chat_id"`
			Document         *gif.GIF        `json:"document"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, rmi, rm}
		hookPayload(&mp, bot.DefaultOptions)
		payload = mp
	default:
		err = errors.New("No struct interface detected")
	}
	return
}

// SendDocumentQuery full function using the query.
func (bot TgBot) SendDocumentQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendDocument", "Document", "document", payload)
}

// Stickers!!!

// SimpleSendSticker just send a sticker!!
func (bot TgBot) SimpleSendSticker(msg Message, sticker interface{}) (res Message, err error) {
	cid := msg.Chat.ID
	ressm := bot.SendSticker(cid, sticker, nil, nil)
	return splitResultInMessageError(ressm)
}

// SendStickerWithKeyboard send a sticker with explicit keyboard.
func (bot TgBot) SendStickerWithKeyboard(cid int, sticker interface{}, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, sticker, rmi, &rkm)
}

// SendStickerWithForceReply send a sticker with explicit force reply.
func (bot TgBot) SendStickerWithForceReply(cid int, sticker interface{}, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, sticker, rmi, &rkm)
}

// SendStickerWithKeyboardHide send a sticker with explicit keyboad hide.
func (bot TgBot) SendStickerWithKeyboardHide(cid int, sticker interface{}, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendSticker(cid, sticker, rmi, &rkm)
}

// SendSticker full function to send a sticker, uses reply markup interface.
func (bot TgBot) SendSticker(cid int, sticker interface{}, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload, err := bot.stickerInterfaceToType(cid, sticker, rmi, rm)
	if err != nil {
		errc := 500
		errs := err.Error()
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return bot.SendStickerQuery(payload)
}

func (bot TgBot) stickerInterfaceToType(cid int, sticker interface{}, rmi *int, rm *ReplyMarkupInt) (payload interface{}, err error) {
	switch pars := sticker.(type) {
	case string:
		payload = SendStickerIDQuery{cid, pars, rmi, rm}
		if looksLikePath(pars) {
			payload = SendStickerPathQuery{cid, pars, rmi, rm}
		}
	case image.Image:
		payload = struct {
			ChatID           int             `json:"chat_id"`
			Photo            image.Image     `json:"photo"`
			ReplyToMessageID *int            `json:"reply_to_message_id,omitempty"`
			ReplyMarkup      *ReplyMarkupInt `json:"reply_markup,omitempty"`
		}{cid, pars, rmi, rm}
	default:
		err = errors.New("No struct interface detected")
	}
	return
}

// SendStickerQuery full function to send an sticker, uses the query.
func (bot TgBot) SendStickerQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendSticker", "Sticker", "sticker", payload)
}

// Send video!!!!

// SimpleSendVideo just send a video from file path or id
func (bot TgBot) SimpleSendVideo(msg Message, photo string) (res Message, err error) {
	cid := msg.Chat.ID
	var payload interface{} = SendVideoIDQuery{cid, photo, nil, nil, nil, nil}
	if looksLikePath(photo) {
		payload = SendVideoPathQuery{cid, photo, nil, nil, nil, nil}
	}
	ressm := bot.SendVideoQuery(payload)
	return splitResultInMessageError(ressm)
}

// SendVideoWithKeyboard send a video with explicit keyboard.
func (bot TgBot) SendVideoWithKeyboard(cid int, photo string, caption *string, duration *int, rmi *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, caption, duration, rmi, &rkm)
}

// SendVideoWithForceReply send a video with explicit force reply.
func (bot TgBot) SendVideoWithForceReply(cid int, photo string, caption *string, duration *int, rmi *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, caption, duration, rmi, &rkm)
}

// SendVideoWithKeyboardHide send a video with explici keyboard hide.
func (bot TgBot) SendVideoWithKeyboardHide(cid int, photo string, caption *string, duration *int, rmi *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendVideo(cid, photo, caption, duration, rmi, &rkm)
}

// SendVideo full function to send a video.
func (bot TgBot) SendVideo(cid int, photo string, caption *string, duration *int, rmi *int, rm *ReplyMarkupInt) ResultWithMessage {
	var payload interface{} = SendVideoIDQuery{cid, photo, duration, caption, rmi, rm}
	if looksLikePath(photo) {
		payload = SendVideoPathQuery{cid, photo, duration, caption, rmi, rm}
	}
	return bot.SendVideoQuery(payload)
}

// SendVideoQuery full function to send video with query.
func (bot TgBot) SendVideoQuery(payload interface{}) ResultWithMessage {
	return bot.sendGenericQuery("sendVideo", "Video", "video", payload)
}

// send Location!!!

// SimpleSendLocation just send a location.
func (bot TgBot) SimpleSendLocation(msg Message, latitude float64, longitude float64) (res Message, err error) {
	ressm := bot.SendLocation(msg.Chat.ID, latitude, longitude, nil, nil)
	return splitResultInMessageError(ressm)
}

// SendLocationWithKeyboard send a location with explicit keyboard.
func (bot TgBot) SendLocationWithKeyboard(cid int, latitude float64, longitude float64, rtmid *int, rm ReplyKeyboardMarkup) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocationWithForceReply send a location with explicit force reply.
func (bot TgBot) SendLocationWithForceReply(cid int, latitude float64, longitude float64, rtmid *int, rm ForceReply) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocationWithKeyboardHide send a location with explicit keyboard hide.
func (bot TgBot) SendLocationWithKeyboardHide(cid int, latitude float64, longitude float64, rtmid *int, rm ReplyKeyboardHide) ResultWithMessage {
	var rkm ReplyMarkupInt = rm
	return bot.SendLocation(cid, latitude, longitude, rtmid, &rkm)
}

// SendLocation full function wrapper for sendLocation
func (bot TgBot) SendLocation(cid int, latitude float64, longitude float64, rtmid *int, rm *ReplyMarkupInt) ResultWithMessage {
	payload := SendLocationQuery{cid, latitude, longitude, rtmid, rm}
	return bot.SendLocationQuery(payload)
}

// SendLocationQuery full sendLocation call with query.
func (bot TgBot) SendLocationQuery(payload SendLocationQuery) ResultWithMessage {
	url := bot.buildPath("sendLocation")
	hookPayload(&payload, bot.DefaultOptions)
	return bot.genericSendPostData(url, payload)
}

// Send chat action!!!

// SimpleSendChatAction just send an action answering a message.
func (bot TgBot) SimpleSendChatAction(msg Message, ca ChatAction) {
	bot.SendChatAction(msg.Chat.ID, ca)
}

// SendChatAction send an action to an id.
func (bot TgBot) SendChatAction(cid int, ca ChatAction) {
	bot.SendChatActionQuery(SendChatActionQuery{cid, ca.String()})
}

// SendChatActionQuery send an action query.
func (bot TgBot) SendChatActionQuery(payload SendChatActionQuery) {
	url := bot.buildPath("sendChatAction")
	hookPayload(&payload, bot.DefaultOptions)
	bot.genericSendPostData(url, payload)
}

// GetUserProfilePhotosQuery raw method that uses the struct to send the petition.
func (bot TgBot) GetUserProfilePhotosQuery(quer GetUserProfilePhotosQuery) ResultWithUserProfilePhotos {
	url := bot.buildPath("getUserProfilePhotos")
	body, error := postPetition(url, quer, nil)

	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultWithUserProfilePhotos{ResultBase{false, &errc, &err}, nil}
	}
	var result ResultWithUserProfilePhotos
	json.Unmarshal([]byte(body), &result)
	return result
}

func (bot TgBot) GetFile(id string) ResultWithGetFile {
	url := bot.buildPath("getFile")
	body, error := postPetition(url, struct {
		ID string `json:"file_id"`
	}{id}, nil)

	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultWithGetFile{ResultBase{false, &errc, &err}, nil}
	}
	var result ResultWithGetFile
	json.Unmarshal([]byte(body), &result)
	return result
}

func (bot TgBot) DownloadFilePathReader(path string) (io.ReadCloser, error) {
	url := bot.buildFilePath(path)
	resp, _, errq := gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Get(url).
		End()
	if errq != nil {
		if len(errq) > 0 {
			return nil, errq[0]
		} else {
			return nil, errors.New("Error in GET petition")
		}
	}

	return resp.Body, nil
}
