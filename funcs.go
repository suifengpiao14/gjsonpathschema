package gjsonpathschema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/kvstruct"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xeipuuv/gojsonschema"
)

var ERROR_INVALID = errors.New("gojsonschemavalidator.Validate")

// Validate 验证
func Validate(input string, jsonLoader gojsonschema.JSONLoader) (err error) {
	if input == "" {
		jsonschema, err := jsonLoader.LoadJSON()
		if err != nil {
			return err
		}
		jsonMap, ok := jsonschema.(map[string]interface{})
		if !ok {
			err = errors.Errorf("can not convert jsonLoader.LoadJSON() to map[string]interface{}")
			return err
		}
		typ, ok := jsonMap["type"]
		if !ok {
			err = errors.Errorf("jsonschema missing property type :%v", jsonschema)
			return err
		}
		typStr, ok := typ.(string)
		if !ok {
			err = errors.Errorf("can not convert  jsonschema type to string :%v", typ)
			return err

		}
		switch strings.ToLower(typStr) {
		case "object":
			input = "{}"
		case "array":
			input = "[]"
		default:
			err = errors.Errorf("invalid jsonschema type:%v", typStr)
			return err
		}

	}
	documentLoader := gojsonschema.NewStringLoader(input)
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

func MergeDefault(data string, defaul string) (merge string, err error) {
	if defaul == "" {
		return data, nil
	}
	kvs := kvstruct.JsonToKVS(defaul, "")
	for _, kv := range kvs {
		if kv.Value == "" {
			continue
		}
		v := gjson.Get(data, kv.Key).String()
		if v == "" {
			data, err = sjson.Set(data, kv.Key, kv.Value)
			if err != nil {
				return "", err
			}
		}
	}
	return data, err
}

//ConvertFomat 转换格式
func ConvertFomat(input string, pathMap string) (output string) {
	if pathMap == "" {
		return input
	}
	output = gjson.Get(input, pathMap).String()
	return output
}

//GenerateDefaultJSON 从jsonschema 中提取默认值，组成json
func GenerateDefaultJSON(jschema string) (defaultJson string, err error) {
	var schema = make(map[string]interface{})
	err = json.Unmarshal([]byte(jschema), &schema)
	if err != nil {
		return "", err
	}
	defaultJsonI := generateDefaultJSON(schema)
	b, err := json.Marshal(defaultJsonI)
	if err != nil {
		return "", err
	}
	defaultJson = string(b)
	return defaultJson, nil

}

//generateDefaultJSON 从jsonschema 中提取默认值，组成json
func generateDefaultJSON(schema map[string]interface{}) interface{} {
	if _, exists := schema["default"]; exists {
		return schema["default"]
	}

	switch schema["type"] {
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

//SchemaToGJSONPath json schema 转 gjson path
func SchemaToGJSONPath(jschema string, rootPath string) (paths []string, err error) {
	var schema = make(map[string]interface{})
	err = json.Unmarshal([]byte(jschema), &schema)
	if err != nil {
		return nil, err
	}
	paths = schemaToGJSONPath(schema, rootPath)
	return paths, nil

}

//schemaToGJSONPath json schema 转 gjson path
func schemaToGJSONPath(schema map[string]interface{}, currentPath string) []string {
	paths := []string{}

	if schema["type"] == "object" {
		properties, ok := schema["properties"].(map[string]interface{})
		if ok {
			for key, prop := range properties {
				newPath := fmt.Sprintf("%s.%s", currentPath, key)
				paths = append(paths, schemaToGJSONPath(prop.(map[string]interface{}), newPath)...)
			}
		}
	} else if schema["type"] == "array" {
		items, ok := schema["items"].(map[string]interface{})
		if ok {
			newPath := fmt.Sprintf("%s.#", currentPath)
			paths = append(paths, schemaToGJSONPath(items, newPath)...)
		}
	} else {
		paths = append(paths, currentPath)
	}

	return paths
}

//JsonToGJSONPath json 转 gjson path
func JsonToGJSONPath(jschema string, rootPath string) (paths []string, err error) {
	var schema = make(map[string]interface{})
	err = json.Unmarshal([]byte(jschema), &schema)
	if err != nil {
		return nil, err
	}
	paths = jsonToGJSONPath(schema, rootPath)
	return paths, nil

}

func jsonToGJSONPath(data interface{}, currentPath string) []string {
	paths := []string{}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPath := currentPath + "." + key
			paths = append(paths, newPath)
			paths = append(paths, jsonToGJSONPath(value, newPath)...)
		}
	case []interface{}:
		for i, value := range v {
			newPath := fmt.Sprintf("%s.%d", currentPath, i)
			paths = append(paths, newPath)
			paths = append(paths, jsonToGJSONPath(value, newPath)...)
		}
	}

	return paths
}

var draftMap = map[string]string{
	"http://json-schema.org/draft-04/schema": `{"id":"http://json-schema.org/draft-04/schema#","$schema":"http://json-schema.org/draft-04/schema#","description":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"positiveInteger":{"type":"integer","minimum":0},"positiveIntegerDefault0":{"allOf":[{"$ref":"#/definitions/positiveInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"minItems":1,"uniqueItems":true}},"type":"object","properties":{"id":{"type":"string"},"$schema":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"multipleOf":{"type":"number","minimum":0,"exclusiveMinimum":true},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"boolean","default":false},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"boolean","default":false},"maxLength":{"$ref":"#/definitions/positiveInteger"},"minLength":{"$ref":"#/definitions/positiveIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/positiveInteger"},"minItems":{"$ref":"#/definitions/positiveIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"maxProperties":{"$ref":"#/definitions/positiveInteger"},"minProperties":{"$ref":"#/definitions/positiveIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"dependencies":{"exclusiveMaximum":["maximum"],"exclusiveMinimum":["minimum"]},"default":{}}`,
	"http://json-schema.org/draft-06/schema": `{"$schema":"http://json-schema.org/draft-06/schema#","$id":"http://json-schema.org/draft-06/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"examples":{"type":"array","items":{}},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"propertyNames":{"format":"regex"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":{},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":{}}`,
	"http://json-schema.org/draft-07/schema": `{"$schema":"http://json-schema.org/draft-07/schema#","$id":"http://json-schema.org/draft-07/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"$comment":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":true,"readOnly":{"type":"boolean","default":false},"writeOnly":{"type":"boolean","default":false},"examples":{"type":"array","items":true},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":true},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"propertyNames":{"format":"regex"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":true,"enum":{"type":"array","items":true,"minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"contentMediaType":{"type":"string"},"contentEncoding":{"type":"string"},"if":{"$ref":"#"},"then":{"$ref":"#"},"else":{"$ref":"#"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":true}`,
}

//ValidateJsonschema 验证schema是否符合规范
func ValidateJsonschema(jschema string) (err error) {
	metaSchemaRef := gjson.Get(jschema, "$schema").String()
	if metaSchemaRef == "" {
		metaSchemaRef = "http://json-schema.org/draft-07/schema"
	}
	metaSchemaRef = strings.TrimSuffix(metaSchemaRef, "#")
	metaSchema, ok := draftMap[metaSchemaRef]
	if !ok {
		err = errors.Errorf("not found meta schema by ref:%s", metaSchemaRef)
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(jschema)

	// 加载 JSON Schema Validation Draft 7（或其他版本）的元规范
	metaSchemaLoader := gojsonschema.NewStringLoader(metaSchema)

	// 进行验证
	result, err := gojsonschema.Validate(metaSchemaLoader, schemaLoader)
	if err != nil {
		err = errors.Errorf("Error loading JSON Schema or meta-schema: %s\n", err.Error())
		return err
	}
	if !result.Valid() {
		errs := make([]string, 0)
		for _, desc := range result.Errors() {
			errs = append(errs, desc.Description())
		}
		err = errors.New(strings.Join(errs, ";"))
		return err
	}
	return nil
}
