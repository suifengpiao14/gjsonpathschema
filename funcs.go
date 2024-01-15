package lineschema

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
	"github.com/suifengpiao14/kvstruct"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
)

var (
	ERROR_INVALID                 = errors.New("lineschema.Validate")
	ERROR_VALIDATE_INPUT_NIL      = errors.New("lineschema.Validate input is nil")
	ERROR_VALIDATE_JSONLoader_NIL = errors.New("lineschema.Validate JSONLoader is nil")
)

// Validate 验证
func Validate(input []byte, jsonLoader gojsonschema.JSONLoader) (err error) {
	if input == nil {
		return ERROR_VALIDATE_INPUT_NIL
	}
	if jsonLoader == nil {
		return ERROR_VALIDATE_JSONLoader_NIL
	}
	documentLoader := gojsonschema.NewBytesLoader(input)
	result, err := gojsonschema.Validate(jsonLoader, documentLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	msgArr := make([]string, 0)
	for _, resultError := range result.Errors() {
		msgArr = append(msgArr, resultError.String())
	}
	err = errors.WithMessagef(ERROR_INVALID, "400:4000001:input args validate errors:%s", strings.Join(msgArr, ","))
	return err
}

func MergeDefault(data []byte, defaul []byte) (merge []byte, err error) {
	if defaul == nil {
		return data, nil
	}
	kvs := kvstruct.JsonToKVS(string(defaul), "")
	for _, kv := range kvs {
		if kv.Value == "" {
			continue
		}
		v := gjson.GetBytes(data, kv.Key).String()
		if v == "" || v == "null" {
			data, err = sjson.SetBytes(data, kv.Key, kv.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	return data, err
}

//ConvertFomat 转换格式
func ConvertFomat(input []byte, pathMap string) (output []byte) {
	if pathMap == "" {
		return input
	}
	outputStr := gjson.GetBytes(input, pathMap).String()
	output = []byte(outputStr)
	return output
}

//GenerateDefaultJSON 从jsonschema 中提取默认值，组成json
func GenerateDefaultJSON(jschema []byte) (defaultJson []byte, err error) {
	var schema = make(map[string]interface{})
	err = json.Unmarshal([]byte(jschema), &schema)
	if err != nil {
		return nil, err
	}
	defaultJsonI := generateDefaultJSON(schema)
	if funcs.IsNil(defaultJsonI) {
		return nil, nil
	}
	defaultJson, err = json.Marshal(defaultJsonI)
	if err != nil {
		return nil, err
	}
	return defaultJson, nil

}

//generateDefaultJSON 从jsonschema 中提取默认值，组成json
func generateDefaultJSON(schema map[string]interface{}) interface{} {
	if _, exists := schema["default"]; exists {
		return schema["default"]
	}

	typ := cast.ToString(schema["type"])
	switch strings.ToLower(typ) {
	case "object":
		properties, ok := schema["properties"].(map[string]interface{})
		if !ok {
			return nil
		}

		result := make(map[string]interface{})
		for key, prop := range properties {
			defaultValue := generateDefaultJSON(prop.(map[string]interface{}))
			result[key] = defaultValue
		}

		return result

	case "array":
		items, ok := schema["items"].(map[string]interface{})
		if !ok {
			return nil
		}

		var result []interface{}
		defaultValue := generateDefaultJSON(items)
		result = append(result, defaultValue)

		return result

	default:
		return nil
	}
}
