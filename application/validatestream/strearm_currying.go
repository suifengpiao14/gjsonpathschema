package validatestream

import (
	"context"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/stream"
	"github.com/xeipuuv/gojsonschema"
)

func MakeMergeDefaultHandler(defaultJson []byte) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		newInput, err := MergeDefault(input, defaultJson)
		if err != nil {
			err = errors.WithMessage(err, "merge default value error")
			return nil, err
		}

		return newInput, nil
	}
}

func MakeValidateHandler(validateLoader gojsonschema.JSONLoader) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {
		if validateLoader == nil {
			return input, nil
		}
		err = Validate(input, validateLoader)
		if err != nil {
			return nil, err
		}
		return input, nil
	}
}

func MakeTransferHandler(pathMap string) (fn stream.HandlerFn) {
	return func(ctx context.Context, input []byte) (out []byte, err error) {

		out = ConvertFomat(input, pathMap)
		return out, nil
	}
}
