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

//Remove 移除部分属性
func (ls *LineschemaItems) Remove(moveItems ...*LineschemaItem) {
	tmp := make(LineschemaItems, 0)
	for _, l := range *ls {
		skip := false
		for _, moveItem := range moveItems {
			if l.Fullname == moveItem.Fullname {
				skip = true
			}
		}
		if skip {
			continue
		}
		tmp = append(tmp, l)
	}
	*ls = tmp
}
func (ls LineschemaItems) Unique() (uniqItems LineschemaItems) {
	uniqItems = make(LineschemaItems, 0)
	for _, item := range ls {
		exists := false
		for _, uinq := range uniqItems {
			if item.Fullname == uinq.Fullname {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		uniqItems = append(uniqItems, item)
	}
	return uniqItems
}

// 检测给定行是否为自定义结构体名称
func (ls *LineschemaItems) IsAsType(lineschemaItem LineschemaItem) (isTypeItem bool) {
	for _, l := range *ls {
		typ := l.Type
		typArr := fmt.Sprintf("[]%s", typ)
		if typ == lineschemaItem.Fullname || typArr == lineschemaItem.Fullname {
			return true
		}
	}
	return false
}

// 获取父类名称下的子类、属性
func (ls *LineschemaItems) GetByParent(parentName string) (children LineschemaItems) {
	children = make(LineschemaItems, 0)
	prefix := fmt.Sprintf("%s.", parentName)
	for _, l := range *ls {
		if strings.Contains(l.Fullname, prefix) {
			children.Add(l)
		}
	}
	return children
}

//GetByType 找出指定类型名的项
func (ls *LineschemaItems) GetByType(typeNames ...string) (subItems LineschemaItems) {
	subItems = make(LineschemaItems, 0)
	for _, l := range *ls {
		for _, typeName := range typeNames {
			if typeName == l.Type {
				subItems.Add(l)
			}
		}
	}
	return subItems
}

//Clone 修改属性值时，先clone 避免引起副作用
func (ls *LineschemaItems) Clone() (clone LineschemaItems) {
	clone = make(LineschemaItems, 0)

	for _, l := range *ls {
		tmp := *l
		clone = append(clone, &tmp)
	}
	return clone
}

//flatten 将非简单类型项展开
func (ls *LineschemaItems) flatten() {

	flattenItems := make(LineschemaItems, 0)
	for _, item := range *ls {
		if ls.IsAsType(*item) {
			flattenItems = append(flattenItems, item)
		}
	}

	for _, flattenItem := range flattenItems {
		children := ls.GetByParent(flattenItem.Fullname)
		typeNames := []string{flattenItem.Fullname, fmt.Sprintf("[]%s", flattenItem.Fullname)}
		refItems := ls.GetByType(typeNames...)
		for _, refItem := range refItems {
			clone := children.Clone()
			clone.ChangeParent(refItem.Fullname, "")
			ls.Add(clone...)
			ls.Remove(refItem)
		}
	}
}

//ChangeParent 修改Fullnamne，达到移动节点效果
func (ls *LineschemaItems) ChangeParent(newParent string, oldParent string) {
	for _, l := range *ls {
		fullname := l.Fullname
		if oldParent != "" {
			fullname = strings.TrimPrefix(fullname, oldParent)
		}
		fullname = fmt.Sprintf("%s.%s", newParent, fullname)
		fullname = strings.TrimLeft(".", fullname)
		l.Fullname = fullname
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

//ResolveRef  Lineschema 新增支持自定义类型，类似引用，调用此函数后，通过重复填充消除引用，在路径转换时必须先展开到基本类型
func (l Lineschema) ResolveRef() (flatten Lineschema) {
	flatten = Lineschema{
		Meta:  l.Meta,
		Items: make(LineschemaItems, 0),
	}
	copy(l.Items, flatten.Items)
	flatten.Items.flatten()
	return flatten
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

func (lineschema Lineschema) JsonExample() (jsonStr string, err error) {
	jsonSchema, err := lineschema.JsonSchema()
	if err != nil {
		return "", err
	}
	b, err := GenerateDefaultJSON(jsonSchema)
	if err != nil {
		return "", err
	}
	jsonStr = string(b)
	return jsonStr, nil
}
