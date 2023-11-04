package lineschema

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

type LineschemaTransferItem struct {
	Src LineschemaItem `json:"src"`
	Dst LineschemaItem `json:"dst"`
}

// 外界不可以直接初始化,
type _LineschemaTransfer []LineschemaTransferItem

func NewLineschemaTransfer() (transfer _LineschemaTransfer) {
	return _LineschemaTransfer{}
}

// 新增，存在替换
func (transfer *_LineschemaTransfer) Replace(transferItems ...LineschemaTransferItem) {
	for _, transferItem := range transferItems {
		switch strings.ToLower(transferItem.Dst.Type) {
		case "array":
			if _, ok := DefaultLineschemaTransferRelations.GetByFormat(transferItem.Dst.Format); !ok {
				continue
			}
		case "object":
			continue
		}

		exists := false
		for i, item := range *transfer {
			if item.Dst.Path == transferItem.Dst.Path {
				(*transfer)[i] = transferItem
				exists = true
				break
			}
		}
		if !exists {
			*transfer = append(*transfer, transferItem)
		}
	}
}

func (transfer _LineschemaTransfer) Reverse() (reversedTransfer _LineschemaTransfer) {
	reversedTransfer = _LineschemaTransfer{}
	for _, item := range transfer {
		refersedItem := LineschemaTransferItem{
			Src: item.Dst,
			Dst: item.Src,
		}
		reversedTransfer = append(reversedTransfer, refersedItem)
	}
	return reversedTransfer
}

func (t _LineschemaTransfer) String() (gjsonPath string) {
	m := &map[string]interface{}{}
	for _, item := range t {
		dst := item.Dst
		switch strings.ToLower(dst.Type) {
		case "array":
			if _, ok := DefaultLineschemaTransferRelations.GetByFormat(dst.Format); !ok {
				continue
			}
		case "object": // 数组、对象需要遍历内部结构,忽略外部的path
			continue
		}
		arr := strings.Split(dst.Path, ".")
		l := len(arr)
		ref := m
		for i, key := range arr {
			if l == i+1 {
				(*ref)[key] = item.Src.Path
				continue
			}
			if _, ok := (*ref)[key]; !ok {
				(*ref)[key] = &map[string]interface{}{}
			}
			ref = (*ref)[key].(*map[string]interface{}) //递进
		}

	}
	if len(*m) == 0 {
		return ""
	}
	w, isArray := t.recursionWrite(m)
	if isArray {
		gjsonPath = fmt.Sprintf("[%s]", w.String())
	} else {
		gjsonPath = fmt.Sprintf("{%s}", w.String())
	}

	return gjsonPath
}

// 生成路径
func (t _LineschemaTransfer) recursionWrite(m *map[string]interface{}) (w bytes.Buffer, isArray bool) {
	writeComma := false
	for k, v := range *m {
		if writeComma {
			w.WriteString(",")
		}
		writeComma = true
		ref, ok := v.(*map[string]interface{})
		if !ok {
			if k == "#" {
				w.WriteString(cast.ToString(v))
				isArray = true
				continue
			}
			w.WriteString(fmt.Sprintf("%s:%s", k, cast.ToString(v)))

			continue
		}
		subw, subIsArray := t.recursionWrite(ref)
		subwKey := subw.String()
		var subStr string
		if k == "#" {
			subStr = fmt.Sprintf("{%s}|@group", subwKey)
			isArray = true
		} else {
			if !subIsArray {
				subwKey = fmt.Sprintf("{%s}", subwKey) // 子字符串不是数组,默认为对象
			}
			subStr = fmt.Sprintf("%s:%s", k, subwKey)
		}
		w.WriteString(subStr)
	}
	return w, isArray
}

type LineschemaTransferRelation struct {
	Format    string `json:"format"`    // 格式
	Type      string `json:"type"`      // 对应类型
	ConvertFn string `json:"convertFn"` // 转换函数名称
}
type LineschemaTransferRelations []LineschemaTransferRelation

func (ms LineschemaTransferRelations) GetByFormat(format string) (m *LineschemaTransferRelation, ok bool) {
	for _, m := range ms {
		if m.Format == format {
			return &m, true
		}
	}
	return nil, false
}
func (ms LineschemaTransferRelations) GetByType(typ string, item *LineschemaItem) (m *LineschemaTransferRelation, ok bool) {
	if strings.EqualFold(typ, "array") && item.Format != "" { // 数组类型修改类型兼容[1,2,3]格式
		item.Path = fmt.Sprintf("%s.#", item.Path)
		item.Type = item.Format
		typ = item.Type
	}
	for _, m := range ms {
		if m.Type == typ {
			return &m, true
		}
	}
	return nil, false
}

// DefaultLineschemaTransferRelations schema format 转类型
var DefaultLineschemaTransferRelations = LineschemaTransferRelations{
	{Format: "int", Type: "int", ConvertFn: ".@tonum"},
	{Format: "number", Type: "number", ConvertFn: ".@tonum"},
	{Format: "bool", Type: "bool", ConvertFn: ".@tobool"},
	{Format: "boolean", Type: "bool", ConvertFn: ".@tobool"},
	{Format: "time", Type: "string", ConvertFn: ".@tostring"},
	{Format: "datetime", Type: "string", ConvertFn: ".@tostring"},
	{Format: "date", Type: "string", ConvertFn: ".@tostring"},
	{Format: "email", Type: "string", ConvertFn: ".@tostring"},
	{Format: "phone", Type: "string", ConvertFn: ".@tostring"},
	{Format: "string", Type: "string", ConvertFn: ".@tostring"},
}
