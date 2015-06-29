package tgbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/oleiade/reflections"
	"github.com/parnurzeal/gorequest"
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

// FindStringSubmatchMap ...
func FindStringSubmatchMap(r *regexp.Regexp, s string) map[string]string {
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

// LooksLikePath ...
func LooksLikePath(p string) bool {
	p = filepath.Clean(p)
	if len(strings.Split(p, ".")) > 1 {
		// The IDS don't have dots :P
		// But let's check if exist, anyway
		_, err := os.Stat(p)
		return err == nil
	}
	return false
}

// IsZeroOfUnderlyingType ...
func IsZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

// IsInList ...
func IsInList(v string, l []string) bool {
	sort.Strings(l)
	i := sort.SearchStrings(l, v)
	return i < len(l) && l[i] == v
}

// ConvertInterfaceMap ...
func ConvertInterfaceMap(p interface{}, except []string) map[string]string {
	nint := map[string]string{}
	var structItems map[string]interface{}

	structItems, _ = reflections.Items(p)
	for v, v2 := range structItems {
		if IsZeroOfUnderlyingType(v2) || IsInList(v, except) {
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

// postPetition ...
func postPetition(url string, payload interface{}, ctype *string) (string, error) {
	request := gorequest.New().Post(url).
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

// getPetition ...
func getPetition(url string, queries []string) (string, error) {
	req := gorequest.New().Get(url)

	for _, q := range queries {
		req.Query(q)
	}
	_, body, errq := req.End()
	if errq != nil {
		return "", errors.New("There were some error trying to do the petition")
	}
	return body, nil
}
