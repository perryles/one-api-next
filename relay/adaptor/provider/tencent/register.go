package tencent

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "tencent",
		Name:           "Tencent",
		DefaultBaseURL: "https://hunyuan.tencentcloudapi.com",
		LegacyType:     23,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
