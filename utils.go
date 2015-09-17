package tgbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gorelic"
	"github.com/oleiade/reflections"
	"github.com/rockneurotiko/gorequest"
)

func convertToCommand(reg string) string {
	if !strings.HasSuffix(reg, "$") {
		reg = reg + "$"
	}
	if !strings.HasPrefix(reg, "^/") {
		if strings.HasPrefix(reg, "/") {
			reg = "^" + reg
		} else {
			reg = "^/" + reg
		}
	}
	return reg
}

func findStringSubmatchMap(r *regexp.Regexp, s string) map[string]string {
	captures := make(map[string]string)
	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}
	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}
	return captures
}

func looksLikePath(p string) bool {
	p = filepath.Clean(p)
	if len(strings.Split(p, ".")) > 1 {
		// The IDS don't have dots :P
		// But let's check if exist, anyway
		_, err := os.Stat(p)
		return err == nil
	}
	return false
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

func isInList(v string, l []string) bool {
	sort.Strings(l)
	i := sort.SearchStrings(l, v)
	return i < len(l) && l[i] == v
}

func convertInterfaceMap(p interface{}, except []string) map[string]string {
	nint := map[string]string{}
	var structItems map[string]interface{}

	structItems, _ = reflections.Items(p)
	for v, v2 := range structItems {
		if isZeroOfUnderlyingType(v2) || isInList(v, except) {
			continue
		}
		v = strings.ToLower(strings.Join(camelcase.Split(v), "_"))
		switch val := v2.(type) {
		case interface{}:
			sv, _ := json.Marshal(val)
			nint[v] = string(sv)
		default:
			nint[v] = fmt.Sprintf("%+v", v2)
		}
	}
	return nint
}

func StartServerMultiplesBotsHostPort(uri string, pathl string, host string, port string, newrelic *RelicConfig, bots ...*TgBot) {
	var puri *url.URL
	if uri != "" {
		tmpuri, err := url.Parse(uri)
		if err != nil {
			fmt.Printf("Bad URL %s", uri)
			return
		}
		puri = tmpuri
	}

	botsmap := make(map[string]*TgBot)
	for _, bot := range bots {
		tokendiv := strings.Split(bot.Token, ":")
		if len(tokendiv) != 2 {
			return
		}

		tokenpath := fmt.Sprintf("%s%s", tokendiv[0], tokendiv[1])
		botpathl := path.Join(pathl, tokenpath)

		nuri, _ := puri.Parse(botpathl)
		remoteuri := nuri.String()
		res, error := bot.SetWebhook(remoteuri)

		if error != nil {
			ec := res.ErrorCode
			fmt.Printf("Error setting the webhook: \nError code: %d\nDescription: %s\n", &ec, res.Description)
			continue
		}
		if bot.MainListener == nil {
			bot.StartMainListener()
		}
		botsmap[tokenpath] = bot
	}

	pathtolisten := path.Join(pathl, "(?P<token>[a-zA-Z0-9-_]+)")

	m := martini.Classic()
	m.Post(pathtolisten, binding.Json(MessageWithUpdateID{}), func(params martini.Params, msg MessageWithUpdateID) {
		bot, ok := botsmap[params["token"]]

		if ok && msg.UpdateID > 0 && msg.Msg.ID > 0 {
			bot.MainListener <- msg
		} else {
			fmt.Println("Someone tried with: ", params["token"], msg)
		}
	})

	if newrelic != nil {
		gorelic.InitNewrelicAgent(newrelic.Token, newrelic.Name, false)
		m.Use(gorelic.Handler)
	}

	if host == "" || port == "" {
		m.Run()
	} else {
		m.RunOnAddr(host + ":" + port)
	}
}

// StartServerMultiplesBots ...
func StartServerMultiplesBots(uri string, pathl string, newrelic *RelicConfig, bots ...*TgBot) {
	StartServerMultiplesBotsHostPort(uri, pathl, "", "", newrelic, bots...)
}

func splitResultInMessageError(ressm ResultWithMessage) (res Message, err error) {
	if ressm.Ok && ressm.Result != nil {
		res = *ressm.Result
		err = nil
	} else {
		res = Message{}
		err = fmt.Errorf("Error in petition.\nError code: %d\nDescription: %s", *ressm.ErrorCode, *ressm.Description)
	}
	return
}

func postPetition(url string, payload interface{}, ctype *string) (string, error) {
	request := gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Post(url).
		Send(payload)
	request.TargetType = "form"

	if ctype != nil {
		request.Set("Content-Type", *ctype)
	}

	_, body, err := request.End()

	if err != nil {
		return "", errors.New("Some error happened")
	}
	return body, nil
}

func getPetition(url string, queries []string) (string, error) {
	req := gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Get(url)

	for _, q := range queries {
		req.Query(q)
	}
	_, body, errq := req.End()
	if errq != nil {
		return "", errors.New("There were some error trying to do the petition")
	}
	return body, nil
}
