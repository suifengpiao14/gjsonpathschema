package lineschemapacket

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/stream"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
)

var clineschemaMap sync.Map

// _Clineschema 编译好的jsonschema
type _Clineschema struct {
	ID                        string `json:"id"`
	Lineschema                lineschema.Lineschema
	Jsonschema                []byte `json:"jsonschema"`
	DefaultJson               []byte `json:"defaultValues"`
	transferToFormatGjsonPath string
	transferToTypeGjsonPath   string
	validateLoader            gojsonschema.JSONLoader
}

func (c _Clineschema) MergeDefaultFn() (fn stream.HandlerFn) {
	return MakeMergeDefaultHandler(c.DefaultJson)
}
func (c _Clineschema) ValidatePacketFn() (fn stream.HandlerFn) {
	return MakeValidateHandler(c.validateLoader)
}

// TransferToFormatFn 采用format 属性转换数据，一般用于input
func (c _Clineschema) TransferToFormatFn() (fn stream.HandlerFn) {
	return MakeTransferHandler(c.transferToFormatGjsonPath)
}

// TransferToTypeFn 采用type 属性转换数据，一般用于output
func (c _Clineschema) TransferToTypeFn() (fn stream.HandlerFn) {
	return MakeTransferHandler(c.transferToTypeGjsonPath)
}

func RegisterLineschema(identify string, lschema lineschema.Lineschema) (err error) {
	v, ok := clineschemaMap.Load(identify)
	if ok {
		err = errors.Errorf("id already registered,id:%s,value:%T", identify, v)
		return err
	}
	err = lschema.Validate()
	if err != nil {
		return err
	}
	jschema, err := lschema.JsonSchema()
	if err != nil {
		return err
	}
	err = ValidateJsonschema(jschema)
	if err != nil {
		return err
	}

	jsonschemaLoader := gojsonschema.NewBytesLoader(jschema)
	defaultJson, err := GenerateDefaultJSON(jschema)
	if err != nil {
		return err
	}
	cJsonschema := _Clineschema{
		ID:                        identify,
		Lineschema:                lschema,
		Jsonschema:                jschema,
		transferToFormatGjsonPath: lschema.TransferToFormatGjsonPath(),
		transferToTypeGjsonPath:   lschema.TransferToTypeGjsonPath(),
		DefaultJson:               defaultJson,
		validateLoader:            jsonschemaLoader,
	}
	clineschemaMap.Store(identify, &cJsonschema)
	return nil
}

var (
	ERROR_NOT_FOUND_CSCHEMA = errors.New("not found _Clineschema")
)

func GetClineschema(id string) (cLineschema *_Clineschema, err error) {
	v, ok := clineschemaMap.Load(id)
	if !ok {
		err = errors.WithMessagef(ERROR_NOT_FOUND_CSCHEMA, "id:%s", id)
		return nil, err
	}
	ref, ok := v.(*_Clineschema)
	if !ok {
		err = errors.Errorf("expect:*_Clineschema,got:%T", v)
		return nil, err
	}
	tmp := *ref // 确保不被外界修改
	return &tmp, nil
}

var draftMap = map[string]string{
	"http://json-schema.org/draft-04/schema": `{"id":"http://json-schema.org/draft-04/schema#","$schema":"http://json-schema.org/draft-04/schema#","description":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"positiveInteger":{"type":"integer","minimum":0},"positiveIntegerDefault0":{"allOf":[{"$ref":"#/definitions/positiveInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"minItems":1,"uniqueItems":true}},"type":"object","properties":{"id":{"type":"string"},"$schema":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"multipleOf":{"type":"number","minimum":0,"exclusiveMinimum":true},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"boolean","default":false},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"boolean","default":false},"maxLength":{"$ref":"#/definitions/positiveInteger"},"minLength":{"$ref":"#/definitions/positiveIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/positiveInteger"},"minItems":{"$ref":"#/definitions/positiveIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"maxProperties":{"$ref":"#/definitions/positiveInteger"},"minProperties":{"$ref":"#/definitions/positiveIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"dependencies":{"exclusiveMaximum":["maximum"],"exclusiveMinimum":["minimum"]},"default":{}}`,
	"http://json-schema.org/draft-06/schema": `{"$schema":"http://json-schema.org/draft-06/schema#","$id":"http://json-schema.org/draft-06/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"examples":{"type":"array","items":{}},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"propertyNames":{"format":"regex"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":{},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":{}}`,
	"http://json-schema.org/draft-07/schema": `{"$schema":"http://json-schema.org/draft-07/schema#","$id":"http://json-schema.org/draft-07/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"$comment":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":true,"readOnly":{"type":"boolean","default":false},"writeOnly":{"type":"boolean","default":false},"examples":{"type":"array","items":true},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":true},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"propertyNames":{"format":"regex"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":true,"enum":{"type":"array","items":true,"minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"contentMediaType":{"type":"string"},"contentEncoding":{"type":"string"},"if":{"$ref":"#"},"then":{"$ref":"#"},"else":{"$ref":"#"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":true}`,
}

// ValidateJsonschema 验证schema是否符合规范
func ValidateJsonschema(lschemaRaw []byte) (err error) {
	metaSchemaRef := gjson.GetBytes(lschemaRaw, "$schema").String()
	if metaSchemaRef == "" {
		metaSchemaRef = "http://json-schema.org/draft-07/schema"
	}
	metaSchemaRef = strings.TrimSuffix(metaSchemaRef, "#")
	metaSchema, ok := draftMap[metaSchemaRef]
	if !ok {
		err = errors.Errorf("not found meta schema by ref:%s", metaSchemaRef)
		return err
	}

	schemaLoader := gojsonschema.NewBytesLoader(lschemaRaw)

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
