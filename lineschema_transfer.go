package lineschema

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
)

type TransferUnit struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type Transfer struct {
	Src TransferUnit `json:"src"`
	Dst TransferUnit `json:"dst"`
}

// 外界不可以直接初始化,
type Transfers []Transfer

func NewTransfer() (transfer Transfers) {
	return Transfers{}
}

// 新增，存在替换
func (transfer *Transfers) Replace(transferItems ...Transfer) {
	for _, transferItem := range transferItems {
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

func (transfer Transfers) Reverse() (reversedTransfer Transfers) {
	reversedTransfer = Transfers{}
	for _, item := range transfer {
		refersedItem := Transfer{
			Src: item.Dst,
			Dst: item.Src,
		}
		reversedTransfer = append(reversedTransfer, refersedItem)
	}
	return reversedTransfer
}

// AddTransferModify 在来源路径上增加上目标类型转换函数
func (t Transfers) AddTransferModify() (newT Transfers) {
	newT = make(Transfers, 0)
	for _, transfer := range t {
		transferFunc, ok := DefaultTransferFuncs.GetByType(transfer.Dst.Type)
		if ok {
			transfer.Src.Path = fmt.Sprintf("%s%s", transfer.Src.Path, transferFunc.ConvertFn)
		}
		newT = append(newT, transfer)
	}

	return newT

}

func (t Transfers) String() (gjsonPath string) {
	newT := t.AddTransferModify()
	m := &map[string]interface{}{}
	for _, item := range newT {
		dst := item.Dst
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
func (t Transfers) recursionWrite(m *map[string]interface{}) (w bytes.Buffer, isArray bool) {
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

//PathModifyFn 路径修改函数
type PathModifyFn func(path string) (newPath string)

//PathModifyFnCameCase 将路径改成小驼峰格式
func PathModifyFnCameCase(path string) (newPath string) {
	newPath = funcs.CamelCase(path, false, false)
	return
}

//PathModifyFnSnakeCase 将路径转为下划线格式
func PathModifyFnSnakeCase(path string) (newPath string) {
	newPath = funcs.SnakeCase(path)
	return
}

//PathModifyFnLower 将路径转为小写格式
func PathModifyFnLower(path string) (newPath string) {
	return strings.ToLower(path)
}

//PathModifyFnTrimPrefixFn 生成剔除前缀修改函数
func PathModifyFnTrimPrefixFn(prefix string) (pathModifyFn PathModifyFn) {
	return func(path string) (newPath string) {
		return strings.TrimPrefix(path, prefix)
	}
}

//ModifyPath 修改转换路径
func (t Transfers) ModifyPath(srcPathModifyFns []PathModifyFn, dstPathModifyFns []PathModifyFn) (nt Transfers) {
	nt = make(Transfers, 0)
	for _, l := range t {
		src := l.Src
		dst := l.Dst
		for _, fn := range srcPathModifyFns {
			src.Path = fn(src.Path)
		}

		for _, fn := range dstPathModifyFns {
			dst.Path = fn(dst.Path)
		}
		item := Transfer{
			Src: src,
			Dst: dst,
		}
		nt.Replace(item)
	}
	return nt
}

type TransferFunc struct {
	Type      string `json:"type"`      // 对应类型
	ConvertFn string `json:"convertFn"` // 转换函数名称
}
type TransferFuncs []TransferFunc

func (ts TransferFuncs) GetByType(typ string) (t *TransferFunc, ok bool) {
	for _, transfer := range ts {
		if strings.EqualFold(transfer.Type, typ) {
			return &transfer, true
		}
	}
	return nil, false
}

// DefaultTransferFuncs schema format 转类型
var DefaultTransferFuncs = TransferFuncs{
	{Type: "int", ConvertFn: ".@tonum"},
	{Type: "number", ConvertFn: ".@tonum"},
	{Type: "float", ConvertFn: ".@tonum"},
	{Type: "bool", ConvertFn: ".@tobool"},
	{Type: "string", ConvertFn: ".@tostring"},
}

func ToGoTypeTransfer(dst any) (lineschemaTransfer Transfers) {
	if dst == nil {
		return nil
	}
	rv := reflect.Indirect(reflect.ValueOf(dst))
	switch rv.Kind() {
	case reflect.Array:
		return str2StructTransfer(rv, "#")
	case reflect.Struct:
		return str2StructTransfer(rv, "")
	case reflect.Int64, reflect.Float64, reflect.Int:
		return str2SimpleTypeTransfer("int", "")
	case reflect.Bool:
		return str2SimpleTypeTransfer("bool", "")
	}
	return
}

func str2SimpleTypeTransfer(typ string, path string) (lineschemaTransfer Transfers) {
	if path == "" {
		path = "@this"
	}
	return Transfers{
		Transfer{
			Dst: TransferUnit{
				Path: path,
				Type: typ,
			},
			Src: TransferUnit{
				Path: path,
				Type: "string",
			},
		},
	}
}

func str2StructTransfer(rv reflect.Value, prefix string) (lineschemaTransfer Transfers) {
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()
	if prefix != "" {
		prefix = strings.TrimRight(prefix, ".")
		prefix = fmt.Sprintf("%s.", prefix)
	}
	lineschemaTransfer = make(Transfers, 0)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		typ := field.Type.String()
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue // Skip fields without json tag or with "-" tag
		}
		path := fmt.Sprintf("%s%s", prefix, typ)
		linschemaT := Transfer{
			Dst: TransferUnit{
				Path: path,
				Type: typ,
			},
			Src: TransferUnit{
				Path: path,
				Type: "string",
			},
		}
		lineschemaTransfer = append(lineschemaTransfer, linschemaT)
	}

	return lineschemaTransfer
}
