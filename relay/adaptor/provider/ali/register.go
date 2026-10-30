package ali

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "ali",
		Name:           "Ali",
		DefaultBaseURL: "https://dashscope.aliyuncs.com",
		LegacyType:     17,
	}, func() adaptor.Adaptor {
		return &Adaptor{}
	})
}
