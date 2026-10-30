package zhipu

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "zhipu",
		Name:           "Zhipu",
		DefaultBaseURL: "https://open.bigmodel.cn",
		LegacyType:     16,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
