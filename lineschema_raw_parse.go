package lineschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/suifengpiao14/kvstruct"
)

const (
	TOKEN_BEGIN = ','
	TOKEN_END   = '='
	EOF         = "\n"
)

// ParseLineschema 解析lineschema
func ParseLineschema(lineschemaRaw string) (jsonline *Lineschema, err error) {
	lineschemaRaw = compress(lineschemaRaw)
	lines := strings.Split(lineschemaRaw, EOF)
	jsonline = &Lineschema{
		Items: make([]*LineschemaItem, 0),
	}
	for _, line := range lines {
		kvs := parserOneLine(line)
		if IsMetaLine(kvs) {
			meta, err := kvs2meta(kvs)
			if err != nil {
				return nil, err
			}
			jsonline.Meta = meta
			continue
		}
		item, err := kv2item(kvs)
		if err != nil {
			return nil, err
		}
		// err = validItem(item)
		// if err != nil {
		// 	err = errors.WithMessage(err, fmt.Sprintf(" got:%s", line))
		// 	return nil, err
		// }
		item.Lineschema = jsonline
		jsonline.Items = append(jsonline.Items, item)
	}

	return jsonline, nil
}

func kvs2meta(kvs kvstruct.KVS) (meta *Meta, err error) {
	meta = new(Meta)
	jb, err := json.Marshal(kvs.Map())
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jb, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func IsMetaLine(lineTags kvstruct.KVS) bool {
	hasFullname, hasId := false, false
	for _, kvPair := range lineTags {
		switch kvPair.Key {
		case "id":
			hasId = true
		case "fullname":
			hasFullname = true
		}
	}
	is := hasId && !hasFullname
	return is
}

//	func validItem(item *LineschemaItem) (err error) {
//		if item.Fullname == "" {
//			err = errors.New("fullname required ")
//			return err
//		}
//		return nil
//	}
func kv2item(kvs kvstruct.KVS) (item *LineschemaItem, err error) {
	item = new(LineschemaItem)
	m := make(map[string]interface{})
	for k, v := range kvs.Map() {
		m[k] = v
	}

	jb, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jb, item)
	if err != nil {
		return nil, err
	}
	item.InitPath()
	return item, nil
}

func compress(lineschema string) (compressedSchema string) {
	lineschema = strings.TrimSpace(lineschema)
	replacer := strings.NewReplacer(" ", "", "\t", "", "\r", "")
	compressedSchema = replacer.Replace(lineschema)
	return compressedSchema
}

// parserOneLine 解析一行数据
func parserOneLine(line string) (kvs kvstruct.KVS) {
	line = compress(line)
	if line == "" {
		return nil
	}
	ret := make([]string, 0)
	separated := strings.Split(line, ",")
	ret = append(ret, separated[0])
	i := 0
	for _, nextTag := range separated[1:] {
		if isToken(nextTag) {
			ret = append(ret, nextTag)
			i++
		} else {
			ret[i] = fmt.Sprintf("%s,%s", ret[i], nextTag)
		}
	}
	kvs = make(kvstruct.KVS, 0)
	for _, pair := range ret {
		arr := strings.SplitN(pair, "=", 2)
		if len(arr) == 1 {
			arr = append(arr, "true")
		}
		k, v := arr[0], arr[1]
		kv := kvstruct.KV{
			Key:   k,
			Value: v,
		}
		kvs.Add(kv)
	}
	// 增加默认type=string，如果存在则忽略
	kvs.AddIgnore(kvstruct.KV{
		Key:   "type",
		Value: "string",
	})
	return kvs
}
func isToken(s string) (yes bool) {
	for _, token := range getTokens() {
		yes = strings.HasPrefix(s, token)
		if yes {
			return yes
		}
	}
	return false
}

func getTokens() (tokens []string) {
	tokens = make([]string, 0)
	meta := new(Meta)
	var rt reflect.Type
	rt = reflect.TypeOf(meta).Elem()
	tokens = append(tokens, getJsonTagname(rt)...)
	item := new(LineschemaItem)
	rt = reflect.TypeOf(item).Elem()
	tokens = append(tokens, getJsonTagname(rt)...)

	return tokens
}

func getJsonTagname(rt reflect.Type) (jsonNames []string) {
	jsonNames = make([]string, 0)
	for i := 0; i < rt.NumField(); i++ {
		jsonTag := rt.Field(i).Tag.Get("json")
		index := strings.Index(jsonTag, ",")
		if index > 0 {
			jsonTag = jsonTag[:index]
		}
		jsonTag = strings.TrimSpace(jsonTag)
		if jsonTag != "-" {
			jsonNames = append(jsonNames, jsonTag)
		}
	}
	return jsonNames
}
