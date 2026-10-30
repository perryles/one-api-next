package baichuan

import oa "github.com/perryles/one-api-next/relay/adaptor/openai"

type Adaptor struct {
	*oa.Adaptor
}

func (a *Adaptor) NeedsRequestBodyConversion() bool {
	return true
}

var ModelList = []string{
	"Baichuan2-Turbo",
	"Baichuan2-Turbo-192k",
	"Baichuan-Text-Embedding",
}

func (a *Adaptor) GetModelList() []string  { return ModelList }
