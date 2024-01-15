package lineschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/kvstruct"
)

// Json2lineSchema
func Json2lineSchema(jsonStr string) (out *Lineschema, err error) {
	out = &Lineschema{
		Meta: &Meta{
			Version: "http://json-schema.org/draft-07/schema#",
			ID:      "example",
		},
		Items: make([]*LineschemaItem, 0),
	}
	var input interface{}
	err = json.Unmarshal([]byte(jsonStr), &input)
	if err != nil {
		return nil, err
	}
	rv := reflect.Indirect(reflect.ValueOf(input))
	items := parseOneJsonKey2Line(rv, "")
	out.Items = items.Unique()
	return out, nil
}

// AssertBasicType 根据案例值（数组、对象不处理，只处理基本类型），推断lineschemaItem 的type 和format，次函数har解析时，Query部分需要在包外使用
func AssertBasicType(rv reflect.Value) (typ string, format string, value any) {
	rv = reflect.Indirect(rv)
	kind := rv.Kind()
	if kind == reflect.Interface {
		rv = reflect.Indirect(rv.Elem())
	}
	switch kind {
	case reflect.Bool:
		return "string", "boolean", cast.ToString(rv.Bool())
	case reflect.Int, reflect.Int64:
		return "string", "int", cast.ToString(rv.Int())
	case reflect.Float32, reflect.Float64:
		example := cast.ToString(rv.Float())
		format := "float"
		if !strings.Contains(example, ".") {
			format = "int"
		}
		return "string", format, example
	case reflect.String:
		value := rv.String()
		return "string", "string", value
	}
	return "null", "null", rv.Interface()
}

func parseOneJsonKey2Line(rv reflect.Value, fullname string) (items LineschemaItems) {
	items = make([]*LineschemaItem, 0)
	rv = reflect.Indirect(rv)
	kind := rv.Kind()
	switch kind {
	case reflect.Array, reflect.Slice:
		l := rv.Len()
		if l == 0 {
			item := &LineschemaItem{
				Type:     "array",
				Fullname: fmt.Sprintf("%s[]", fullname),
				Example:  "[]",
			}
			items = append(items, item)
			break
		}
		for i := 0; i < l; i++ {
			v := rv.Index(i)
			subFullname := fmt.Sprintf("%s[]", fullname)
			subItems := parseOneJsonKey2Line(v, subFullname)
			items = append(items, subItems...)
		}
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			subFullname := k
			if fullname != "" {
				subFullname = fmt.Sprintf("%s.%s", fullname, k)
			}
			subItems := parseOneJsonKey2Line(iter.Value(), subFullname)
			items = append(items, subItems...)
		}
	case reflect.Interface, reflect.Ptr:
		rv = rv.Elem()
		return parseOneJsonKey2Line(rv, fullname)
	default: // 默认返回null,避免字段丢失
		typ, format, value := AssertBasicType(rv)
		if format == typ {
			format = "" // 对应类型和格式一致的，忽略格式
		}
		item := &LineschemaItem{
			Type:     typ,
			Fullname: fullname,
			Format:   format,
			Example:  cast.ToString(value),
		}
		items = append(items, item)
	}
	for i := range items {
		items[i].InitPath()
	}
	return items
}

// Jsonschema2Lineschema json schema 转 line schema
func Jsonschema2Lineschema(jsonschema string) (lineschema *Lineschema, err error) {
	var schema map[string]interface{}
	err = json.Unmarshal([]byte(jsonschema), &schema)
	if err != nil {
		return nil, err
	}
	kvs := jsonSchema2KVS(schema, "")
	version, _ := kvs.GetFirstByKey("$schema")
	id, _ := kvs.GetFirstByKey("$id")
	if id.Value == "" {
		id.Value = "example"
	}
	kvs1 := dealPropertiesAndItemsdealRequired(kvs)
	kvs2 := dealRequired(kvs1)
	m := make(map[string][][2]string)
	for _, kv := range kvs2 {
		if strings.HasPrefix(kv.Key, "$") {
			continue
		}
		lastDot := strings.LastIndex(strings.Trim(kv.Key, "."), ".")
		fullname := ""
		key := kv.Key
		if lastDot > -1 {
			fullname, key = kv.Key[:lastDot], kv.Key[lastDot+1:]
		}
		if _, ok := m[fullname]; !ok {
			m[fullname] = make([][2]string, 0)
		}

		m[fullname] = append(m[fullname], [2]string{key, kv.Value})
	}

	var w bytes.Buffer
	w.WriteString(fmt.Sprintf("version=%s,id=%s\n", version.Value, id.Value))
	for fullname, linePairs := range m {
		if fullname == "" {
			continue
		}
		pairs := make([]string, 0)
		pairs = append(pairs, fmt.Sprintf("fullname=%s", fullname))
		for _, pair := range linePairs {
			pairs = append(pairs, strings.Join(pair[:], "="))
		}
		w.WriteString(strings.Join(pairs, ","))
		w.WriteString("\n")
	}

	lineschemastr := w.String()

	lineschema, err = ParseLineschema(lineschemastr)
	if err != nil {
		return nil, err
	}

	return lineschema, nil
}

