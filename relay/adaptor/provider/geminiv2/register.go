package geminiv2

import (
	"github.com/perryles/one-api-next/relay/adaptor"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/registry"
)

func init() {
	registry.Register(registry.ChannelMeta{
		ID:             "gemini_openai_compatible",
		Name:           "Gemini OpenAI Compatible",
		DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/",
		LegacyType:     51,
	}, func() adaptor.Adaptor {
		return &Adaptor{Adaptor: &openai.Adaptor{}}
	})
}
