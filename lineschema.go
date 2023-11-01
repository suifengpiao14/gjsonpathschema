package lineschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/kvstruct"
)

type LineschemaItem struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	Format      string `json:"format,omitempty"`
	Description string `json:"description,omitempty"`

	Comments         string `json:"comment,omitempty"`                 // section 8.3
	Enum             string `json:"enum,omitempty"`                    // section 6.1.2
	EnumNames        string `json:"enumNames,omitempty"`               // section 6.1.2
	Const            string `json:"const,omitempty"`                   // section 6.1.3
	MultipleOf       int    `json:"multipleOf,omitempty,string"`       // section 6.2.1
	Maximum          int    `json:"maximum,omitempty,string"`          // section 6.2.2
	ExclusiveMaximum bool   `json:"exclusiveMaximum,omitempty,string"` // section 6.2.3
	Minimum          int    `json:"minimum,omitempty,string"`          // section 6.2.4
	ExclusiveMinimum bool   `json:"exclusiveMinimum,omitempty,string"` // section 6.2.5
	MaxLength        int    `json:"maxLength,omitempty,string"`        // section 6.3.1
	MinLength        int    `json:"minLength,omitempty,string"`        // section 6.3.2
	Pattern          string `json:"pattern,omitempty"`                 // section 6.3.3
	MaxItems         int    `json:"maxItems,omitempty,string"`         // section 6.4.1
	MinItems         int    `json:"minItems,omitempty,string"`         // section 6.4.2
	UniqueItems      bool   `json:"uniqueItems,omitempty,string"`      // section 6.4.3
	MaxContains      uint   `json:"maxContains,omitempty,string"`      // section 6.4.4
	MinContains      uint   `json:"minContains,omitempty,string"`      // section 6.4.5
	MaxProperties    int    `json:"maxProperties,omitempty,string"`    // section 6.5.1
	MinProperties    int    `json:"minProperties,omitempty,string"`    // section 6.5.2
	Required         bool   `json:"required,omitempty,string"`         // section 6.5.3

	// RFC draft-bhutton-json-schema-validation-00, section 8
	ContentEncoding  string      `json:"contentEncoding,omitempty"`   // section 8.3
	ContentMediaType string      `json:"contentMediaType,omitempty"`  // section 8.4
	Title            string      `json:"title,omitempty"`             // section 9.1
	Default          string      `json:"default,omitempty"`           // section 9.2
	Deprecated       bool        `json:"deprecated,omitempty,string"` // section 9.3
	ReadOnly         bool        `json:"readOnly,omitempty,string"`   // section 9.4
	WriteOnly        bool        `json:"writeOnly,omitempty,string"`  // section 9.4
	Example          string      `json:"example,omitempty"`           // section 9.5
	Examples         string      `json:"examples,omitempty"`          // section 9.5
	Src              string      `json:"src,omitempty"`
	Dst              string      `json:"dst,omitempty"`
	Fullname         string      `json:"fullname,omitempty"`
	AllowEmptyValue  bool        `json:"allowEmptyValue,omitempty,string"`
	Lineschema       *Lineschema `json:"-"`
}

type Meta struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type Lineschema struct {
	Meta  *Meta
	Items []*LineschemaItem
}

const (
	TOKEN_BEGIN = ','
	TOKEN_END   = '='
	EOF         = "\n"
)

// ParseLineschema 解析lineschema
func ParseLineschema(lineschema string) (jsonline *Lineschema, err error) {
	lineschema = compress(lineschema)
	lines := strings.Split(lineschema, EOF)
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
		err = validItem(item)
		if err != nil {
			err = errors.WithMessage(err, fmt.Sprintf(" got:%s", line))
			return nil, err
		}
		srcOrDst := strings.ReplaceAll(item.Fullname, "[]", ".#")
		if item.Src == "" {
			item.Src = srcOrDst
		} else if item.Dst == "" {
			item.Dst = srcOrDst
		}
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

func validItem(item *LineschemaItem) (err error) {
	if item.Fullname == "" {
		err = errors.New("fullname required ")
		return err
	}
	if item.Src == "" && item.Dst == "" {
		err = errors.New("at least one of dst/src required ")
		return err
	}
	return nil
}
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