// jsonschema 转 kvs
func jsonSchema2KVS(schema map[string]interface{}, prefix string) kvstruct.KVS {
	kvs := kvstruct.KVS{}
	for key, value := range schema {
		fieldName := fmt.Sprintf("%s%s", prefix, key)
		switch valueType := value.(type) {
		case map[string]interface{}:
			// 递归处理子对象
			kvs.Add(jsonSchema2KVS(valueType, fieldName+".")...)
		case string:
			kvs.Add(kvstruct.KV{
				Key:   fieldName,
				Value: valueType,
			})
		case []interface{}:
			b, _ := json.Marshal(value)
			valueStr := string(b)
			kvs.Add(kvstruct.KV{
				Key:   fieldName,
				Value: valueStr,
			})
		default:
			kvs.Add(kvstruct.KV{
				Key:   fieldName,
				Value: cast.ToString(value),
			})
		}
	}
	return kvs
}

func dealPropertiesAndItemsdealRequired(kvs kvstruct.KVS) (newKvs kvstruct.KVS) {
	newKvs = kvstruct.KVS{}
	keywordItem := "items."
	tmpKvs := make(kvstruct.KVS, 0)
	for _, kv := range kvs {
		segments := strings.Split(kv.Key, keywordItem)
		prefix := ""
		for i, segment := range segments {
			parent := fmt.Sprintf("%s%s", prefix, segment)
			parentType := fmt.Sprintf("%stype", parent)
			parentTypeKv, _ := kvs.GetFirstByKey(parentType)
			if parentTypeKv.Value == "array" {
				segments[i] = "[]"
			}
		}
		key := strings.Join(segments, "")
		tmpKvs.Add(kvstruct.KV{Key: key, Value: kv.Value})
	}

	keywordProperties := "properties."
	for _, kv := range kvs {
		segments := strings.Split(kv.Key, keywordProperties)
		prefix := ""
		for i, segment := range segments {
			parent := fmt.Sprintf("%s%s", prefix, segment)
			parentType := fmt.Sprintf("%stype", parent)
			parentTypeKv, _ := kvs.GetFirstByKey(parentType)
			if parentTypeKv.Value == "object" {
				segments[i] = ""
			}
		}
		key := strings.Join(segments, "")
		newKvs.Add(kvstruct.KV{Key: key, Value: kv.Value})
	}
	return newKvs
}

func dealRequired(kvs kvstruct.KVS) (newKvs kvstruct.KVS) {
	newKvs = make(kvstruct.KVS, 0)

	for _, kv := range kvs {
		requiredLastIndex := strings.LastIndex(kv.Key, "required")
		if requiredLastIndex < 0 {
			newKvs.Add(kv)
			continue
		}
		prefix := kv.Key[:requiredLastIndex]
		typeKv, _ := kvs.GetFirstByKey(fmt.Sprintf("%stype", prefix))
		if typeKv.Value != "array" && typeKv.Value != "object" {
			newKvs.Add(kv)
			continue
		}
		var keys = make([]string, 0)
		err := json.Unmarshal([]byte(kv.Value), &keys)
		if err != nil {
			panic(err)
		}
		for _, k := range keys {
			newKvs.Add(kvstruct.KV{
				Key:   fmt.Sprintf("%s%s.required", prefix, k),
				Value: "true",
			})
		}
	}
	return newKvs
}
