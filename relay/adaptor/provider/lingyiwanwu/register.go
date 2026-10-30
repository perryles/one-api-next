package lingyiwanwu

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "lingyiwanwu",
		Name:           "LingYiWanWu",
		DefaultBaseURL: "https://api.lingyiwanwu.com",
		LegacyType:     31,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
