package lineschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func (item LineschemaItem) TransferWithFormat() (formated LineschemaItem) {
	transferConfig, ok := DefaultLineschemaTransferRelations.GetByFormat(item.Format)
	if !ok {
		return item
	}
	formated = item
	formated.Path = fmt.Sprintf("%s%s", item.Path, transferConfig.ConvertFn)
	formated.Type = transferConfig.Type
	return formated
}
func (item LineschemaItem) TransferWithType() (formated LineschemaItem) {
	transferConfig, ok := DefaultLineschemaTransferRelations.GetByType(item.Type)
	if !ok {
		return item
	}
	formated = item
	formated.Path = fmt.Sprintf("%s%s", item.Path, transferConfig.ConvertFn)
	formated.Type = transferConfig.Type
	return formated
}

func (lineschema Lineschema) TransferByFormat() (transfer LineschemaTransfer) {
	transfer = NewLineschemaTransfer(lineschema.Meta.Type)
	for _, item := range lineschema.Items {
		tItem := LineschemaTransferItem{
			Src: *item,
			Dst: item.TransferWithFormat(),
		}
		transfer.Items = append(transfer.Items, tItem)
	}
	return transfer
}

//JsonSchemaToLineschema json schema 转 lineschema
func JsonSchemaToLineschema(jschema string, pathPrefix string) (lineschema Lineschema) {
	var schema = make(map[string]interface{})
	err := json.Unmarshal([]byte(jschema), &schema)
	if err != nil {
		panic(err)
	}
	rootPath := LineschemaItem{
		Path: pathPrefix,
	}
	lineschema = jsonSchemaToLineschema(schema, rootPath)
	return lineschema
}

//jsonSchemaToLineschema json schema 转 gjson path
func jsonSchemaToLineschema(schema map[string]interface{}, currentPath LineschemaItem) (lineschema Lineschema) {

	return lineschema
}

//JsonToLineschema json 转 lineschema
func JsonToLineschema(data interface{}, parentKey string) (lineschema Lineschema) {
	schema := JsonToJsonSchema(data, parentKey)
	lineschema = jsonSchemaToLineschema(schema, LineschemaItem{})
	return lineschema
}

//JsonToJsonSchema json 转 jsonschema
func JsonToJsonSchema(data interface{}, parentKey string) map[string]interface{} {
	schema := make(map[string]interface{})

	switch reflect.TypeOf(data).Kind() {
	case reflect.Map:
		schema["type"] = "object"
		properties := make(map[string]interface{})
		for key, value := range data.(map[string]interface{}) {
			properties[key] = JsonToJsonSchema(value, key)
		}
		schema["properties"] = properties
	case reflect.Slice:
		schema["type"] = "array"
		items := JsonToJsonSchema(data.([]interface{})[0], "")
		schema["items"] = items
	default:
		schema["type"] = getType(data)
	}
	return schema
}

func getType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case int:
		return "int"
	default:
		return "null"
	}
}

type LineschemaTransferItem struct {
	Src LineschemaItem `json:"src"`
	Dst LineschemaItem `json:"dst"`
}

const (
	LineschemaTransfer_Type_object = "object"
	LineschemaTransfer_Type_array  = "array"
)

type LineschemaTransfer struct {
	Type  string
	Items []LineschemaTransferItem
}

func NewLineschemaTransfer(typ string) (transfer LineschemaTransfer) {
	return LineschemaTransfer{
		Type: typ,
	}
}

//新增，存在替换
func (transfer LineschemaTransfer) Replace(transferItems ...LineschemaTransferItem) {
	for _, transferItem := range transferItems {
		exists := false
		for i, item := range transfer.Items {
			if item.Src.Path == transferItem.Src.Path {
				transfer.Items[i] = transferItem
				exists = true
				break
			}
		}
		if !exists {
			transfer.Items = append(transfer.Items, transferItem)
		}
	}
}
func (transfer LineschemaTransfer) IsArray() bool {
	return transfer.Type == LineschemaTransfer_Type_array
}

func (transfer LineschemaTransfer) Reverse() (reversedTransfer LineschemaTransfer) {
	reversedTransfer = LineschemaTransfer{
		Type:  transfer.Type,
		Items: make([]LineschemaTransferItem, 0),
	}
	for _, item := range transfer.Items {
		refersedItem := LineschemaTransferItem{
			Src: item.Dst,
			Dst: item.Src,
		}
		reversedTransfer.Items = append(reversedTransfer.Items, refersedItem)
	}
	return reversedTransfer
}

func (transfer LineschemaTransfer) String() (gojsonPath string) {
	var (
		w     bytes.Buffer
		begin rune = '{'
		end   rune = '}'
	)
	if transfer.IsArray() {
		begin = '['
		end = ']'
	}
	w.WriteRune(begin)
	for i, item := range transfer.Items {
		if i > 0 {
			w.WriteRune(',')
		}
		w.WriteString(item.Src.Path)
		w.WriteRune(':')
		w.WriteString(item.Dst.Path)
	}
	w.WriteRune(end)
	return w.String()
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
func (ms LineschemaTransferRelations) GetByType(typ string) (m *LineschemaTransferRelation, ok bool) {
	for _, m := range ms {
		if m.Type == typ {
			return &m, true
		}
	}
	return nil, false
}

//DefaultLineschemaTransferRelations schema format 转类型
var DefaultLineschemaTransferRelations = LineschemaTransferRelations{
	{Format: "int", Type: "int", ConvertFn: "@tonum"},
	{Format: "number", Type: "number", ConvertFn: "@tonum"},
	{Format: "bool", Type: "bool", ConvertFn: "@tobool"},
	{Format: "boolean", Type: "bool", ConvertFn: "@tobool"},
	{Format: "time", Type: "string", ConvertFn: "@tostring"},
	{Format: "datetime", Type: "string", ConvertFn: "@tostring"},
	{Format: "date", Type: "string", ConvertFn: "@tostring"},
	{Format: "email", Type: "string", ConvertFn: "@tostring"},
	{Format: "phone", Type: "string", ConvertFn: "@tostring"},
	{Format: "string", Type: "string", ConvertFn: "@tostring"},
}
