package baidu

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "baidu",
		Name:           "Baidu",
		DefaultBaseURL: "https://aip.baidubce.com",
		LegacyType:     15,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
