package ratio

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
)

// 上下文分档计价（context-tiered pricing），机制移植自 omnirelays（见 docs/UPSTREAM.md）。
// 适用场景：同一模型按输入长度分档定价（如 Gemini 2.5 Pro >200k tokens 档）。
// 档位由输入 tokens 决定，作用于整个请求；倍率沿用现有 quota/token 刻度，
// 不含分组倍率与折扣。MaxInputTokens 含端点，0 表示最后一档（不限）。

// ContextPriceTier applies to the WHOLE request, selected by input tokens only.
// MaxInputTokens is inclusive; zero denotes the final, unbounded tier.
type ContextPriceTier struct {
	Name            string  `json:"name"`
	MaxInputTokens  int     `json:"max_input_tokens"`
	PromptRatio     float64 `json:"prompt_ratio"`
	CompletionRatio float64 `json:"completion_ratio"`
}

var contextPrices = struct {
	sync.RWMutex
	values map[string][]ContextPriceTier
}{values: map[string][]ContextPriceTier{}}

func ParseContextPriceTiers(raw string) (map[string][]ContextPriceTier, error) {
	var values map[string][]ContextPriceTier
	if err := json.Unmarshal([]byte(raw), &values); err != nil || values == nil {
		return nil, fmt.Errorf("上下文分档必须是模型名称到档位数组的 JSON 对象")
	}
	var fields map[string][]map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, fmt.Errorf("上下文档位必须是包含名称、输入上限和两种倍率的对象")
	}
	for model, tiers := range values {
		if strings.TrimSpace(model) == "" || len(model) > 256 || tiers == nil || len(tiers) > 16 || len(tiers) == 1 {
			return nil, fmt.Errorf("%s：请提供至少两档价格，或用空数组关闭分档", model)
		}
		last, names := 0, map[string]bool{}
		for i, tier := range tiers {
			for _, key := range []string{"name", "max_input_tokens", "prompt_ratio", "completion_ratio"} {
				value := fields[model][i][key]
				if len(value) == 0 || string(value) == "null" {
					return nil, fmt.Errorf("%s：档位缺少 %s", model, key)
				}
			}
			if strings.TrimSpace(tier.Name) == "" || len(tier.Name) > 128 || names[tier.Name] ||
				tier.PromptRatio <= 0 || math.IsNaN(tier.PromptRatio) || math.IsInf(tier.PromptRatio, 0) ||
				tier.CompletionRatio < 0 || math.IsNaN(tier.CompletionRatio) || math.IsInf(tier.CompletionRatio, 0) ||
				math.IsInf(tier.PromptRatio*tier.CompletionRatio, 0) {
				return nil, fmt.Errorf("%s：档位名称须唯一，输入倍率须大于零，补全倍率须为非负有限数", model)
			}
			if (i == len(tiers)-1 && tier.MaxInputTokens != 0) ||
				(i < len(tiers)-1 && (tier.MaxInputTokens <= last || tier.MaxInputTokens > 9007199254740991)) {
				return nil, fmt.Errorf("%s：输入长度上限须递增，最后一档上限须为 0（不限）", model)
			}
			names[tier.Name], last = true, tier.MaxInputTokens
		}
	}
	return values, nil
}

func UpdateContextPriceTiers(raw string) error {
	values, err := ParseContextPriceTiers(raw)
	if err != nil {
		return err
	}
	contextPrices.Lock()
	contextPrices.values = values
	contextPrices.Unlock()
	return nil
}

func ContextPriceTiersJSON() string {
	contextPrices.RLock()
	defer contextPrices.RUnlock()
	raw, _ := json.Marshal(contextPrices.values)
	return string(raw)
}

// An explicit empty type-specific entry disables a generic model's tiers.
// Return a copy so in-flight requests retain an immutable pricing snapshot.
func GetContextPriceTiers(model string, channelType int) []ContextPriceTier {
	contextPrices.RLock()
	defer contextPrices.RUnlock()
	tiers, found := contextPrices.values[fmt.Sprintf("%s(%d)", model, channelType)]
	if !found {
		tiers = contextPrices.values[model]
	}
	return append([]ContextPriceTier(nil), tiers...)
}

func SelectContextPriceTier(tiers []ContextPriceTier, inputTokens int) (ContextPriceTier, bool) {
	for _, tier := range tiers {
		if tier.MaxInputTokens == 0 || inputTokens <= tier.MaxInputTokens {
			return tier, true
		}
	}
	return ContextPriceTier{}, false
}
