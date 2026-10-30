package lingyiwanwu

import oa "github.com/perryles/one-api-next/relay/adaptor/openai"

type Adaptor struct {
	*oa.Adaptor
}

var ModelList = []string{
	"yi-34b-chat-0205",
	"yi-34b-chat-200k",
	"yi-vl-plus",
}

func (a *Adaptor) GetModelList() []string  { return ModelList }
