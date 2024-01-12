package lineschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cast"
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
	Ref              string      `json:"ref,omitempty"`
	Fullname         string      `json:"fullname,omitempty"`
	AllowEmptyValue  bool        `json:"allowEmptyValue,omitempty,string"`
	Lineschema       *Lineschema `json:"-"`
}

func (jItem LineschemaItem) String() (jsonStr string) {
	copy := jItem
	copy.Required = false // 转换成json schema时 required 单独处理
	// 这部分字段隐藏
	copy.Fullname = ""
	b, _ := json.Marshal(copy)
	jsonStr = string(b)
	return jsonStr
}

func (jItem *LineschemaItem) InitPath() {
	if jItem.Path == "" {
		pathArrPlaceHold := ".#"
		jItem.Path = strings.ReplaceAll(jItem.Fullname, "[]", pathArrPlaceHold)
		if strings.EqualFold(jItem.Type, "array") { // type 为array是路径增加数组标识
			jItem.Path = fmt.Sprintf("%s%s", strings.TrimSuffix(jItem.Path, pathArrPlaceHold), pathArrPlaceHold)
		}
		jItem.Path = strings.Trim(jItem.Path, ".") // 开头为.#-数组,需要改成 #开头
	}
}

func (jItem LineschemaItem) ToKVS(namespance string) (kvs kvstruct.KVS) {
	jsonStr := jItem.String()
	kvs = kvstruct.JsonToKVS(jsonStr, namespance)
	return kvs
}
func (jItem LineschemaItem) enum2Array() (enum []interface{}, enumNames []interface{}, err error) {
	if jItem.Enum != "" {
		err = json.Unmarshal([]byte(jItem.Enum), &enum)
		if err != nil {
			return nil, nil, err
		}
	}
	if jItem.EnumNames != "" {
		err = json.Unmarshal([]byte(jItem.EnumNames), &enumNames)
		if err != nil {
			return nil, nil, err
		}
	}
	return enum, enumNames, nil
}

func (jItem LineschemaItem) ToJsonSchemaKVS() (kvs kvstruct.KVS, err error) {
	kvs = make(kvstruct.KVS, 0)
	arrSuffix := "[]"
	fullname := strings.Trim(jItem.Fullname, ".")
	if !strings.HasPrefix(fullname, arrSuffix) {
		fullname = fmt.Sprintf(".%s", fullname) //增加顶级对象
	}
	arr := strings.Split(fullname, ".")
	kv := kvstruct.KV{
		Key:   `$schema`,
		Value: `http://json-schema.org/draft-07/schema#`,
	}
	kvs = append(kvs, kv)
	prefix := ""
	l := len(arr)
	for i := 0; i < l; i++ {
		key := arr[i]
		//处理数组
		if strings.HasSuffix(key, arrSuffix) {
			key = strings.TrimSuffix(key, arrSuffix)
			prefix = strings.Trim(fmt.Sprintf("%s.%s", prefix, key), ".")
			kv := kvstruct.KV{
				Key:   strings.Trim(fmt.Sprintf("%s.type", prefix), "."),
				Value: "array",
			}
			kvs = append(kvs, kv)
			if i == l-1 {
				fullKey := strings.Trim(fmt.Sprintf("%s.items", prefix), ".")
				attrKvs := jItem.ToKVS(fullKey)
				kvs.AddReplace(attrKvs...)
				enum, enumNames, err := jItem.enum2Array()
				if err != nil {
					return nil, err
				}
				subKvs := enumNames2KVS(enum, enumNames, fullKey)
				kvs.AddReplace(subKvs...)
				continue
			}
			prefix = fmt.Sprintf("%s.items", prefix)
			kv = kvstruct.KV{
				Key:   strings.Trim(fmt.Sprintf("%s.type", prefix), "."),
				Value: "object",
			}
			kvs = append(kvs, kv)
			prefix = fmt.Sprintf("%s.properties", prefix)
			continue
		}

		//处理对象
		if i == l-1 {
			if jItem.Required {
				parentKey := strings.TrimSuffix(prefix, ".properties")
				kv := kvstruct.KV{
					Key:   strings.Trim(fmt.Sprintf("%s.required.-1", parentKey), "."),
					Value: key,
				}
				kvs.AddReplace(kv)
			}
			fullKey := strings.Trim(fmt.Sprintf("%s.%s", prefix, key), ".")
			attrKvs := jItem.ToKVS(fullKey)
			kvs.AddReplace(attrKvs...)
			enum, enumNames, err := jItem.enum2Array()
			if err != nil {
				return nil, err
			}
			subKvs := enumNames2KVS(enum, enumNames, fullKey)
			kvs.AddReplace(subKvs...)
			continue
		}

		prefix = strings.Trim(fmt.Sprintf("%s.%s", prefix, key), ".")
		kv := kvstruct.KV{
			Key:   strings.Trim(fmt.Sprintf("%s.type", prefix), "."),
			Value: "object",
		}
		kvs = append(kvs, kv)
		prefix = fmt.Sprintf("%s.properties", prefix)
	}
	return kvs, nil
}

func enumNames2KVS(enums []interface{}, enumNames []interface{}, prefix string) (kvs kvstruct.KVS) {
	kvs = make(kvstruct.KVS, 0)
	if len(enumNames) < 1 {
		return kvs
	}
	enumLen := len(enums)
	for i, enumName := range enumNames {
		if i >= enumLen {
			continue
		}
		enum := enums[i]
		typ := ""
		switch enum.(type) {
		case int, float64, int64:
			typ = "int"
		}
		kv := kvstruct.KV{
			Type:  kvstruct.KVType(typ),
			Key:   strings.Trim(fmt.Sprintf("%s.oneOf.%d.const", prefix, i), "."),
			Value: cast.ToString(enum),
		}
		kvs.Add(kv)
		kv = kvstruct.KV{
			Key:   strings.Trim(fmt.Sprintf("%s.oneOf.%d.title", prefix, i), "."),
			Value: cast.ToString(enumName),
		}
		kvs.Add(kv)
	}
	return kvs
}
