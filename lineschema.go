package lineschema

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	_ "github.com/suifengpiao14/gjsonmodifier"
	"github.com/suifengpiao14/kvstruct"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Meta struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type Lineschema struct {
	Meta  *Meta
	Items LineschemaItems
}

type LineschemaItems []*LineschemaItem

func (ls *LineschemaItems) Add(lineschemaItems ...*LineschemaItem) {
	for _, l := range lineschemaItems {
		l.InitPath()
		*ls = append(*ls, l)
	}
}

var jsonschemalineItemOrder = []string{
	"fullname", "src", "dst", "type", "format", "pattern", "enum", "required", "allowEmptyValue", "title", "description", "default", "comment", "example", "deprecated", "const",
	"multipleOf", "maximum", "exclusiveMaximum", "minimum", "exclusiveMinimum", "maxLength", "minLength",
	"maxItems",
	"minItems",
	"uniqueItems",
	"maxContains",
	"minContains",
	"maxProperties",
	"minProperties",
	"contentEncoding",
	"contentMediaType",
	"readOnly",
	"writeOnly",
}

func (l *Lineschema) Validate() (err error) {
	if l.Meta == nil {
		return errors.Errorf("lineschema meta required")
	}
	if l.Meta.ID == "" {
		return errors.Errorf("lineschema meta.Id required")
	}
	return err
}

func (l *Lineschema) UniqKey() (uniqKey string) {
	s := l.String()
	uniqKey = funcs.Md5Lower(s)
	return uniqKey
}

func (l *Lineschema) String() string {
	lineArr := make([]string, 0)
	lineArr = append(lineArr, fmt.Sprintf("version=%s,id=%s", l.Meta.Version, l.Meta.ID))
	var linemap []map[string]string
	b, err := json.Marshal(l.Items)
	if err != nil {
		err = errors.WithStack(err)
		panic(err)
	}
	err = json.Unmarshal(b, &linemap)
	if err != nil {
		err = errors.WithStack(err)
		panic(err)
	}

	for _, m := range linemap {
		kvArr := make([]string, 0)
		for _, k := range jsonschemalineItemOrder {
			v, ok := m[k]
			if ok {
				if k == "type" && v == "string" {
					continue // 字符串类型,默认不写
				}
				if v == "true" {
					kvArr = append(kvArr, k)
				} else {
					kvArr = append(kvArr, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
		line := strings.Join(kvArr, ",")
		lineArr = append(lineArr, line)
	}
	out := strings.Join(lineArr, EOF)
	return out
}

// BaseNames 获取所有基础名称
func (l *Lineschema) BaseNames() (names []string) {
	names = make([]string, 0)
	for _, item := range l.Items {
		names = append(names, BaseName(item.Fullname))
	}
	return names
}

func (l *Lineschema) JsonSchema() (jsonschemaByte []byte, err error) {
	kvs := kvstruct.KVS{
		{Key: "$schema", Value: "http://json-schema.org/draft-07/schema#"},
	}
	for _, item := range l.Items {
		subKvs, err := item.ToJsonSchemaKVS()
		if err != nil {
			return nil, err
		}
		kvs.Add(subKvs...)
	}

	jsonschemaByte = []byte("")
	for _, kv := range kvs {
		if gjson.GetBytes(jsonschemaByte, kv.Key).Exists() { // 已经存在的，不覆盖（防止 array、object 在其子属性说明后，导致覆盖）
			continue
		}
		if kvstruct.IsJsonStr(kv.Value) {
			jsonschemaByte, err = sjson.SetRawBytes(jsonschemaByte, kv.Key, []byte(kv.Value))
			if err != nil {
				return nil, err
			}
			continue
		}
		var value interface{}
		value = kv.Value
		baseKey := BaseName(kv.Key)
		switch baseKey {
		case "exclusiveMaximum", "exclusiveMinimum", "deprecated", "readOnly", "writeOnly", "uniqueItems":
			value = kv.Value == "true"
		case "multipleOf", "maximum", "minimum", "maxLength", "minLength", "maxItems", "minItems", "maxContains", "minContains", "maxProperties", "minProperties":
			value, _ = strconv.Atoi(kv.Value)
		}
		jsonschemaByte, err = sjson.SetBytes(jsonschemaByte, kv.Key, value)
		if err != nil {
			return nil, err
		}
	}
	return jsonschemaByte, nil
}

// TransferToFormat 获取转换对象 源为type，目标为format
func (lineschema Lineschema) TransferToFormat() (transfers Transfers) {
	transfers = make(Transfers, 0)
	for _, item := range lineschema.Items {

		src := TransferUnit{
			Path: item.Path,
			Type: item.Type,
		}
		typ := item.Type
		transferFunc, ok := DefaultTransferFuncs.GetByType(item.Format)
		if ok {
			typ = transferFunc.Type
		}

		dst := TransferUnit{
			Path: item.Path,
			Type: typ,
		}
		transfer := Transfer{
			Src: src,
			Dst: dst,
		}
		transfers.Replace(transfer)
	}
	return transfers
}
