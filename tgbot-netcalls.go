package tgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oleiade/reflections"
)

func (bot TgBot) sendGenericQuery(path string, ignore string, file string, payload interface{}) ResultWithMessage {
	url := bot.buildPath(path)
	switch val := payload.(type) {
	//WebHook
	case SetWebhookQuery:
		return bot.genericSendPostData(url, val)
	case SetWebhookCertQuery:
		return bot.sendConvertingFile(url, ignore, file, val)
	// ID
	case SendPhotoIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
	case SendAudioIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
	case SendVoiceIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
	case SendDocumentIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
	case SendStickerIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
	case SendVideoIDQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.genericSendPostData(url, val)
		// Path
	case SendPhotoPathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	case SendAudioPathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	case SendVoicePathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	case SendDocumentPathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	case SendStickerPathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	case SendVideoPathQuery:
		hookPayload(&val, bot.DefaultOptions)
		return bot.sendConvertingFile(url, ignore, file, val)
	default:
		ipath, err := reflections.GetField(val, ignore)
		if err != nil {
			break
		}
		params := convertInterfaceMap(val, []string{ignore})
		return bot.uploadFileWithResult(url, params, file, ipath)
	}
	errc := 400
	errs := "Wrong Query!"
	return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
}

func (bot TgBot) genericSendPostData(url string, payload interface{}) ResultWithMessage {
	// hook the payload :P
	body, error := postPetition(url, payload, nil)
	if error != nil {
		errc := 500
		err := "Some error happened while sending the message"
		return ResultWithMessage{ResultBase{false, &errc, &err}, nil}
	}
	var result ResultWithMessage
	json.Unmarshal([]byte(body), &result)
	return result
}

func (bot TgBot) sendConvertingFile(url string, ignore string, file string, val interface{}) ResultWithMessage {
	ipath, err := reflections.GetField(val, ignore)
	if err != nil {
		errc := 400
		errs := "Wrong Query!"
		return ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	fpath := fmt.Sprintf("%+v", ipath)
	params := convertInterfaceMap(val, []string{ignore})
	return bot.uploadFileWithResult(url, params, file, fpath)
}

func (bot TgBot) uploadFileWithResult(url string, params map[string]string, fieldname string, filename interface{}) ResultWithMessage {
	res, err := bot.uploadFile(url, params, fieldname, filename)
	if err != nil {
		errc := 500
		errs := err.Error()
		res = ResultWithMessage{ResultBase{false, &errc, &errs}, nil}
	}
	return res
}

func (bot TgBot) uploadFileNoResult(url string, params map[string]string, fieldname string, filename interface{}) ([]byte, error) {
	defaultb := make([]byte, 0)

	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer

	switch rfile := filename.(type) {
	case string:
		rfile = filepath.Clean(rfile)
		f, err := os.Open(rfile)
		if err != nil {
			return defaultb, err
		}

		fw, err := w.CreateFormFile(fieldname, rfile)
		if err != nil {
			return defaultb, err
		}

		if _, err = io.Copy(fw, f); err != nil {
			return defaultb, err
		}
	case ReaderSender:
		if fw, err = w.CreateFormFile(fieldname, rfile.Name); err != nil {
			return defaultb, err
		}
		if _, err = io.Copy(fw, rfile.Read); err != nil {
			return defaultb, err
		}
	case *gif.GIF:
		if fw, err = w.CreateFormFile("document", "image.gif"); err != nil {
			return defaultb, err
		}
		if err = gif.EncodeAll(fw, rfile); err != nil {
			return defaultb, err
		}
	case image.Image:
		imageQuality := jpeg.Options{Quality: jpeg.DefaultQuality}
		if fw, err = w.CreateFormFile("photo", "image.jpeg"); err != nil {
			return defaultb, err
		}
		if err = jpeg.Encode(fw, rfile, &imageQuality); err != nil {
			return defaultb, err
		}
	}

	for key, val := range params {
		if fw, err = w.CreateFormField(key); err != nil {
			return defaultb, err
		}

		if _, err = fw.Write([]byte(val)); err != nil {
			return defaultb, err
		}
	}

	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return defaultb, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return defaultb, err
	}

	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return defaultb, err
	}
	return bytes, nil
}

func (bot TgBot) uploadFile(url string, params map[string]string, fieldname string, filename interface{}) (ResultWithMessage, error) {
	bytes, err := bot.uploadFileNoResult(url, params, fieldname, filename)
	if err != nil {
		return ResultWithMessage{}, err
	}
	var apiResp ResultWithMessage
	json.Unmarshal(bytes, &apiResp)

	return apiResp, nil
}

// func (bot TgBot) uploadFile(url string, params map[string]string, fieldname string, filename interface{}) (ResultWithMessage, error) {
// 	var b bytes.Buffer
// 	var err error
// 	w := multipart.NewWriter(&b)
// 	var fw io.Writer

// 	switch rfile := filename.(type) {
// 	case string:
// 		rfile = filepath.Clean(rfile)
// 		f, err := os.Open(rfile)
// 		if err != nil {
// 			return ResultWithMessage{}, err
// 		}

// 		fw, err := w.CreateFormFile(fieldname, rfile)
// 		if err != nil {
// 			return ResultWithMessage{}, err
// 		}

// 		if _, err = io.Copy(fw, f); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 	case *gif.GIF:
// 		if fw, err = w.CreateFormFile("document", "image.gif"); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 		if err = gif.EncodeAll(fw, rfile); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 	case image.Image:
// 		imageQuality := jpeg.Options{Quality: jpeg.DefaultQuality}
// 		if fw, err = w.CreateFormFile("photo", "image.jpeg"); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 		if err = jpeg.Encode(fw, rfile, &imageQuality); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 	}

// 	for key, val := range params {
// 		if fw, err = w.CreateFormField(key); err != nil {
// 			return ResultWithMessage{}, err
// 		}

// 		if _, err = fw.Write([]byte(val)); err != nil {
// 			return ResultWithMessage{}, err
// 		}
// 	}

// 	w.Close()

// 	req, err := http.NewRequest("POST", url, &b)
// 	if err != nil {
// 		return ResultWithMessage{}, err
// 	}

// 	req.Header.Set("Content-Type", w.FormDataContentType())

// 	client := &http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return ResultWithMessage{}, err
// 	}

// 	bytes, err := ioutil.ReadAll(res.Body)

// 	if err != nil {
// 		return ResultWithMessage{}, err
// 	}

// 	var apiResp ResultWithMessage
// 	json.Unmarshal(bytes, &apiResp)

// 	return apiResp, nil
// }
