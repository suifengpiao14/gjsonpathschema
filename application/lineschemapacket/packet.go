package lineschemapacket

import (
	"fmt"

	"github.com/suifengpiao14/lineschema"
	"github.com/suifengpiao14/stream"
)

// lineschema 格式数据包
type LineschemaPacketI interface {
	GetRoute() (mehtod string, path string) // 网络传输地址，http可用method,path标记
	UnpackSchema() (lineschema string)      // 解包配置 从网络数据到程序
	PackSchema() (lineschema string)        // 封包配置 程序到网络
}

func RegisterLineschemaPacket(pack LineschemaPacketI) (err error) {
	method, path := pack.GetRoute()
	unpackId, packId := makeLineschemaPacketKey(method, path)
	unpackSchema, packSchema := pack.UnpackSchema(), pack.PackSchema()
	unpackLineschema, err := lineschema.ParseLineschema(unpackSchema)
	if err != nil {
		return err
	}
	packLineschema, err := lineschema.ParseLineschema(packSchema)
	if err != nil {
		return err
	}
	err = RegisterLineschema(unpackId, *unpackLineschema)
	if err != nil {
		return err
	}
	err = RegisterLineschema(packId, *packLineschema)
	if err != nil {
		return err
	}
	return err
}

func GetLineschemaPackageHandlerFn(api LineschemaPacketI) (unpackHandlerFns []stream.HandlerFn, packHandlerFns []stream.HandlerFn, err error) {
	method, path := api.GetRoute()
	idIn, idOut := makeLineschemaPacketKey(method, path)
	inClineshema, err := GetClineschema(idIn)
	if err != nil {
		return nil, nil, err
	}

	outClineshema, err := GetClineschema(idOut)
	if err != nil {
		return nil, nil, err
	}
	unpackHandlerFns = []stream.HandlerFn{
		inClineshema.ValidatePacketFn(),
		inClineshema.MergeDefaultFn(),
		inClineshema.TransferToFormatFn(),
	}
	packHandlerFns = []stream.HandlerFn{
		outClineshema.TransferToTypeFn(),
		outClineshema.MergeDefaultFn(),
		outClineshema.ValidatePacketFn(),
	}

	return unpackHandlerFns, packHandlerFns, nil

}

func makeLineschemaPacketKey(method string, path string) (idIn string, idOut string) {
	idIn = fmt.Sprintf("%s-%s-input", method, path)
	idOut = fmt.Sprintf("%s-%s-output", method, path)
	return idIn, idOut
}
