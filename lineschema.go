package lineschema

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
	_ "github.com/suifengpiao14/gjsonmodifier"
	"github.com/suifengpiao14/kvstruct"
	"github.com/suifengpiao14/pathtransfer"
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

func NewLineschema(id string) (lschema *Lineschema) {
	return &Lineschema{
		Meta: &Meta{
			Version: "http://json-schema.org/draft-07/schema#",
			ID:      id,
		},
		Items: make(LineschemaItems, 0),
	}
}

func (ls *Lineschema) Init() {
	for i := range ls.Items {
		ls.Items[i].InitPath()
	}
}

type LineschemaItems []*LineschemaItem

func (ls *LineschemaItems) Add(lineschemaItems ...*LineschemaItem) {
	for _, l := range lineschemaItems {
		l.InitPath()
		*ls = append(*ls, l)
	}
}

// Remove 移除部分属性
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

func (ls LineschemaItems) Len() int           { return len(ls) }
func (ls LineschemaItems) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }
func (ls LineschemaItems) Less(i, j int) bool { return ls[i].Fullname < ls[j].Fullname }

// 检测给定行是否为自定义结构体名称
func CustomDefineStruct(typeName string) (structName string, isCustomDefineStruct bool) {
	typeName = strings.TrimPrefix(typeName, "[]")
	isCustomDefineStruct = !strings.Contains(Type_base_set, fmt.Sprintf(",%s,", typeName))
	if isCustomDefineStruct {
		return typeName, true
	}
	return "", false
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

// GetByType 找出指定类型名的项
func (ls *LineschemaItems) GetByFullName(fullname string) (subItem *LineschemaItem, ok bool) {
	for _, l := range *ls {
		if l.Fullname == fullname {
			return l, true
		}
	}
	return nil, false
}

// GetByType 找出指定类型名的项
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

// Clone 修改属性值时，先clone 避免引起副作用
func (ls *LineschemaItems) Clone() (items *LineschemaItems) {
	clone := make(LineschemaItems, 0)

	for _, l := range *ls {
		tmp := *l
		clone = append(clone, &tmp)
	}
	return &clone
}

const (
	//字段基本类型,其他类型会被认定为自定义结构体
	Type_base_set = `,string,int,float,boolean,numeber,object,array,[]string,[]int,[]float,[]boolean,[]numeber,[]object,[]array,`
)

func (ls *LineschemaItems) flattenArray() {
	for {
		arrayStructName := ""
		for _, item := range *ls {
			if !strings.HasPrefix(item.Type, "[]") {
				continue
			}

			_, ok := CustomDefineStruct(item.Type)
			if !ok {
				continue
			}
			// 替换数组类型
			arrayStructName = item.Type
			for _, item2 := range *ls {
				if item2.Type == item.Fullname {
					item2.Fullname = fmt.Sprintf("%s[]", item2.Fullname)
					item2.Path = fmt.Sprintf("%s.#", item2.Path)
					item2.Type = strings.TrimPrefix(arrayStructName, "[]")
				}
			}
			ls.Remove(item) // 移除数组定义
			break           // 修改ls后从新循环
		}
		if arrayStructName == "" {
			break
		}
	}
}

// flattenObject 将非简单类型项展开
func (ls *LineschemaItems) flattenObject() {

	for {
		ok := false
		customDefineStruct := ""
		for _, item := range *ls {
			customDefineStruct, ok = CustomDefineStruct(item.Type)
			if !ok {
				continue
			}
			//替换对象
			children := ls.GetByParent(customDefineStruct)
			//检测子类是否有递归引用情况,存在则panic报错
			for _, item := range children { // 递归平铺子类
				subStructName, ok := CustomDefineStruct(item.Type)
				if !ok {
					continue
				}
				if subStructName == customDefineStruct {
					err := errors.Errorf("type name(%s) circular reference. children fullname:%s", customDefineStruct, item.Fullname)
					panic(err)
				}
			}

			refItems := ls.GetByType(item.Type)
			for _, refItem := range refItems {
				clone := children.Clone()
				clone.ChangeParent(refItem.Fullname, fmt.Sprintf("%s.", item.Type))
				ls.Add((*clone)...)
				ls.Remove(refItem)
			}
			ls.Remove(children...)

		}
		if customDefineStruct == "" { // 经过循环搜索后不再有自定义类型,则退出
			break
		}
	}
}

// ChangeParent 修改Fullnamne，达到移动节点效果
func (ls *LineschemaItems) ChangeParent(newParent string, oldParent string) {
	for _, l := range *ls {
		fullname := l.Fullname
		if oldParent != "" {
			fullname = strings.TrimPrefix(fullname, oldParent)
		}
		fullname = fmt.Sprintf("%s.%s", newParent, fullname)
		fullname = strings.TrimLeft(fullname, ".")
		l.Fullname = fullname
		l.Path = strings.ReplaceAll(fullname, "[]", ".#")
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

// ResolveRef  Lineschema 新增支持自定义类型，类似引用，调用此函数后，通过重复填充消除引用，在路径转换时必须先展开到基本类型
func (l Lineschema) ResolveRef() (flatten Lineschema) {
	flatten = Lineschema{
		Meta:  l.Meta,
		Items: make(LineschemaItems, 0),
	}
	items := l.Items.Clone()
	items.flattenArray()  // 解决数组结构体定义
	items.flattenObject() // 解决对象结构体定义
	flatten.Items = *items
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
	lineschema := l.ResolveRef()
	for _, item := range lineschema.Items {
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
func (lineschema Lineschema) TransferToFormat() (transfers pathtransfer.Transfers) {
	resolveRef := lineschema.ResolveRef()
	transfers = make(pathtransfer.Transfers, 0)
	for _, item := range resolveRef.Items {

		src := pathtransfer.TransferUnit{
			Path: pathtransfer.Path(item.Path),
			Type: item.Type,
		}
		typ := item.Type
		transferType, ok := pathtransfer.DefaultTransferTypes.GetByType(item.Format)
		if ok {
			typ = transferType.Type
		}

		dst := pathtransfer.TransferUnit{
			Path: pathtransfer.Path(item.Path),
			Type: typ,
		}
		transfer := pathtransfer.Transfer{
			Src: src,
			Dst: dst,
		}
		transfers.AddReplace(transfer)
	}
	return transfers
}

func (lineschema Lineschema) JsonExample() (jsonStr string, err error) {
	resolved := lineschema.ResolveRef()
	for _, item := range resolved.Items {
		valueStr := item.Example
		if valueStr == "" {
			valueStr = item.Default
		}

		setPath := strings.ReplaceAll(item.Fullname, "[]", ".0") // 生成案例时，数组只设置第一个,fullname ,基本数组类型，item.Fullname最后有[],而item.Path 没有.#
		var value any
		value = valueStr
		switch item.Type {
		case "int":
			value = cast.ToInt(valueStr)
		case "boolean", "bool":
			value = cast.ToBool(valueStr)
		}
		jsonStr, err = sjson.Set(jsonStr, setPath, value)
		if err != nil {
			return "", err
		}
	}
	return jsonStr, nil
}
