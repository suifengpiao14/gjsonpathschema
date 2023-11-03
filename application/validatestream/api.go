package validatestream

import (
	"fmt"

	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/stream"
)

type LineschemaApi interface {
	GetRoute() (method string, path string)
	GetInputSchema() (lineschema string)
	GetOutputSchema() (lineschema string)
}

func RegisterLineschemaApi(api LineschemaApi) (err error) {
	method, path := api.GetRoute()
	idIn, idOut := MakeLineschemaApiKey(method, path)
	inschema, outschema := api.GetInputSchema(), api.GetOutputSchema()
	inLineschema, err := lineschema.ParseLineschema(inschema)
	if err != nil {
		return err
	}
	outLineschema, err := lineschema.ParseLineschema(outschema)
	if err != nil {
		return err
	}
	err = RegisterLineschema(idIn, *inLineschema)
	if err != nil {
		return err
	}
	err = RegisterLineschema(idOut, *outLineschema)
	if err != nil {
		return err
	}
	return err
}

func GetApiStreamHandlerFn(api LineschemaApi) (inHandlerFns []stream.HandlerFn, outHandlerFns []stream.HandlerFn, err error) {
	method, path := api.GetRoute()
	idIn, idOut := MakeLineschemaApiKey(method, path)
	inClineshema, err := GetClineschema(idIn)
	if err != nil {
		return nil, nil, err
	}

	outClineshema, err := GetClineschema(idOut)
	if err != nil {
		return nil, nil, err
	}
	inHandlerFns = []stream.HandlerFn{
		inClineshema.ValidateStreamFn(),
		inClineshema.MergeDefaultStreamFn(),
		inClineshema.TransferToFormatStreamFn(),
	}
	outHandlerFns = []stream.HandlerFn{
		outClineshema.TransferToTypeStreamFn(),
		outClineshema.ValidateStreamFn(),
	}

	return inHandlerFns, outHandlerFns, nil

}

func MakeLineschemaApiKey(method string, path string) (idIn string, idOut string) {
	idIn = fmt.Sprintf("%s-%s-input", method, path)
	idOut = fmt.Sprintf("%s-%s-output", method, path)
	return idIn, idOut
}
