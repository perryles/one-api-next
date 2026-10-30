package xunfeiv2

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "xunfeiv2",
		Name:           "Xunfei V2",
		DefaultBaseURL: "https://spark-api-open.xf-yun.com",
		LegacyType:     48,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
