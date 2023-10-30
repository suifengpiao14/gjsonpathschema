package gjsonpathschema

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
)

var jsonschemaMap sync.Map

//_Cjsonschema 编译好的jsonschema
type _Cjsonschema struct {
	ID             string `json:"id"`
	Jsonschema     string `json:"jsonschema"`
	DefaultJson    string `json:"defaultValues"`
	validateLoader gojsonschema.JSONLoader
}

func (c _Cjsonschema) MergeDefault(input string) (output string, err error) {
	output, err = MergeDefault(input, c.DefaultJson)
	return output, err
}
func (c _Cjsonschema) Validate(input string) (err error) {
	err = Validate(input, c.validateLoader)
	return err
}

func (c _Cjsonschema) ConvertFomat(input string, pathMap string) (output string) {
	output = ConvertFomat(input, pathMap)
	return output
}

func RegisterSchema(jschema string) (err error) {
	if jschema == "" {
		err = errors.Errorf("json schema required")
		return err
	}
	err = ValidateJsonschema(jschema)
	if err != nil {
		return err
	}

	jsonschemaLoader := gojsonschema.NewStringLoader(jschema)
	defaultJson, err := GenerateDefaultJSON(jschema)
	if err != nil {
		return err
	}
	id := getID(jschema)
	cJsonschema := _Cjsonschema{
		ID:             id,
		Jsonschema:     jschema,
		DefaultJson:    defaultJson,
		validateLoader: jsonschemaLoader,
	}
	jsonschemaMap.Store(id, &cJsonschema)
	return nil
}

//GetSetCJsonschema 获取c jsonschema 或者设置
func GetSetCJsonschema(jschema string) (cJson *_Cjsonschema, err error) {
	id := getID(jschema)
	cJson, err = GetCJsonschema(id)
	if err != nil {
		err = RegisterSchema(jschema)
		if err != nil {
			return nil, err
		}
		cJson, _ = GetCJsonschema(id)
	}
	return cJson, nil
}

func getID(jschema string) (id string) {
	id = gjson.Get(jschema, "$id").String()
	if id == "" {
		id = funcs.Md5Lower(jschema)
	}
	return id
}

func GetCJsonschema(id string) (cJson *_Cjsonschema, err error) {
	v, ok := jsonschemaMap.Load(id)
	if !ok {
		err = errors.Errorf("not found jsonschema by id:%s", id)
		return nil, err
	}
	ref, ok := v.(*_Cjsonschema)
	if !ok {
		err = errors.Errorf("expect:*_Cjsonschema,got:%T", v)
		return nil, err
	}
	tmp := *ref // 确保不被外界修改
	return &tmp, nil
}
