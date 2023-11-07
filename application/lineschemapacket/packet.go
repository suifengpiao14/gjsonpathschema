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

func ServerPackHandlers(api LineschemaPacketI) (packHandlers stream.PackHandlers, err error) {
	method, path := api.GetRoute()
	unpackId, packId := makeLineschemaPacketKey(method, path)
	unpackLineschema, err := GetClineschema(unpackId)
	if err != nil {
		return nil, err
	}

	packLineschema, err := GetClineschema(packId)
	if err != nil {
		return nil, err
	}
	packHandlers = make(stream.PackHandlers, 0)
	packHandlers.Add(
		stream.NewPackHandler(unpackLineschema.ValidatePacketFn(), packLineschema.TransferToTypeFn()),
		stream.NewPackHandler(unpackLineschema.MergeDefaultFn(), packLineschema.MergeDefaultFn()),
		stream.NewPackHandler(unpackLineschema.TransferToFormatFn(), packLineschema.ValidatePacketFn()),
	)
	return packHandlers, nil
}

func SDKPackHandlers(api LineschemaPacketI) (packHandlers stream.PackHandlers, err error) {
	serverPackHandlers, err := ServerPackHandlers(api)
	if err != nil {
		return nil, err
	}
	return serverPackHandlers.Reverse(), nil
}

func makeLineschemaPacketKey(method string, path string) (unpackId string, packId string) {
	unpackId = fmt.Sprintf("%s-%s-input", method, path)
	packId = fmt.Sprintf("%s-%s-output", method, path)
	return unpackId, packId
}
