package ratio

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
)

const (
	USD2RMB   = 7
	USD       = 500 // $0.002 = 1 -> $1 = 500
	MILLI_USD = 1.0 / 1000 * USD
	RMB       = USD / USD2RMB
)

var modelRatioLock sync.RWMutex
var completionRatioLock sync.RWMutex

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// https://openai.com/pricing
	"gpt-4":                      15,
	"gpt-4-0314":                 15,
	"gpt-4-0613":                 15,
	"gpt-4-32k":                  30,
	"gpt-4-32k-0314":             30,
	"gpt-4-32k-0613":             30,
	"gpt-4-1106-preview":         5,     // $0.01 / 1K tokens
	"gpt-4-0125-preview":         5,     // $0.01 / 1K tokens
	"gpt-4-turbo-preview":        5,     // $0.01 / 1K tokens
	"gpt-4-turbo":                5,     // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09":     5,     // $0.01 / 1K tokens
	"gpt-4o":                     2.5,   // $0.005 / 1K tokens
	"chatgpt-4o-latest":          2.5,   // $0.005 / 1K tokens
	"gpt-4o-2024-05-13":          2.5,   // $0.005 / 1K tokens
	"gpt-4o-2024-08-06":          1.25,  // $0.0025 / 1K tokens
	"gpt-4o-2024-11-20":          1.25,  // $0.0025 / 1K tokens
	"gpt-4o-mini":                0.075, // $0.00015 / 1K tokens
	"gpt-4o-mini-2024-07-18":     0.075, // $0.00015 / 1K tokens
	"gpt-4-vision-preview":       5,     // $0.01 / 1K tokens
	"gpt-3.5-turbo":              0.25,  // $0.0005 / 1K tokens
	"gpt-3.5-turbo-0301":         0.75,
	"gpt-3.5-turbo-0613":         0.75,
	"gpt-3.5-turbo-16k":          1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":     1.5,
	"gpt-3.5-turbo-instruct":     0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":         0.5,  // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":         0.25, // $0.0005 / 1K tokens
	"o1":                         7.5,  // $15.00 / 1M input tokens
	"o1-2024-12-17":              7.5,
	"o1-preview":                 7.5, // $15.00 / 1M input tokens
	"o1-preview-2024-09-12":      7.5,
	"o1-mini":                    1.5, // $3.00 / 1M input tokens
	"o1-mini-2024-09-12":         1.5,
	"o3-mini":                    1.5, // $3.00 / 1M input tokens
	"o3-mini-2025-01-31":         1.5,
	"davinci-002":                1,   // $0.002 / 1K tokens
	"babbage-002":                0.2, // $0.0004 / 1K tokens
	"text-ada-001":               0.2,
	"text-babbage-001":           0.25,
	"text-curie-001":             1,
	"text-davinci-002":           10,
	"text-davinci-003":           10,
	"text-davinci-edit-001":      10,
	"code-davinci-edit-001":      10,
	"whisper-1":                  15,  // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
	"tts-1":                      7.5, // $0.015 / 1K characters
	"tts-1-1106":                 7.5,
	"tts-1-hd":                   15, // $0.030 / 1K characters
	"tts-1-hd-1106":              15,
	"davinci":                    10,
	"curie":                      10,
	"babbage":                    10,
	"ada":                        10,
	"text-embedding-ada-002":     0.05,
	"text-embedding-3-small":     0.01,
	"text-embedding-3-large":     0.065,
	"text-search-ada-doc-001":    10,
	"text-moderation-stable":     0.1,
	"text-moderation-latest":     0.1,
	"omni-moderation-latest":     0.1, // ported from omnirelays: moderation endpoint is free; placeholder avoids zero-ratio ambiguity
	"omni-moderation-2024-09-26": 0.1,
	"dall-e-2":                   0.02 * USD, // $0.016 - $0.020 / image
	"dall-e-3":                   0.04 * USD, // $0.040 - $0.120 / image
	// https://docs.anthropic.com/en/docs/about-claude/models
	"claude-instant-1.2":         0.8 / 1000 * USD,
	"claude-2.0":                 8.0 / 1000 * USD,
	"claude-2.1":                 8.0 / 1000 * USD,
	"claude-3-haiku-20240307":    0.25 / 1000 * USD,
	"claude-3-5-haiku-20241022":  1.0 / 1000 * USD,
	"claude-3-5-haiku-latest":    1.0 / 1000 * USD,
	"claude-3-sonnet-20240229":   3.0 / 1000 * USD,
	"claude-3-5-sonnet-20240620": 3.0 / 1000 * USD,
	"claude-3-5-sonnet-20241022": 3.0 / 1000 * USD,
	"claude-3-5-sonnet-latest":   3.0 / 1000 * USD,
	"claude-3-opus-20240229":     15.0 / 1000 * USD,
	// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/hlrk4akp7
	"ERNIE-4.0-8K":       0.120 * RMB,
	"ERNIE-3.5-8K":       0.012 * RMB,
	"ERNIE-3.5-8K-0205":  0.024 * RMB,
	"ERNIE-3.5-8K-1222":  0.012 * RMB,
	"ERNIE-Bot-8K":       0.024 * RMB,
	"ERNIE-3.5-4K-0205":  0.012 * RMB,
	"ERNIE-Speed-8K":     0.004 * RMB,
	"ERNIE-Speed-128K":   0.004 * RMB,
	"ERNIE-Lite-8K-0922": 0.008 * RMB,
	"ERNIE-Lite-8K-0308": 0.003 * RMB,
	"ERNIE-Tiny-8K":      0.001 * RMB,
	"BLOOMZ-7B":          0.004 * RMB,
	"Embedding-V1":       0.002 * RMB,
	"bge-large-zh":       0.002 * RMB,
	"bge-large-en":       0.002 * RMB,
	"tao-8k":             0.002 * RMB,
	// https://ai.google.dev/pricing
	// https://cloud.google.com/vertex-ai/generative-ai/pricing
	// "gemma-2-2b-it":                       0,
	// "gemma-2-9b-it":                       0,
	// "gemma-2-27b-it":                      0,
	"gemini-pro":                          0.25 * MILLI_USD, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-1.0-pro":                      0.125 * MILLI_USD,
	"gemini-1.5-pro":                      1.25 * MILLI_USD,
	"gemini-1.5-pro-001":                  1.25 * MILLI_USD,
	"gemini-1.5-pro-experimental":         1.25 * MILLI_USD,
	"gemini-1.5-flash":                    0.075 * MILLI_USD,
	"gemini-1.5-flash-001":                0.075 * MILLI_USD,
	"gemini-1.5-flash-8b":                 0.0375 * MILLI_USD,
	"gemini-2.0-flash-exp":                0.075 * MILLI_USD,
	"gemini-2.0-flash":                    0.15 * MILLI_USD,
	"gemini-2.0-flash-001":                0.15 * MILLI_USD,
	"gemini-2.0-flash-lite-preview-02-05": 0.075 * MILLI_USD,
	"gemini-2.0-flash-thinking-exp-01-21": 0.075 * MILLI_USD,
	"gemini-2.0-pro-exp-02-05":            1.25 * MILLI_USD,
	"aqa":                                 1,
	// https://open.bigmodel.cn/pricing
	"glm-zero-preview": 0.01 * RMB,
	"glm-4-plus":       0.05 * RMB,
	"glm-4-0520":       0.1 * RMB,
	"glm-4-airx":       0.01 * RMB,
	"glm-4-air":        0.0005 * RMB,
	"glm-4-long":       0.001 * RMB,
	"glm-4-flashx":     0.0001 * RMB,
	"glm-4-flash":      0,
	"glm-4":            0.1 * RMB,   // deprecated model, available until 2025/06
	"glm-3-turbo":      0.001 * RMB, // deprecated model, available until 2025/06
	"glm-4v-plus":      0.004 * RMB,
	"glm-4v":           0.05 * RMB,
	"glm-4v-flash":     0,
	"cogview-3-plus":   0.06 * RMB,
	"cogview-3":        0.1 * RMB,
	"cogview-3-flash":  0,
	"cogviewx":         0.5 * RMB,
	"cogviewx-flash":   0,
	"charglm-4":        0.001 * RMB,
	"emohaa":           0.015 * RMB,
	"codegeex-4":       0.0001 * RMB,
	"embedding-2":      0.0005 * RMB,
	"embedding-3":      0.0005 * RMB,
	// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	"qwen-turbo":                    0.0003 * RMB,
	"qwen-turbo-latest":             0.0003 * RMB,
	"qwen-plus":                     0.0008 * RMB,
	"qwen-plus-latest":              0.0008 * RMB,
	"qwen-max":                      0.0024 * RMB,
	"qwen-max-latest":               0.0024 * RMB,
	"qwen-max-longcontext":          0.0005 * RMB,
	"qwen-vl-max":                   0.003 * RMB,
	"qwen-vl-max-latest":            0.003 * RMB,
	"qwen-vl-plus":                  0.0015 * RMB,
	"qwen-vl-plus-latest":           0.0015 * RMB,
	"qwen-vl-ocr":                   0.005 * RMB,
	"qwen-vl-ocr-latest":            0.005 * RMB,
	"qwen-audio-turbo":              1.4286,
	"qwen-math-plus":                0.004 * RMB,
	"qwen-math-plus-latest":         0.004 * RMB,
	"qwen-math-turbo":               0.002 * RMB,
	"qwen-math-turbo-latest":        0.002 * RMB,
	"qwen-coder-plus":               0.0035 * RMB,
	"qwen-coder-plus-latest":        0.0035 * RMB,
	"qwen-coder-turbo":              0.002 * RMB,
	"qwen-coder-turbo-latest":       0.002 * RMB,
	"qwen-mt-plus":                  0.015 * RMB,
	"qwen-mt-turbo":                 0.001 * RMB,
	"qwq-32b-preview":               0.002 * RMB,
	"qwen2.5-72b-instruct":          0.004 * RMB,
	"qwen2.5-32b-instruct":          0.03 * RMB,
	"qwen2.5-14b-instruct":          0.001 * RMB,
	"qwen2.5-7b-instruct":           0.0005 * RMB,
	"qwen2.5-3b-instruct":           0.006 * RMB,
	"qwen2.5-1.5b-instruct":         0.0003 * RMB,
	"qwen2.5-0.5b-instruct":         0.0003 * RMB,
	"qwen2-72b-instruct":            0.004 * RMB,
	"qwen2-57b-a14b-instruct":       0.0035 * RMB,
	"qwen2-7b-instruct":             0.001 * RMB,
	"qwen2-1.5b-instruct":           0.001 * RMB,
	"qwen2-0.5b-instruct":           0.001 * RMB,
	"qwen1.5-110b-chat":             0.007 * RMB,
	"qwen1.5-72b-chat":              0.005 * RMB,
	"qwen1.5-32b-chat":              0.0035 * RMB,
	"qwen1.5-14b-chat":              0.002 * RMB,
	"qwen1.5-7b-chat":               0.001 * RMB,
	"qwen1.5-1.8b-chat":             0.001 * RMB,
	"qwen1.5-0.5b-chat":             0.001 * RMB,
	"qwen-72b-chat":                 0.02 * RMB,
	"qwen-14b-chat":                 0.008 * RMB,
	"qwen-7b-chat":                  0.006 * RMB,
	"qwen-1.8b-chat":                0.006 * RMB,
	"qwen-1.8b-longcontext-chat":    0.006 * RMB,
	"qvq-72b-preview":               0.012 * RMB,
	"qwen2.5-vl-72b-instruct":       0.016 * RMB,
	"qwen2.5-vl-7b-instruct":        0.002 * RMB,
	"qwen2.5-vl-3b-instruct":        0.0012 * RMB,
	"qwen2-vl-7b-instruct":          0.016 * RMB,
	"qwen2-vl-2b-instruct":          0.002 * RMB,
	"qwen-vl-v1":                    0.002 * RMB,
	"qwen-vl-chat-v1":               0.002 * RMB,
	"qwen2-audio-instruct":          0.002 * RMB,
	"qwen-audio-chat":               0.002 * RMB,
	"qwen2.5-math-72b-instruct":     0.004 * RMB,
	"qwen2.5-math-7b-instruct":      0.001 * RMB,
	"qwen2.5-math-1.5b-instruct":    0.001 * RMB,
	"qwen2-math-72b-instruct":       0.004 * RMB,
	"qwen2-math-7b-instruct":        0.001 * RMB,
	"qwen2-math-1.5b-instruct":      0.001 * RMB,
	"qwen2.5-coder-32b-instruct":    0.002 * RMB,
	"qwen2.5-coder-14b-instruct":    0.002 * RMB,
	"qwen2.5-coder-7b-instruct":     0.001 * RMB,
	"qwen2.5-coder-3b-instruct":     0.001 * RMB,
	"qwen2.5-coder-1.5b-instruct":   0.001 * RMB,
	"qwen2.5-coder-0.5b-instruct":   0.001 * RMB,
	"text-embedding-v1":             0.0007 * RMB, // ￥0.0007 / 1k tokens
	"text-embedding-v3":             0.0007 * RMB,
	"text-embedding-v2":             0.0007 * RMB,
	"text-embedding-async-v2":       0.0007 * RMB,
	"text-embedding-async-v1":       0.0007 * RMB,
	"ali-stable-diffusion-xl":       8.00,
	"ali-stable-diffusion-v1.5":     8.00,
	"wanx-v1":                       8.00,
	"deepseek-r1":                   0.002 * RMB,
	"deepseek-v3":                   0.001 * RMB,
	"deepseek-r1-distill-qwen-1.5b": 0.001 * RMB,
	"deepseek-r1-distill-qwen-7b":   0.0005 * RMB,
	"deepseek-r1-distill-qwen-14b":  0.001 * RMB,
	"deepseek-r1-distill-qwen-32b":  0.002 * RMB,
	"deepseek-r1-distill-llama-8b":  0.0005 * RMB,
	"deepseek-r1-distill-llama-70b": 0.004 * RMB,
	"SparkDesk":                     1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v1.1":                1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":                1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":                1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1-128K":           1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":                1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5-32K":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":                1.2858, // ￥0.018 / 1k tokens
	"360GPT_S2_V9":                  0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":         0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":               0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1":     0.0715, // ¥0.001 / 1k tokens
	// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	"hunyuan-turbo":             0.015 * RMB,
	"hunyuan-large":             0.004 * RMB,
	"hunyuan-large-longcontext": 0.006 * RMB,
	"hunyuan-standard":          0.0008 * RMB,
	"hunyuan-standard-256K":     0.0005 * RMB,
	"hunyuan-translation-lite":  0.005 * RMB,
	"hunyuan-role":              0.004 * RMB,
	"hunyuan-functioncall":      0.004 * RMB,
	"hunyuan-code":              0.004 * RMB,
	"hunyuan-turbo-vision":      0.08 * RMB,
	"hunyuan-vision":            0.018 * RMB,
	"hunyuan-embedding":         0.0007 * RMB,
	// https://platform.moonshot.cn/pricing
	"moonshot-v1-8k":   0.012 * RMB,
	"moonshot-v1-32k":  0.024 * RMB,
	"moonshot-v1-128k": 0.06 * RMB,
	// https://platform.baichuan-ai.com/price
	"Baichuan2-Turbo":      0.008 * RMB,
	"Baichuan2-Turbo-192k": 0.016 * RMB,
	"Baichuan2-53B":        0.02 * RMB,
	// https://api.minimax.chat/document/price
	"abab6.5-chat":  0.03 * RMB,
	"abab6.5s-chat": 0.01 * RMB,
	"abab6-chat":    0.1 * RMB,
	"abab5.5-chat":  0.015 * RMB,
	"abab5.5s-chat": 0.005 * RMB,
	// https://docs.mistral.ai/platform/pricing/
	"open-mistral-7b":       0.25 / 1000 * USD,
	"open-mixtral-8x7b":     0.7 / 1000 * USD,
	"mistral-small-latest":  2.0 / 1000 * USD,
	"mistral-medium-latest": 2.7 / 1000 * USD,
	"mistral-large-latest":  8.0 / 1000 * USD,
	"mistral-embed":         0.1 / 1000 * USD,
	// https://wow.groq.com/#:~:text=inquiries%C2%A0here.-,Model,-Current%20Speed
	"gemma-7b-it":                           0.07 / 1000000 * USD,
	"gemma2-9b-it":                          0.20 / 1000000 * USD,
	"llama-3.1-70b-versatile":               0.59 / 1000000 * USD,
	"llama-3.1-8b-instant":                  0.05 / 1000000 * USD,
	"llama-3.2-11b-text-preview":            0.05 / 1000000 * USD,
	"llama-3.2-11b-vision-preview":          0.05 / 1000000 * USD,
	"llama-3.2-1b-preview":                  0.05 / 1000000 * USD,
	"llama-3.2-3b-preview":                  0.05 / 1000000 * USD,
	"llama-3.2-90b-text-preview":            0.59 / 1000000 * USD,
	"llama-guard-3-8b":                      0.05 / 1000000 * USD,
	"llama3-70b-8192":                       0.59 / 1000000 * USD,
	"llama3-8b-8192":                        0.05 / 1000000 * USD,
	"llama3-groq-70b-8192-tool-use-preview": 0.89 / 1000000 * USD,
	"llama3-groq-8b-8192-tool-use-preview":  0.19 / 1000000 * USD,
	"mixtral-8x7b-32768":                    0.24 / 1000000 * USD,

	// https://platform.lingyiwanwu.com/docs#-计费单元
	"yi-34b-chat-0205": 2.5 / 1000 * RMB,
	"yi-34b-chat-200k": 12.0 / 1000 * RMB,
	"yi-vl-plus":       6.0 / 1000 * RMB,
	// https://platform.stepfun.com/docs/pricing/details
	"step-1-8k":    0.005 / 1000 * RMB,
	"step-1-32k":   0.015 / 1000 * RMB,
	"step-1-128k":  0.040 / 1000 * RMB,
	"step-1-256k":  0.095 / 1000 * RMB,
	"step-1-flash": 0.001 / 1000 * RMB,
	"step-2-16k":   0.038 / 1000 * RMB,
	"step-1v-8k":   0.005 / 1000 * RMB,
	"step-1v-32k":  0.015 / 1000 * RMB,
	// aws llama3 https://aws.amazon.com/cn/bedrock/pricing/
	"llama3-8b-8192(33)":  0.0003 / 0.002,  // $0.0003 / 1K tokens
	"llama3-70b-8192(33)": 0.00265 / 0.002, // $0.00265 / 1K tokens
	// https://cohere.com/pricing
	"command":               0.5,
	"command-nightly":       0.5,
	"command-light":         0.5,
	"command-light-nightly": 0.5,
	"command-r":             0.5 / 1000 * USD,
	"command-r-plus":        3.0 / 1000 * USD,
	// https://platform.deepseek.com/api-docs/pricing/
	"deepseek-chat":     0.14 * MILLI_USD,
	"deepseek-reasoner": 0.55 * MILLI_USD,
	// https://www.deepl.com/pro?cta=header-prices
	"deepl-zh": 25.0 / 1000 * USD,
	"deepl-en": 25.0 / 1000 * USD,
	"deepl-ja": 25.0 / 1000 * USD,
	// https://console.x.ai/
	"grok-beta": 5.0 / 1000 * USD,
	// replicate charges based on the number of generated images
	// https://replicate.com/pricing
	// replicate chat models
	//https://openrouter.ai/models
	// ===================================================================
	// one-api-next price refresh 2026-10-31
	// Source: first-party official prices via models.dev open dataset;
	// ratio = $/1M input / 2. DeepSeek uses off-peak list price (peak = 2x).
	// The legacy openrouter/* snapshot (2025-02) is fully replaced here.
	// ===================================================================
	"deepseek-flash":                    0.075,    // $0.15 in / $0.6 out per 1M, deepseek (off-peak basis, peak=2x)
	"deepseek-v4-flash":                 0.075,    // $0.15 in / $0.6 out per 1M, deepseek (off-peak basis, peak=2x)
	"deepseek-v4-flash-vision-exp":      0.075,    // $0.15 in / $0.6 out per 1M, deepseek (off-peak basis, peak=2x)
	"deepseek-v4-pro":                   0.33,     // $0.66 in / $1.98 out per 1M, deepseek (off-peak basis, peak=2x)
	"glm-4.5":                           0.3,      // $0.6 in / $2.2 out per 1M, zhipuai
	"glm-4.5-air":                       0.05976,  // $0.11952 in / $0.89637 out per 1M, zhipuai
	"glm-4.5-flash":                     0,        // free, zhipuai
	"glm-4.5v":                          0.149395, // $0.29879 in / $0.89637 out per 1M, zhipuai
	"glm-4.6":                           0.3,      // $0.6 in / $2.2 out per 1M, zhipuai
	"glm-4.6v":                          0.0747,   // $0.1494 in / $0.44819 out per 1M, zhipuai
	"glm-4.6v-flash":                    0,        // free, zhipuai
	"glm-4.7":                           0.224095, // $0.44819 in / $2.09153 out per 1M, zhipuai
	"glm-4.7-flash":                     0,        // free, zhipuai
	"glm-4.7-flashx":                    0.03735,  // $0.0747 in / $0.44819 out per 1M, zhipuai
	"glm-5":                             0.29879,  // $0.59758 in / $2.68911 out per 1M, zhipuai
	"glm-5.1":                           0.448185, // $0.89637 in / $3.58548 out per 1M, zhipuai
	"glm-5.2":                           0.59758,  // $1.19516 in / $4.18306 out per 1M, zhipuai
	"glm-5.3":                           0.59758,  // $1.19516 in / $4.18306 out per 1M, zhipuai
	"glm-5.3-flash":                     0.05976,  // $0.11952 in / $0.41831 out per 1M, zhipuai
	"glm-5.3-flashx":                    0.149395, // $0.29879 in / $1.04577 out per 1M, zhipuai
	"glm-5v-turbo":                      0.37349,  // $0.74698 in / $3.28669 out per 1M, zhipuai
	"kimi-k2.6":                         0.48555,  // $0.9711 in / $4.0337 out per 1M, moonshotai-cn
	"kimi-k2.7-code":                    0.48555,  // $0.9711 in / $4.0337 out per 1M, moonshotai-cn
	"kimi-k2.7-code-highspeed":          0.97105,  // $1.9421 in / $8.0673 out per 1M, moonshotai-cn
	"kimi-k3":                           1.49395,  // $2.9879 in / $14.9395 out per 1M, moonshotai-cn
	"kimi-k3.1-pro":                     0.97185,  // $1.9437 in / $10.1671 out per 1M, moonshotai-cn
	"kimi-k3.1-pro-highspeed":           1.9437,   // $3.8874 in / $20.3342 out per 1M, moonshotai-cn
	"MiniMax-M2":                        0.15685,  // $0.3137 in / $1.2549 out per 1M, minimax-cn
	"MiniMax-M2.1":                      0.15685,  // $0.3137 in / $1.2549 out per 1M, minimax-cn
	"MiniMax-M2.5":                      0.15685,  // $0.3137 in / $1.2549 out per 1M, minimax-cn
	"MiniMax-M2.5-highspeed":            0.31375,  // $0.6275 in / $2.5098 out per 1M, minimax-cn
	"MiniMax-M2.7":                      0.15685,  // $0.3137 in / $1.2549 out per 1M, minimax-cn
	"MiniMax-M2.7-highspeed":            0.31375,  // $0.6275 in / $2.5098 out per 1M, minimax-cn
	"MiniMax-M3":                        0.15685,  // $0.3137 in / $1.2549 out per 1M, minimax-cn
	"MiniMax-M3.1-Flash":                0.0747,   // $0.1494 in / $0.5976 out per 1M, minimax-cn
	"grok-4.20-0309-non-reasoning":      0.625,    // $1.25 in / $2.5 out per 1M, xai
	"grok-4.20-0309-reasoning":          0.625,    // $1.25 in / $2.5 out per 1M, xai
	"grok-4.20-multi-agent-0309":        0.625,    // $1.25 in / $2.5 out per 1M, xai
	"grok-4.3":                          0.625,    // $1.25 in / $2.5 out per 1M, xai
	"grok-4.5":                          1,        // $2 in / $6 out per 1M, xai
	"grok-4.6":                          1,        // $2 in / $6 out per 1M, xai
	"grok-4.7":                          1,        // $2 in / $6 out per 1M, xai
	"grok-build-0.1":                    0.5,      // $1 in / $2 out per 1M, xai
	"gpt-4.1":                           1,        // $2 in / $8 out per 1M, openai
	"gpt-4.1-mini":                      0.2,      // $0.4 in / $1.6 out per 1M, openai
	"gpt-4.1-nano":                      0.05,     // $0.1 in / $0.4 out per 1M, openai
	"gpt-5":                             0.625,    // $1.25 in / $10 out per 1M, openai
	"gpt-5-mini":                        0.125,    // $0.25 in / $2 out per 1M, openai
	"gpt-5-nano":                        0.025,    // $0.05 in / $0.4 out per 1M, openai
	"gpt-5-pro":                         7.5,      // $15 in / $120 out per 1M, openai
	"gpt-5.1":                           0.625,    // $1.25 in / $10 out per 1M, openai
	"gpt-5.2":                           0.875,    // $1.75 in / $14 out per 1M, openai
	"gpt-5.2-pro":                       10.5,     // $21 in / $168 out per 1M, openai
	"gpt-5.3-codex":                     0.875,    // $1.75 in / $14 out per 1M, openai
	"gpt-5.3-codex-spark":               0.875,    // $1.75 in / $14 out per 1M, openai
	"gpt-5.4":                           1.25,     // $2.5 in / $15 out per 1M, openai
	"gpt-5.4-mini":                      0.375,    // $0.75 in / $4.5 out per 1M, openai
	"gpt-5.4-nano":                      0.1,      // $0.2 in / $1.25 out per 1M, openai
	"gpt-5.4-pro":                       15,       // $30 in / $180 out per 1M, openai
	"gpt-5.5":                           2.5,      // $5 in / $30 out per 1M, openai
	"gpt-5.5-pro":                       15,       // $30 in / $180 out per 1M, openai
	"gpt-5.6":                           2,        // $4 in / $20 out per 1M, openai
	"gpt-5.6-luna":                      0.1,      // $0.2 in / $1.2 out per 1M, openai
	"gpt-5.6-sol":                       2,        // $4 in / $20 out per 1M, openai
	"gpt-5.6-terra":                     1,        // $2 in / $12 out per 1M, openai
	"gpt-6-astra":                       5,        // $10 in / $50 out per 1M, openai
	"gpt-6-luna":                        0.05,     // $0.1 in / $0.5 out per 1M, openai
	"gpt-6-sol":                         1,        // $2 in / $10 out per 1M, openai
	"gpt-6.1-sol":                       1,        // $2 in / $10 out per 1M, openai
	"gpt-6.1-terra":                     0.5,      // $1 in / $5 out per 1M, openai
	"gpt-daybreak-blue-latest":          2,        // $4 in / $20 out per 1M, openai
	"gpt-daybreak-red-latest":           6.25,     // $12.5 in / $75 out per 1M, openai
	"gpt-image-2":                       2.5,      // $5 in / $30 out per 1M, openai
	"gpt-realtime-2.1":                  2,        // $4 in / $24 out per 1M, openai
	"o1-pro":                            75,       // $150 in / $600 out per 1M, openai
	"o3":                                1,        // $2 in / $8 out per 1M, openai
	"o3-pro":                            10,       // $20 in / $80 out per 1M, openai
	"o4-mini":                           0.55,     // $1.1 in / $4.4 out per 1M, openai
	"claude-fable-5":                    5,        // $10 in / $50 out per 1M, anthropic
	"claude-fable-5-1":                  5,        // $10 in / $50 out per 1M, anthropic
	"claude-haiku-4-5":                  0.5,      // $1 in / $5 out per 1M, anthropic
	"claude-haiku-4-5-20251001":         0.5,      // $1 in / $5 out per 1M, anthropic
	"claude-haiku-5-5":                  0.05,     // $0.1 in / $0.5 out per 1M, anthropic
	"claude-opus-4-5":                   2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-4-5-20251101":          2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-4-6":                   2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-4-7":                   2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-4-8":                   2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-5":                     2.5,      // $5 in / $25 out per 1M, anthropic
	"claude-opus-5-5":                   2,        // $4 in / $20 out per 1M, anthropic
	"claude-sonnet-4-5":                 1.5,      // $3 in / $15 out per 1M, anthropic
	"claude-sonnet-4-5-20250929":        1.5,      // $3 in / $15 out per 1M, anthropic
	"claude-sonnet-4-6":                 1.5,      // $3 in / $15 out per 1M, anthropic
	"claude-sonnet-5":                   1,        // $2 in / $10 out per 1M, anthropic
	"claude-sonnet-5-5":                 1,        // $2 in / $10 out per 1M, anthropic
	"deep-research-max-preview-04-2026": 1,        // $2 in / $12 out per 1M, google
	"deep-research-preview-04-2026":     1,        // $2 in / $12 out per 1M, google
	"gemini-2.5-computer-use-preview-10-2025": 0.625,    // $1.25 in / $10 out per 1M, google
	"gemini-2.5-flash":                        0.15,     // $0.3 in / $2.5 out per 1M, google
	"gemini-2.5-flash-image":                  0.15,     // $0.3 in / $30 out per 1M, google
	"gemini-2.5-flash-lite":                   0.05,     // $0.1 in / $0.4 out per 1M, google
	"gemini-2.5-flash-preview-tts":            0.25,     // $0.5 in / $10 out per 1M, google
	"gemini-2.5-pro":                          0.625,    // $1.25 in / $10 out per 1M, google
	"gemini-2.5-pro-preview-tts":              0.5,      // $1 in / $20 out per 1M, google
	"gemini-3-flash-preview":                  0.25,     // $0.5 in / $3 out per 1M, google
	"gemini-3-pro-image":                      1,        // $2 in / $120 out per 1M, google
	"gemini-3.1-flash-image":                  0.25,     // $0.5 in / $60 out per 1M, google
	"gemini-3.1-flash-lite":                   0.125,    // $0.25 in / $1.5 out per 1M, google
	"gemini-3.1-flash-lite-image":             0.125,    // $0.25 in / $30 out per 1M, google
	"gemini-3.1-flash-lite-preview":           0.125,    // $0.25 in / $1.5 out per 1M, google
	"gemini-3.1-flash-live-preview":           0.375,    // $0.75 in / $4.5 out per 1M, google
	"gemini-3.1-flash-tts-preview":            0.5,      // $1 in / $20 out per 1M, google
	"gemini-3.1-pro-preview":                  1,        // $2 in / $12 out per 1M, google
	"gemini-3.1-pro-preview-customtools":      1,        // $2 in / $12 out per 1M, google
	"gemini-3.5-flash":                        0.75,     // $1.5 in / $9 out per 1M, google
	"gemini-3.5-flash-lite":                   0.15,     // $0.3 in / $2.5 out per 1M, google
	"gemini-3.5-live-translate-preview":       1.75,     // $3.5 in / $21 out per 1M, google
	"gemini-3.6-flash":                        0.375,    // $0.75 in / $3.75 out per 1M, google
	"gemini-3.7-flash":                        0.375,    // $0.75 in / $3.75 out per 1M, google
	"gemini-3.8-flash":                        0.375,    // $0.75 in / $3.75 out per 1M, google
	"gemini-4-argon":                          1,        // $2 in / $10 out per 1M, google
	"gemini-embedding-001":                    0.075,    // $0.15 in / $0 out per 1M, google
	"gemini-embedding-2":                      0.1,      // $0.2 in / $0 out per 1M, google
	"gemini-flash-latest":                     0.375,    // $0.75 in / $3.75 out per 1M, google
	"gemini-flash-lite-latest":                0.15,     // $0.3 in / $2.5 out per 1M, google
	"gemini-omni-flash-preview":               0.75,     // $1.5 in / $17.5 out per 1M, google
	"gemini-pro-latest":                       1,        // $2 in / $12 out per 1M, google
	"lyria-3-clip-preview":                    0,        // free, google
	"lyria-3-pro-preview":                     0,        // free, google
	"deepseek-r1-0528":                        0.287,    // $0.574 in / $2.294 out per 1M, alibaba-cn
	"deepseek-r1-distill-qwen-1-5b":           0,        // free, alibaba-cn
	"deepseek-v3-1":                           0.287,    // $0.574 in / $1.721 out per 1M, alibaba-cn
	"deepseek-v3-2-exp":                       0.1435,   // $0.287 in / $0.431 out per 1M, alibaba-cn
	"deepseek-v4.1-flash":                     0.14877,  // $0.29754 in / $1.19015 out per 1M, alibaba-cn
	"kimi-k2-thinking":                        0.287,    // $0.574 in / $2.294 out per 1M, alibaba-cn
	"kimi-k2.5":                               0.287,    // $0.574 in / $3.011 out per 1M, alibaba-cn
	"moonshot-kimi-k2-instruct":               0.287,    // $0.574 in / $2.294 out per 1M, alibaba-cn
	"qvq-max":                                 0.5735,   // $1.147 in / $4.588 out per 1M, alibaba-cn
	"qwen-deep-research":                      3.871,    // $7.742 in / $23.367 out per 1M, alibaba-cn
	"qwen-doc-turbo":                          0.0435,   // $0.087 in / $0.144 out per 1M, alibaba-cn
	"qwen-flash":                              0.011,    // $0.022 in / $0.216 out per 1M, alibaba-cn
	"qwen-long":                               0.036,    // $0.072 in / $0.287 out per 1M, alibaba-cn
	"qwen-omni-turbo":                         0.029,    // $0.058 in / $0.23 out per 1M, alibaba-cn
	"qwen-omni-turbo-realtime":                0.115,    // $0.23 in / $0.918 out per 1M, alibaba-cn
	"qwen-plus-character":                     0.0575,   // $0.115 in / $0.287 out per 1M, alibaba-cn
	"qwen2-5-omni-7b":                         0.0435,   // $0.087 in / $0.345 out per 1M, alibaba-cn
	"qwen3-14b":                               0.072,    // $0.144 in / $0.574 out per 1M, alibaba-cn
	"qwen3-235b-a22b":                         0.1435,   // $0.287 in / $1.147 out per 1M, alibaba-cn
	"qwen3-32b":                               0.1435,   // $0.287 in / $1.147 out per 1M, alibaba-cn
	"qwen3-8b":                                0.036,    // $0.072 in / $0.287 out per 1M, alibaba-cn
	"qwen3-coder-30b-a3b-instruct":            0.108,    // $0.216 in / $0.861 out per 1M, alibaba-cn
	"qwen3-coder-480b-a35b-instruct":          0.4305,   // $0.861 in / $3.441 out per 1M, alibaba-cn
	"qwen3-coder-flash":                       0.072,    // $0.144 in / $0.574 out per 1M, alibaba-cn
	"qwen3-coder-plus":                        0.287,    // $0.574 in / $2.296 out per 1M, alibaba-cn
	"qwen3-max":                               0.1795,   // $0.359 in / $1.434 out per 1M, alibaba-cn
	"qwen3-next-80b-a3b-instruct":             0.072,    // $0.144 in / $0.574 out per 1M, alibaba-cn
	"qwen3-next-80b-a3b-thinking":             0.072,    // $0.144 in / $1.434 out per 1M, alibaba-cn
	"qwen3-omni-flash":                        0.129,    // $0.258 in / $0.989 out per 1M, alibaba-cn
	"qwen3-omni-flash-realtime":               0.115,    // $0.23 in / $0.918 out per 1M, alibaba-cn
	"qwen3-vl-235b-a22b":                      0.143352, // $0.286705 in / $1.14682 out per 1M, alibaba-cn
	"qwen3-vl-30b-a3b":                        0.054,    // $0.108 in / $0.431 out per 1M, alibaba-cn
	"qwen3-vl-plus":                           0.071677, // $0.143353 in / $1.433525 out per 1M, alibaba-cn
	"qwen3.5-397b-a17b":                       0.086,    // $0.172 in / $1.032 out per 1M, alibaba-cn
	"qwen3.5-flash":                           0.0145,   // $0.029 in / $0.287 out per 1M, alibaba-cn
	"qwen3.5-plus":                            0.0575,   // $0.115 in / $0.688 out per 1M, alibaba-cn
	"qwen3.6-flash":                           0.0825,   // $0.165 in / $0.99 out per 1M, alibaba-cn
	"qwen3.6-max-preview":                     0.66,     // $1.32 in / $7.9 out per 1M, alibaba-cn
	"qwen3.6-plus":                            0.138,    // $0.276 in / $1.651 out per 1M, alibaba-cn
	"qwen3.7-flash":                           0.01481,  // $0.02962 in / $0.1185 out per 1M, alibaba-cn
	"qwen3.7-max":                             0.825,    // $1.65 in / $4.951 out per 1M, alibaba-cn
	"qwen3.7-plus":                            0.138,    // $0.276 in / $1.101 out per 1M, alibaba-cn
	"qwen3.8-flash":                           0.059375, // $0.11875 in / $0.40073 out per 1M, alibaba-cn
	"qwen3.8-max":                             0.88872,  // $1.77744 in / $5.33231 out per 1M, alibaba-cn
	"qwen3.8-omni-flash":                      0.0565,   // $0.113 in / $0.382 out per 1M, alibaba-cn
	"qwq-plus":                                0.115,    // $0.23 in / $0.574 out per 1M, alibaba-cn
	"tongyi-intent-detect-v3":                 0.029,    // $0.058 in / $0.144 out per 1M, alibaba-cn
	"deepseek-v4-flash-0731":                  0.22,     // $0.44 in / $1.32 out per 1M, alibaba
	"qwen-plus-character-ja":                  0.25,     // $0.5 in / $1.4 out per 1M, alibaba
	"qwen3-livetranslate-flash-realtime":      5,        // $10 in / $10 out per 1M, alibaba
	"qwen3.5-122b-a10b":                       0.2,      // $0.4 in / $3.2 out per 1M, alibaba
	"qwen3.5-27b":                             0.15,     // $0.3 in / $2.4 out per 1M, alibaba
	"qwen3.5-35b-a3b":                         0.125,    // $0.25 in / $2 out per 1M, alibaba
	"qwen3.6-27b":                             0.3,      // $0.6 in / $3.6 out per 1M, alibaba
	"qwen3.6-35b-a3b":                         0.1875,   // $0.375 in / $2.25 out per 1M, alibaba
	"deepseek-v4-flash-ga-260731":             0.22265,  // $0.4453 in / $1.3359 out per 1M, volcengine
	"deepseek-v4-pro-ga-260813":               0.66795,  // $1.3359 in / $4.00771 out per 1M, volcengine
	"doubao-seed-1-6-251015":                  0.059375, // $0.11875 in / $1.18747 out per 1M, volcengine
	"doubao-seed-1.6-251015":                  0.059375, // $0.11875 in / $1.18747 out per 1M, volcengine
	"doubao-seed-1-6-flash-250828":            0.011135, // $0.02227 in / $0.22265 out per 1M, volcengine
	"doubao-seed-1-6-vision-250815":           0.059375, // $0.11875 in / $1.18747 out per 1M, volcengine
	"doubao-seed-1.8-251228":                  0.059375, // $0.11875 in / $1.18747 out per 1M, volcengine
	"doubao-seed-1-8-251228":                  0.059375, // $0.11875 in / $1.18747 out per 1M, volcengine
	"doubao-seed-2-0-code-preview-260215":     0.237495, // $0.47499 in / $2.37494 out per 1M, volcengine
	"doubao-seed-2-0-lite-260428":             0.04453,  // $0.08906 in / $0.53436 out per 1M, volcengine
	"doubao-seed-2-0-mini-260428":             0.014845, // $0.02969 in / $0.29687 out per 1M, volcengine
	"doubao-seed-2-0-pro-260215":              0.237495, // $0.47499 in / $2.37494 out per 1M, volcengine
	"doubao-seed-2-1-pro-260628":              0.4453,   // $0.8906 in / $4.45301 out per 1M, volcengine
	"doubao-seed-2-1-turbo-260628":            0.22265,  // $0.4453 in / $2.22651 out per 1M, volcengine
	"doubao-seed-character-260628":            0.059375, // $0.11875 in / $0.29687 out per 1M, volcengine
	"doubao-seed-evolving":                    0.4453,   // $0.8906 in / $4.45301 out per 1M, volcengine
	"glm-5.2-260617":                          0.593735, // $1.18747 in / $4.15615 out per 1M, volcengine
	"glm-5-2-260617":                          0.593735, // $1.18747 in / $4.15615 out per 1M, volcengine
	"glm-5-3-flash-260828":                    0.059375, // $0.11875 in / $0.41563 out per 1M, volcengine
	"step-3.5-flash":                          0.052305, // $0.10461 in / $0.31384 out per 1M, stepfun
	"step-3.5-flash-2603":                     0.052305, // $0.10461 in / $0.31384 out per 1M, stepfun
	"step-3.7-flash":                          0.100875, // $0.20175 in / $1.21052 out per 1M, stepfun
	"step-5-preview":                          0.523065, // $1.04613 in / $2.98895 out per 1M, stepfun
	"codestral-2508":                          0.15,     // $0.3 in / $0.9 out per 1M, mistral
	"codestral-latest":                        0.15,     // $0.3 in / $0.9 out per 1M, mistral
	"deepseek-v4-1-flash-1017":                0.15,     // $0.3 in / $1.2 out per 1M, mistral
	"devstral-2512":                           0.2,      // $0.4 in / $2 out per 1M, mistral
	"devstral-latest":                         0.2,      // $0.4 in / $2 out per 1M, mistral
	"devstral-medium-latest":                  0.2,      // $0.4 in / $2 out per 1M, mistral
	"glm-5-2":                                 0.7,      // $1.4 in / $4.4 out per 1M, mistral
	"labs-devstral-small-2512":                0,        // free, mistral
	"labs-leanstral-1-5-1":                    0,        // free, mistral
	"magistral-medium-latest":                 1,        // $2 in / $5 out per 1M, mistral
	"ministral-14b-2512":                      0.1,      // $0.2 in / $0.2 out per 1M, mistral
	"ministral-3b-2512":                       0.05,     // $0.1 in / $0.1 out per 1M, mistral
	"ministral-3b-latest":                     0.05,     // $0.1 in / $0.1 out per 1M, mistral
	"ministral-8b-2512":                       0.075,    // $0.15 in / $0.15 out per 1M, mistral
	"ministral-8b-latest":                     0.075,    // $0.15 in / $0.15 out per 1M, mistral
	"mistral-large-2512":                      0.25,     // $0.5 in / $1.5 out per 1M, mistral
	"mistral-large-4":                         0.35,     // $0.7 in / $2.1 out per 1M, mistral
	"mistral-large-4-0-8":                     0.35,     // $0.7 in / $2.1 out per 1M, mistral
	"mistral-medium-2505":                     0.2,      // $0.4 in / $2 out per 1M, mistral
	"mistral-medium-2508":                     0.2,      // $0.4 in / $2 out per 1M, mistral
	"mistral-medium-2604":                     0.75,     // $1.5 in / $7.5 out per 1M, mistral
	"mistral-small-2506":                      0.05,     // $0.1 in / $0.3 out per 1M, mistral
	"mistral-small-2603":                      0.075,    // $0.15 in / $0.6 out per 1M, mistral
	"open-mistral-nemo":                       0.075,    // $0.15 in / $0.15 out per 1M, mistral
	"pixtral-12b":                             0.075,    // $0.15 in / $0.15 out per 1M, mistral
	"voxtral-small-2507":                      0.05,     // $0.1 in / $0.4 out per 1M, mistral
	"voxtral-small-latest":                    0.05,     // $0.1 in / $0.4 out per 1M, mistral
	"zai-glm-5-2":                             0.7,      // $1.4 in / $4.4 out per 1M, mistral
	"zai-glm-5-3":                             0.7,      // $1.4 in / $4.4 out per 1M, mistral
	"zai-glm-5-3-flash":                       0.075,    // $0.15 in / $0.5 out per 1M, mistral
	"c4ai-aya-expanse-32b":                    0.25,     // $0.5 in / $1.5 out per 1M, cohere
	"command-a-03-2025":                       1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-a-plus-05-2026":                  1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-a-reasoning-08-2025":             1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-a-translate-08-2025":             1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-a-vision-07-2025":                1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-r-08-2024":                       0.075,    // $0.15 in / $0.6 out per 1M, cohere
	"command-r-plus-08-2024":                  1.25,     // $2.5 in / $10 out per 1M, cohere
	"command-r7b-12-2024":                     0.01875,  // $0.0375 in / $0.15 out per 1M, cohere
	"command-r7b-arabic-02-2025":              0.01875,  // $0.0375 in / $0.15 out per 1M, cohere
	"north-mini-code-1-0":                     0,        // free, cohere
	"north-small-translate-09-2026":           0,        // free, cohere
	"sonar":                                   0.5,      // $1 in / $1 out per 1M, perplexity
	"sonar-deep-research":                     1,        // $2 in / $8 out per 1M, perplexity
	"sonar-pro":                               1.5,      // $3 in / $15 out per 1M, perplexity
	"sonar-reasoning-pro":                     1,        // $2 in / $8 out per 1M, perplexity
	"allam-2-7b":                              0,        // free, groq
	"llama-3.3-70b-versatile":                 0.295,    // $0.59 in / $0.79 out per 1M, groq
	"aion-labs/aion-2.0":                      0.4,      // $0.8 in / $1.6 out per 1M, openrouter
	"aion-labs/aion-3.0":                      1.5,      // $3 in / $6 out per 1M, openrouter
	"aion-labs/aion-3.0-mini":                 0.35,     // $0.7 in / $1.4 out per 1M, openrouter
	"aion-labs/aion-3.5":                      1.5,      // $3 in / $6 out per 1M, openrouter
	"aion-labs/aion-3.5-mini":                 0.35,     // $0.7 in / $1.4 out per 1M, openrouter
	"aion-labs/aion-rp-llama-3.1-8b":          0.4,      // $0.8 in / $1.6 out per 1M, openrouter
	"amazon/nova-2-lite-v1":                   0.15,     // $0.3 in / $2.5 out per 1M, openrouter
	"amazon/nova-lite-v1":                     0.03,     // $0.06 in / $0.24 out per 1M, openrouter
	"amazon/nova-micro-v1":                    0.0175,   // $0.035 in / $0.14 out per 1M, openrouter
	"amazon/nova-premier-v1":                  1.25,     // $2.5 in / $12.5 out per 1M, openrouter
	"amazon/nova-pro-v1":                      0.4,      // $0.8 in / $3.2 out per 1M, openrouter
	"anthracite-org/magnum-v4-72b":            1.25,     // $2.5 in / $5 out per 1M, openrouter
	"anthropic/claude-fable-5":                5,        // $10 in / $50 out per 1M, openrouter
	"anthropic/claude-fable-5.1":              5,        // $10 in / $50 out per 1M, openrouter
	"anthropic/claude-haiku-4.5":              0.5,      // $1 in / $5 out per 1M, openrouter
	"anthropic/claude-haiku-5.5":              0.05,     // $0.1 in / $0.5 out per 1M, openrouter
	"anthropic/claude-opus-4.1":               7.5,      // $15 in / $75 out per 1M, openrouter
	"anthropic/claude-opus-4.5":               2.5,      // $5 in / $25 out per 1M, openrouter
	"anthropic/claude-opus-4.6":               2.5,      // $5 in / $25 out per 1M, openrouter
	"anthropic/claude-opus-4.7":               2.5,      // $5 in / $25 out per 1M, openrouter
	"anthropic/claude-opus-4.8":               2.5,      // $5 in / $25 out per 1M, openrouter
	"anthropic/claude-opus-5":                 2.5,      // $5 in / $25 out per 1M, openrouter
	"anthropic/claude-opus-5.5":               2,        // $4 in / $20 out per 1M, openrouter
	"anthropic/claude-sonnet-4.5":             1.5,      // $3 in / $15 out per 1M, openrouter
	"anthropic/claude-sonnet-4.6":             1.5,      // $3 in / $15 out per 1M, openrouter
	"anthropic/claude-sonnet-5":               1,        // $2 in / $10 out per 1M, openrouter
	"anthropic/claude-sonnet-5.5":             1,        // $2 in / $10 out per 1M, openrouter
	"bytedance-seed/seed-1.6":                 0.125,    // $0.25 in / $2 out per 1M, openrouter
	"bytedance-seed/seed-1.6-flash":           0.0375,   // $0.075 in / $0.3 out per 1M, openrouter
	"bytedance-seed/seed-2-1-turbo":           0.25,     // $0.5 in / $2.5 out per 1M, openrouter
	"bytedance-seed/seed-2.0-code":            0.25,     // $0.5 in / $3 out per 1M, openrouter
	"bytedance-seed/seed-2.0-lite":            0.125,    // $0.25 in / $2 out per 1M, openrouter
	"bytedance-seed/seed-2.0-mini":            0.05,     // $0.1 in / $0.4 out per 1M, openrouter
	"bytedance/ui-tars-1.5-7b":                0.05,     // $0.1 in / $0.2 out per 1M, openrouter
	"cognitivecomputations/dolphin-mistral-24b-venice-edition": 0.1, // $0.2 in / $0.9 out per 1M, openrouter
	"cohere/command-a":                                   1.25,    // $2.5 in / $10 out per 1M, openrouter
	"cohere/command-a-plus":                              0.15,    // $0.3 in / $1.5 out per 1M, openrouter
	"cohere/command-r-08-2024":                           0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"cohere/command-r-plus-08-2024":                      1.25,    // $2.5 in / $10 out per 1M, openrouter
	"cohere/command-r7b-12-2024":                         0.01875, // $0.0375 in / $0.15 out per 1M, openrouter
	"cohere/north-mini-code:free":                        0,       // free, openrouter
	"deepseek/deepseek-chat":                             0.1287,  // $0.2574 in / $1.0287 out per 1M, openrouter
	"deepseek/deepseek-chat-v3-0324":                     0.145,   // $0.29 in / $1.14 out per 1M, openrouter
	"deepseek/deepseek-chat-v3.1":                        0.125,   // $0.25 in / $0.95 out per 1M, openrouter
	"deepseek/deepseek-r1-0528":                          0.25,    // $0.5 in / $2.15 out per 1M, openrouter
	"deepseek/deepseek-v3.1-terminus":                    0.135,   // $0.27 in / $1 out per 1M, openrouter
	"deepseek/deepseek-v3.2":                             0.13,    // $0.26 in / $0.38 out per 1M, openrouter
	"deepseek/deepseek-v3.2-exp":                         0.135,   // $0.27 in / $0.41 out per 1M, openrouter
	"deepseek/deepseek-v4-flash":                         0.0455,  // $0.091 in / $0.182 out per 1M, openrouter
	"deepseek/deepseek-v4-flash-0731":                    0.0595,  // $0.119 in / $0.238 out per 1M, openrouter
	"deepseek/deepseek-v4-flash-vision-exp":              0.1078,  // $0.2156 in / $0.6468 out per 1M, openrouter
	"deepseek/deepseek-v4-pro":                           0.47415, // $0.9483 in / $1.8966 out per 1M, openrouter
	"deepseek/deepseek-v4-pro-0813":                      0.33,    // $0.66 in / $1.98 out per 1M, openrouter
	"deepseek/deepseek-v4.1-flash":                       0.0225,  // $0.045 in / $0.5 out per 1M, openrouter
	"deepseek/deepseek-v4.1-flash-1017":                  0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"dots-studio/dots-3-note-preview:free":               0,       // free, openrouter
	"elorian/glimpse":                                    0.3,     // $0.6 in / $3 out per 1M, openrouter
	"fireworks/ember-1":                                  1.5,     // $3 in / $15 out per 1M, openrouter
	"google/gemini-2.5-flash":                            0.15,    // $0.3 in / $2.5 out per 1M, openrouter
	"google/gemini-2.5-flash-image":                      0.15,    // $0.3 in / $2.5 out per 1M, openrouter
	"google/gemini-2.5-flash-lite":                       0.05,    // $0.1 in / $0.4 out per 1M, openrouter
	"google/gemini-2.5-pro":                              0.625,   // $1.25 in / $10 out per 1M, openrouter
	"google/gemini-2.5-pro-preview":                      0.625,   // $1.25 in / $10 out per 1M, openrouter
	"google/gemini-3-flash-preview":                      0.25,    // $0.5 in / $3 out per 1M, openrouter
	"google/gemini-3-pro-image":                          1,       // $2 in / $12 out per 1M, openrouter
	"google/gemini-3-pro-image-preview":                  1,       // $2 in / $12 out per 1M, openrouter
	"google/gemini-3.1-flash-image":                      0.25,    // $0.5 in / $3 out per 1M, openrouter
	"google/gemini-3.1-flash-image-preview":              0.25,    // $0.5 in / $3 out per 1M, openrouter
	"google/gemini-3.1-flash-lite":                       0.125,   // $0.25 in / $1.5 out per 1M, openrouter
	"google/gemini-3.1-flash-lite-image":                 0.125,   // $0.25 in / $1.5 out per 1M, openrouter
	"google/gemini-3.1-flash-lite-preview":               0.125,   // $0.25 in / $1.5 out per 1M, openrouter
	"google/gemini-3.1-pro-preview":                      1,       // $2 in / $12 out per 1M, openrouter
	"google/gemini-3.1-pro-preview-customtools":          1,       // $2 in / $12 out per 1M, openrouter
	"google/gemini-3.5-flash":                            0.75,    // $1.5 in / $9 out per 1M, openrouter
	"google/gemini-3.5-flash-lite":                       0.15,    // $0.3 in / $2.5 out per 1M, openrouter
	"google/gemini-3.6-flash":                            0.375,   // $0.75 in / $3.75 out per 1M, openrouter
	"google/gemini-3.7-flash":                            0.375,   // $0.75 in / $3.75 out per 1M, openrouter
	"google/gemini-3.8-flash":                            0.375,   // $0.75 in / $3.75 out per 1M, openrouter
	"google/gemini-4-argon":                              1,       // $2 in / $10 out per 1M, openrouter
	"google/gemini-nano-banana-2.1":                      0.75,    // $1.5 in / $7.5 out per 1M, openrouter
	"google/gemma-2-27b-it":                              0.325,   // $0.65 in / $0.65 out per 1M, openrouter
	"google/gemma-3-12b-it":                              0.025,   // $0.05 in / $0.15 out per 1M, openrouter
	"google/gemma-3-27b-it":                              0.04,    // $0.08 in / $0.35 out per 1M, openrouter
	"google/gemma-3-4b-it":                               0.025,   // $0.05 in / $0.1 out per 1M, openrouter
	"google/gemma-4-26b-a4b-it":                          0.045,   // $0.09 in / $0.3 out per 1M, openrouter
	"google/gemma-4-26b-a4b-it:free":                     0,       // free, openrouter
	"google/gemma-4-31b-it":                              0.1,     // $0.2 in / $0.4 out per 1M, openrouter
	"google/gemma-4-31b-it:free":                         0,       // free, openrouter
	"google/lyria-3-clip-preview":                        0,       // free, openrouter
	"google/lyria-3-pro-preview":                         0,       // free, openrouter
	"gryphe/mythomax-l2-13b":                             0.04,    // $0.08 in / $0.11 out per 1M, openrouter
	"ibm-granite/granite-4.2-8b":                         0.03,    // $0.06 in / $0.15 out per 1M, openrouter
	"inception/mercury-2":                                0.125,   // $0.25 in / $0.75 out per 1M, openrouter
	"inception/mercury-2.5":                              0.02,    // $0.04 in / $0.15 out per 1M, openrouter
	"inception/mercury-3":                                0.05,    // $0.1 in / $0.35 out per 1M, openrouter
	"inclusionai/ling-3.0-flash":                         0.0105,  // $0.021 in / $0.063 out per 1M, openrouter
	"inclusionai/ling-3.0-flash-vl":                      0.03,    // $0.06 in / $0.18 out per 1M, openrouter
	"inclusionai/ling-3.1-flash":                         0.0375,  // $0.075 in / $0.225 out per 1M, openrouter
	"inference-net/schematron-v2-small":                  0.025,   // $0.05 in / $0.23 out per 1M, openrouter
	"inference-net/schematron-v2-turbo":                  0.015,   // $0.03 in / $0.15 out per 1M, openrouter
	"liquid/lfm-2.5-2.6b":                                0.175,   // $0.35 in / $0.5 out per 1M, openrouter
	"liquid/lfm-2.5-2.6b:free":                           0,       // free, openrouter
	"liquid/lfm-2.5-24b-a2b":                             0.025,   // $0.05 in / $0.7 out per 1M, openrouter
	"mancer/weaver":                                      0.2,     // $0.4 in / $0.75 out per 1M, openrouter
	"meituan/longcat-2.0":                                0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"meta-llama/llama-3.1-70b-instruct":                  0.2,     // $0.4 in / $0.4 out per 1M, openrouter
	"meta-llama/llama-3.1-8b-instruct":                   0.025,   // $0.05 in / $0.08 out per 1M, openrouter
	"meta-llama/llama-3.2-1b-instruct":                   0.0135,  // $0.027 in / $0.201 out per 1M, openrouter
	"meta-llama/llama-3.2-3b-instruct":                   0.025,   // $0.05 in / $0.33 out per 1M, openrouter
	"meta-llama/llama-3.3-70b-instruct":                  0.05,    // $0.1 in / $0.32 out per 1M, openrouter
	"meta-llama/llama-4-maverick":                        0.09375, // $0.1875 in / $0.6525 out per 1M, openrouter
	"meta-llama/llama-4-scout":                           0.05,    // $0.1 in / $0.3 out per 1M, openrouter
	"meta-llama/llama-guard-4-12b":                       0.09,    // $0.18 in / $0.18 out per 1M, openrouter
	"meta/muse-glimmer-30b":                              0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"meta/muse-spark-1.1":                                0.625,   // $1.25 in / $4.25 out per 1M, openrouter
	"meta/muse-spark-1.2":                                0.625,   // $1.25 in / $4.25 out per 1M, openrouter
	"meta/muse-spark-1.2-contributor":                    0.05,    // $0.1 in / $0.2 out per 1M, openrouter
	"meta/muse-spark-1.3":                                0.625,   // $1.25 in / $4.25 out per 1M, openrouter
	"meta/muse-spark-1.3-contributor":                    0.05,    // $0.1 in / $0.2 out per 1M, openrouter
	"meta/muse-spark-open":                               0.625,   // $1.25 in / $4.25 out per 1M, openrouter
	"meta/muse-wonder":                                   2,       // $4 in / $20 out per 1M, openrouter
	"microsoft/phi-4":                                    0.035,   // $0.07 in / $0.14 out per 1M, openrouter
	"minimax/minimax-01":                                 0.1,     // $0.2 in / $1.1 out per 1M, openrouter
	"minimax/minimax-m1":                                 0.275,   // $0.55 in / $2.2 out per 1M, openrouter
	"minimax/minimax-m2":                                 0.1275,  // $0.255 in / $1.02 out per 1M, openrouter
	"minimax/minimax-m2-her":                             0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"minimax/minimax-m2.1":                               0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"minimax/minimax-m2.5":                               0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"minimax/minimax-m2.7":                               0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"minimax/minimax-m3":                                 0.15,    // $0.3 in / $1.2 out per 1M, openrouter
	"minimax/minimax-m3.1-flash":                         0.0375,  // $0.075 in / $0.3 out per 1M, openrouter
	"mistralai/codestral-2508":                           0.15,    // $0.3 in / $0.9 out per 1M, openrouter
	"mistralai/devstral-2512":                            0.2,     // $0.4 in / $2 out per 1M, openrouter
	"mistralai/ministral-14b-2512":                       0.1,     // $0.2 in / $0.2 out per 1M, openrouter
	"mistralai/ministral-3b-2512":                        0.05,    // $0.1 in / $0.1 out per 1M, openrouter
	"mistralai/ministral-8b-2512":                        0.075,   // $0.15 in / $0.15 out per 1M, openrouter
	"mistralai/mistral-large":                            1,       // $2 in / $6 out per 1M, openrouter
	"mistralai/mistral-large-2407":                       1,       // $2 in / $6 out per 1M, openrouter
	"mistralai/mistral-large-2512":                       0.25,    // $0.5 in / $1.5 out per 1M, openrouter
	"mistralai/mistral-large-4-0":                        0.35,    // $0.7 in / $2.1 out per 1M, openrouter
	"mistralai/mistral-medium-3":                         0.2,     // $0.4 in / $2 out per 1M, openrouter
	"mistralai/mistral-medium-3-5":                       0.75,    // $1.5 in / $7.5 out per 1M, openrouter
	"mistralai/mistral-medium-3.1":                       0.2,     // $0.4 in / $2 out per 1M, openrouter
	"mistralai/mistral-nemo":                             0.0145,  // $0.029 in / $0.03 out per 1M, openrouter
	"mistralai/mistral-saba":                             0.1,     // $0.2 in / $0.6 out per 1M, openrouter
	"mistralai/mistral-small-24b-instruct-2501":          0.025,   // $0.05 in / $0.08 out per 1M, openrouter
	"mistralai/mistral-small-2603":                       0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"mistralai/mistral-small-3.1-24b-instruct":           0.1755,  // $0.351 in / $0.555 out per 1M, openrouter
	"mistralai/mistral-small-3.2-24b-instruct":           0.0375,  // $0.075 in / $0.2 out per 1M, openrouter
	"mistralai/mixtral-8x22b-instruct":                   1,       // $2 in / $6 out per 1M, openrouter
	"mistralai/voxtral-small-24b-2507":                   0.05,    // $0.1 in / $0.3 out per 1M, openrouter
	"moondream/moondream3-preview":                       0.15,    // $0.3 in / $2.5 out per 1M, openrouter
	"moondream/moondream3.1-9b-a2b":                      0.15,    // $0.3 in / $2.5 out per 1M, openrouter
	"moonshotai/kimi-k2":                                 0.285,   // $0.57 in / $2.3 out per 1M, openrouter
	"moonshotai/kimi-k2-0905":                            0.3,     // $0.6 in / $2.5 out per 1M, openrouter
	"moonshotai/kimi-k2-thinking":                        0.3,     // $0.6 in / $2.5 out per 1M, openrouter
	"moonshotai/kimi-k2.5":                               0.25,    // $0.5 in / $2.5 out per 1M, openrouter
	"moonshotai/kimi-k2.6":                               0.325,   // $0.65 in / $3.41 out per 1M, openrouter
	"moonshotai/kimi-k2.7-code":                          0.3,     // $0.6 in / $3.5 out per 1M, openrouter
	"moonshotai/kimi-k3":                                 0.096,   // $0.192 in / $15 out per 1M, openrouter
	"moonshotai/kimi-k3.1-pro":                           1,       // $2 in / $10 out per 1M, openrouter
	"morph/morph-v3-fast":                                0.4,     // $0.8 in / $1.2 out per 1M, openrouter
	"morph/morph-v3-large":                               0.45,    // $0.9 in / $1.9 out per 1M, openrouter
	"nex-agi/nex-n2.5-mini":                              0.0125,  // $0.025 in / $0.1 out per 1M, openrouter
	"nex-agi/nex-n2.5-pro":                               0.0375,  // $0.075 in / $0.25 out per 1M, openrouter
	"nousresearch/hermes-3-llama-3.1-405b":               0.5,     // $1 in / $1 out per 1M, openrouter
	"nousresearch/hermes-3-llama-3.1-70b":                0.35,    // $0.7 in / $0.7 out per 1M, openrouter
	"nousresearch/hermes-4-405b":                         0.5,     // $1 in / $3 out per 1M, openrouter
	"nvidia/nemotron-3-nano-30b-a3b":                     0.03,    // $0.06 in / $0.24 out per 1M, openrouter
	"nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free": 0,       // free, openrouter
	"nvidia/nemotron-3-super-120b-a12b":                  0.15,    // $0.3 in / $2.4 out per 1M, openrouter
	"nvidia/nemotron-3-ultra-550b-a55b":                  0.3,     // $0.6 in / $2.4 out per 1M, openrouter
	"nvidia/nemotron-3-ultra-550b-a55b:free":             0,       // free, openrouter
	"nvidia/nemotron-3.5-content-safety":                 0.1,     // $0.2 in / $0.2 out per 1M, openrouter
	"nvidia/nemotron-3.5-content-safety:free":            0,       // free, openrouter
	"nvidia/nemotron-3.5-lightning":                      0.035,   // $0.07 in / $0.2 out per 1M, openrouter
	"nvidia/nemotron-3.5-lightning:free":                 0,       // free, openrouter
	"nvidia/nemotron-3.5-super-vl":                       0.075,   // $0.15 in / $0.5 out per 1M, openrouter
	"nvidia/nemotron-3.5-super-vl:free":                  0,       // free, openrouter
	"openai/gpt-3.5-turbo-0613":                          0.5,     // $1 in / $2 out per 1M, openrouter
	"openai/gpt-3.5-turbo-16k":                           1.5,     // $3 in / $4 out per 1M, openrouter
	"openai/gpt-4":                                       15,      // $30 in / $60 out per 1M, openrouter
	"openai/gpt-4.1":                                     1,       // $2 in / $8 out per 1M, openrouter
	"openai/gpt-4.1-mini":                                0.2,     // $0.4 in / $1.6 out per 1M, openrouter
	"openai/gpt-4.1-nano":                                0.05,    // $0.1 in / $0.4 out per 1M, openrouter
	"openai/gpt-4o":                                      1.25,    // $2.5 in / $10 out per 1M, openrouter
	"openai/gpt-4o-2024-05-13":                           2.5,     // $5 in / $15 out per 1M, openrouter
	"openai/gpt-4o-2024-08-06":                           1.25,    // $2.5 in / $10 out per 1M, openrouter
	"openai/gpt-4o-2024-11-20":                           1.25,    // $2.5 in / $10 out per 1M, openrouter
	"openai/gpt-4o-mini":                                 0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"openai/gpt-4o-mini-2024-07-18":                      0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"openai/gpt-5":                                       0.625,   // $1.25 in / $10 out per 1M, openrouter
	"openai/gpt-5-image":                                 5,       // $10 in / $10 out per 1M, openrouter
	"openai/gpt-5-image-mini":                            1.25,    // $2.5 in / $2 out per 1M, openrouter
	"openai/gpt-5-mini":                                  0.125,   // $0.25 in / $2 out per 1M, openrouter
	"openai/gpt-5-nano":                                  0.025,   // $0.05 in / $0.4 out per 1M, openrouter
	"openai/gpt-5-pro":                                   7.5,     // $15 in / $120 out per 1M, openrouter
	"openai/gpt-5.1":                                     0.625,   // $1.25 in / $10 out per 1M, openrouter
	"openai/gpt-5.1-codex":                               0.625,   // $1.25 in / $10 out per 1M, openrouter
	"openai/gpt-5.1-codex-max":                           0.625,   // $1.25 in / $10 out per 1M, openrouter
	"openai/gpt-5.1-codex-mini":                          0.125,   // $0.25 in / $2 out per 1M, openrouter
	"openai/gpt-5.2":                                     0.875,   // $1.75 in / $14 out per 1M, openrouter
	"openai/gpt-5.2-codex":                               0.875,   // $1.75 in / $14 out per 1M, openrouter
	"openai/gpt-5.2-pro":                                 10.5,    // $21 in / $168 out per 1M, openrouter
	"openai/gpt-5.3-codex":                               0.875,   // $1.75 in / $14 out per 1M, openrouter
	"openai/gpt-5.4":                                     1.25,    // $2.5 in / $15 out per 1M, openrouter
	"openai/gpt-5.4-image-2":                             4,       // $8 in / $15 out per 1M, openrouter
	"openai/gpt-5.4-mini":                                0.375,   // $0.75 in / $4.5 out per 1M, openrouter
	"openai/gpt-5.4-nano":                                0.1,     // $0.2 in / $1.25 out per 1M, openrouter
	"openai/gpt-5.4-pro":                                 15,      // $30 in / $180 out per 1M, openrouter
	"openai/gpt-5.5":                                     2.5,     // $5 in / $30 out per 1M, openrouter
	"openai/gpt-5.5-pro":                                 15,      // $30 in / $180 out per 1M, openrouter
	"openai/gpt-5.6-luna":                                0.1,     // $0.2 in / $1.2 out per 1M, openrouter
	"openai/gpt-5.6-luna-pro":                            0.1,     // $0.2 in / $1.2 out per 1M, openrouter
	"openai/gpt-5.6-sol":                                 2,       // $4 in / $20 out per 1M, openrouter
	"openai/gpt-5.6-sol-pro":                             2,       // $4 in / $20 out per 1M, openrouter
	"openai/gpt-5.6-terra":                               1,       // $2 in / $12 out per 1M, openrouter
	"openai/gpt-5.6-terra-pro":                           1,       // $2 in / $12 out per 1M, openrouter
	"openai/gpt-6-astra":                                 5,       // $10 in / $50 out per 1M, openrouter
	"openai/gpt-6-astra-pro":                             5,       // $10 in / $50 out per 1M, openrouter
	"openai/gpt-6-luna":                                  0.05,    // $0.1 in / $0.5 out per 1M, openrouter
	"openai/gpt-6-luna-pro":                              0.05,    // $0.1 in / $0.5 out per 1M, openrouter
	"openai/gpt-6-sol":                                   1,       // $2 in / $10 out per 1M, openrouter
	"openai/gpt-6-sol-pro":                               1,       // $2 in / $10 out per 1M, openrouter
	"openai/gpt-6.1-sol":                                 1,       // $2 in / $10 out per 1M, openrouter
	"openai/gpt-6.1-sol-pro":                             1,       // $2 in / $10 out per 1M, openrouter
	"openai/gpt-6.1-terra":                               0.5,     // $1 in / $5 out per 1M, openrouter
	"openai/gpt-6.1-terra-pro":                           0.5,     // $1 in / $5 out per 1M, openrouter
	"openai/gpt-audio":                                   1.25,    // $2.5 in / $10 out per 1M, openrouter
	"openai/gpt-audio-mini":                              0.3,     // $0.6 in / $2.4 out per 1M, openrouter
	"openai/gpt-chat-latest":                             2.5,     // $5 in / $30 out per 1M, openrouter
	"openai/gpt-oss-120b":                                0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"openai/gpt-oss-20b":                                 0.009,   // $0.018 in / $0.09 out per 1M, openrouter
	"openai/gpt-oss-safeguard-20b":                       0.0375,  // $0.075 in / $0.3 out per 1M, openrouter
	"openai/o3":                                          1,       // $2 in / $8 out per 1M, openrouter
	"openai/o3-pro":                                      10,      // $20 in / $80 out per 1M, openrouter
	"openrouter/free":                                    0,       // free, openrouter
	"particle-ai/particle-1":                             0.025,   // $0.05 in / $0.5 out per 1M, openrouter
	"perceptron/perceptron-mk1":                          0.075,   // $0.15 in / $1.5 out per 1M, openrouter
	"perceptron/perceptron-mk1.5":                        0.075,   // $0.15 in / $1.5 out per 1M, openrouter
	"perplexity/sonar":                                   0.5,     // $1 in / $1 out per 1M, openrouter
	"perplexity/sonar-deep-research":                     1,       // $2 in / $8 out per 1M, openrouter
	"perplexity/sonar-pro":                               1.5,     // $3 in / $15 out per 1M, openrouter
	"perplexity/sonar-pro-search":                        1.5,     // $3 in / $15 out per 1M, openrouter
	"perplexity/sonar-reasoning-pro":                     1,       // $2 in / $8 out per 1M, openrouter
	"poolside/laguna-s-2.1":                              0.045,   // $0.09 in / $0.18 out per 1M, openrouter
	"poolside/laguna-s-2.1:free":                         0,       // free, openrouter
	"poolside/laguna-xs-2.1":                             0.03,    // $0.06 in / $0.12 out per 1M, openrouter
	"poolside/laguna-xs-2.1:free":                        0,       // free, openrouter
	"prism-ml/ternary-bonsai-2-27b":                      0.0375,  // $0.075 in / $0.5 out per 1M, openrouter
	"prism-ml/ternary-bonsai-2-9b":                       0.04,    // $0.08 in / $0.13 out per 1M, openrouter
	"qwen/qwen-2.5-72b-instruct":                         0.18,    // $0.36 in / $0.4 out per 1M, openrouter
	"qwen/qwen-2.5-7b-instruct":                          0.05,    // $0.1 in / $0.2 out per 1M, openrouter
	"qwen/qwen-2.5-coder-32b-instruct":                   0.33,    // $0.66 in / $1 out per 1M, openrouter
	"qwen/qwen-plus":                                     0.13,    // $0.26 in / $0.78 out per 1M, openrouter
	"qwen/qwen2.5-vl-72b-instruct":                       0.4,     // $0.8 in / $1 out per 1M, openrouter
	"qwen/qwen3-14b":                                     0.06,    // $0.12 in / $0.24 out per 1M, openrouter
	"qwen/qwen3-235b-a22b-2507":                          0.045,   // $0.09 in / $0.55 out per 1M, openrouter
	"qwen/qwen3-235b-a22b-thinking-2507":                 0.225,   // $0.45 in / $3.5 out per 1M, openrouter
	"qwen/qwen3-30b-a3b":                                 0.06,    // $0.12 in / $0.5 out per 1M, openrouter
	"qwen/qwen3-30b-a3b-instruct-2507":                   0.05,    // $0.1 in / $0.3 out per 1M, openrouter
	"qwen/qwen3-32b":                                     0.04,    // $0.08 in / $0.28 out per 1M, openrouter
	"qwen/qwen3-coder":                                   0.15,    // $0.3 in / $1 out per 1M, openrouter
	"qwen/qwen3-coder-30b-a3b-instruct":                  0.035,   // $0.07 in / $0.28 out per 1M, openrouter
	"qwen/qwen3-coder-flash":                             0.0975,  // $0.195 in / $0.975 out per 1M, openrouter
	"qwen/qwen3-coder-next":                              0.06,    // $0.12 in / $0.8 out per 1M, openrouter
	"qwen/qwen3-next-80b-a3b-instruct":                   0.045,   // $0.09 in / $1.1 out per 1M, openrouter
	"qwen/qwen3-vl-235b-a22b-instruct":                   0.105,   // $0.21 in / $1.9 out per 1M, openrouter
	"qwen/qwen3-vl-30b-a3b-instruct":                     0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"qwen/qwen3-vl-30b-a3b-thinking":                     0.145,   // $0.29 in / $1 out per 1M, openrouter
	"qwen/qwen3-vl-8b-instruct":                          0.125,   // $0.25 in / $0.75 out per 1M, openrouter
	"qwen/qwen3.5-122b-a10b":                             0.13,    // $0.26 in / $2.08 out per 1M, openrouter
	"qwen/qwen3.5-27b":                                   0.13,    // $0.26 in / $2.6 out per 1M, openrouter
	"qwen/qwen3.5-35b-a3b":                               0.07,    // $0.14 in / $1 out per 1M, openrouter
	"qwen/qwen3.5-397b-a17b":                             0.25,    // $0.5 in / $3.6 out per 1M, openrouter
	"qwen/qwen3.5-9b":                                    0.085,   // $0.17 in / $0.25 out per 1M, openrouter
	"qwen/qwen3.5-flash-02-23":                           0.0325,  // $0.065 in / $0.26 out per 1M, openrouter
	"qwen/qwen3.5-plus-02-15":                            0.13,    // $0.26 in / $1.56 out per 1M, openrouter
	"qwen/qwen3.5-plus-20260420":                         0.15,    // $0.3 in / $1.8 out per 1M, openrouter
	"qwen/qwen3.6-27b":                                   0.16,    // $0.32 in / $3.2 out per 1M, openrouter
	"qwen/qwen3.6-35b-a3b":                               0.05,    // $0.1 in / $0.9 out per 1M, openrouter
	"qwen/qwen3.6-flash":                                 0.09375, // $0.1875 in / $1.125 out per 1M, openrouter
	"qwen/qwen3.6-plus":                                  0.1625,  // $0.325 in / $1.95 out per 1M, openrouter
	"qwen/qwen3.7-flash":                                 0.015,   // $0.03 in / $0.13 out per 1M, openrouter
	"qwen/qwen3.7-max":                                   0.7375,  // $1.475 in / $4.425 out per 1M, openrouter
	"qwen/qwen3.7-plus":                                  0.16,    // $0.32 in / $1.28 out per 1M, openrouter
	"qwen/qwen3.8-2.4t-a95b":                             1,       // $2 in / $6 out per 1M, openrouter
	"qwen/qwen3.8-27b":                                   0.075,   // $0.15 in / $1.875 out per 1M, openrouter
	"qwen/qwen3.8-flash":                                 0.075,   // $0.15 in / $0.47 out per 1M, openrouter
	"qwen/qwen3.8-max-0902":                              1,       // $2 in / $6 out per 1M, openrouter
	"qwen/qwen3.8-max-prime":                             2,       // $4 in / $12 out per 1M, openrouter
	"qwen/qwen3.8-omni-flash":                            0.075,   // $0.15 in / $0.47 out per 1M, openrouter
	"recursive/flint-v1.0:free":                          0,       // free, openrouter
	"reflection/beam-1.0-501b-a23b":                      0.125,   // $0.25 in / $1 out per 1M, openrouter
	"rekaai/reka-edge":                                   0.05,    // $0.1 in / $0.1 out per 1M, openrouter
	"rekaai/reka-flash-3":                                0.05,    // $0.1 in / $0.2 out per 1M, openrouter
	"relace/relace-apply-3":                              0.425,   // $0.85 in / $1.25 out per 1M, openrouter
	"relace/relace-search":                               0.5,     // $1 in / $3 out per 1M, openrouter
	"sakana/fugu-max":                                    1,       // $2 in / $6 out per 1M, openrouter
	"sakana/fugu-ultra":                                  2.5,     // $5 in / $30 out per 1M, openrouter
	"sakana/fugu-ultra-v2":                               2.5,     // $5 in / $30 out per 1M, openrouter
	"sakana/sakana-namazu":                               0.475,   // $0.95 in / $4 out per 1M, openrouter
	"sao10k/l3-lunaris-8b":                               0.02,    // $0.04 in / $0.05 out per 1M, openrouter
	"sao10k/l3.1-euryale-70b":                            0.425,   // $0.85 in / $0.85 out per 1M, openrouter
	"scaledown/extract":                                  0.02,    // $0.04 in / $0 out per 1M, openrouter
	"scaledown/summarize":                                0.02,    // $0.04 in / $0 out per 1M, openrouter
	"stepfun/step-3.5-flash":                             0.05,    // $0.1 in / $0.3 out per 1M, openrouter
	"stepfun/step-3.7-flash":                             0.1,     // $0.2 in / $1.15 out per 1M, openrouter
	"stepfun/step-5-preview":                             0.5,     // $1 in / $2.7 out per 1M, openrouter
	"tencent/hunyuan-a13b-instruct":                      0.07,    // $0.14 in / $0.57 out per 1M, openrouter
	"tencent/hy-mt2-1.8b":                                0.022,   // $0.044 in / $0.177 out per 1M, openrouter
	"tencent/hy-mt2-30b-a3b":                             0.037,   // $0.074 in / $0.295 out per 1M, openrouter
	"tencent/hy-mt2-7b":                                  0.037,   // $0.074 in / $0.295 out per 1M, openrouter
	"tencent/hy3":                                        0.04125, // $0.0825 in / $0.33 out per 1M, openrouter
	"tencent/hy3-preview":                                0.09,    // $0.18 in / $0.6 out per 1M, openrouter
	"tencent/hy4-preview":                                0.3753,  // $0.7506 in / $2.2509 out per 1M, openrouter
	"thedrummer/cydonia-24b-v4.1":                        0.15,    // $0.3 in / $0.5 out per 1M, openrouter
	"thedrummer/skyfall-36b-v2":                          0.275,   // $0.55 in / $0.8 out per 1M, openrouter
	"thedrummer/unslopnemo-12b":                          0.2,     // $0.4 in / $0.4 out per 1M, openrouter
	"thinkingmachines/inkling":                           0.5,     // $1 in / $4.05 out per 1M, openrouter
	"thinkingmachines/inkling-small":                     0.225,   // $0.45 in / $1.2 out per 1M, openrouter
	"thinkingmachines/inkling-small:free":                0,       // free, openrouter
	"thinkingmachines/inkling:free":                      0,       // free, openrouter
	"unbiased/pareto-26.10":                              0.4,     // $0.8 in / $3.2 out per 1M, openrouter
	"unbiased/pareto-26.9":                               1.25,    // $2.5 in / $7.5 out per 1M, openrouter
	"undi95/remm-slerp-l2-13b":                           0.175,   // $0.35 in / $0.65 out per 1M, openrouter
	"upstage/solar-mini4":                                0.025,   // $0.05 in / $0.2 out per 1M, openrouter
	"upstage/solar-pro4":                                 0.075,   // $0.15 in / $0.6 out per 1M, openrouter
	"writer/palmyra-x5":                                  0.3,     // $0.6 in / $6 out per 1M, openrouter
	"x-ai/grok-4.20":                                     0.625,   // $1.25 in / $2.5 out per 1M, openrouter
	"x-ai/grok-4.20-multi-agent":                         0.625,   // $1.25 in / $2.5 out per 1M, openrouter
	"x-ai/grok-4.3":                                      0.625,   // $1.25 in / $2.5 out per 1M, openrouter
	"x-ai/grok-4.5":                                      1,       // $2 in / $6 out per 1M, openrouter
	"x-ai/grok-4.6":                                      1,       // $2 in / $6 out per 1M, openrouter
	"x-ai/grok-4.7":                                      1,       // $2 in / $6 out per 1M, openrouter
	"x-ai/grok-build-0.1":                                0.5,     // $1 in / $2 out per 1M, openrouter
	"xiaomi/mimo-v2.5":                                   0.2,     // $0.4 in / $2 out per 1M, openrouter
	"xiaomi/mimo-v2.5-pro":                               0.24,    // $0.48 in / $1.8 out per 1M, openrouter
	"xiaomi/mimo-v2.6-flash":                             0.07,    // $0.14 in / $0.28 out per 1M, openrouter
	"xiaomi/mimo-v2.6-pro":                               0.2175,  // $0.435 in / $0.87 out per 1M, openrouter
	"xiaomi/mimo-v2.6-pro-ultraspeed":                    2.175,   // $4.35 in / $8.7 out per 1M, openrouter
	"xingchen-agi/xing4.0-29b-a4b:free":                  0,       // free, openrouter
	"z-ai/glm-4.5":                                       0.3,     // $0.6 in / $2.2 out per 1M, openrouter
	"z-ai/glm-4.5-air":                                   0.065,   // $0.13 in / $0.85 out per 1M, openrouter
	"z-ai/glm-4.5v":                                      0.3,     // $0.6 in / $1.8 out per 1M, openrouter
	"z-ai/glm-4.6":                                       0.25,    // $0.5 in / $2 out per 1M, openrouter
	"z-ai/glm-4.6v":                                      0.15,    // $0.3 in / $0.9 out per 1M, openrouter
	"z-ai/glm-4.7":                                       0.27,    // $0.54 in / $1.98 out per 1M, openrouter
	"z-ai/glm-4.7-flash":                                 0.03025, // $0.0605 in / $0.4 out per 1M, openrouter
	"z-ai/glm-5":                                         0.5,     // $1 in / $3.2 out per 1M, openrouter
	"z-ai/glm-5-turbo":                                   0.6,     // $1.2 in / $4 out per 1M, openrouter
	"z-ai/glm-5.1":                                       0.7,     // $1.4 in / $4.4 out per 1M, openrouter
	"z-ai/glm-5.2":                                       0.003,   // $0.006 in / $7 out per 1M, openrouter
	"z-ai/glm-5.3":                                       0.04,    // $0.08 in / $3.6 out per 1M, openrouter
	"z-ai/glm-5.3-flash":                                 0.009,   // $0.018 in / $0.1402 out per 1M, openrouter
	"z-ai/glm-5.3-flashx":                                0.185,   // $0.37 in / $1.25 out per 1M, openrouter
	"z-ai/glm-5.3-prime":                                 1.4,     // $2.8 in / $8.8 out per 1M, openrouter
	"z-ai/glm-5v-turbo":                                  0.6,     // $1.2 in / $4 out per 1M, openrouter
	"~anthropic/claude-fable-latest":                     5,       // $10 in / $50 out per 1M, openrouter
	"~anthropic/claude-haiku-latest":                     0.05,    // $0.1 in / $0.5 out per 1M, openrouter
	"~anthropic/claude-opus-latest":                      2,       // $4 in / $20 out per 1M, openrouter
	"~anthropic/claude-sonnet-latest":                    1,       // $2 in / $10 out per 1M, openrouter
	"~deepseek/deepseek-flash-latest":                    0.011,   // $0.022 in / $0.192 out per 1M, openrouter
	"~deepseek/deepseek-pro-latest":                      0.05175, // $0.1035 in / $4 out per 1M, openrouter
	"~deepseek/deepseek-v4-flash-latest":                 0.01655, // $0.0331 in / $0.3 out per 1M, openrouter
	"~google/gemini-flash-latest":                        0.375,   // $0.75 in / $3.75 out per 1M, openrouter
	"~google/gemini-pro-latest":                          1,       // $2 in / $10 out per 1M, openrouter
	"~moonshotai/kimi-latest":                            0.07645, // $0.1529 in / $10.5572 out per 1M, openrouter
	"~openai/gpt-astra-latest":                           5,       // $10 in / $50 out per 1M, openrouter
	"~openai/gpt-luna-latest":                            0.05,    // $0.1 in / $0.5 out per 1M, openrouter
	"~openai/gpt-mini-latest":                            0.375,   // $0.75 in / $4.5 out per 1M, openrouter
	"~openai/gpt-sol-latest":                             1,       // $2 in / $10 out per 1M, openrouter
	"~openai/gpt-terra-latest":                           0.5,     // $1 in / $5 out per 1M, openrouter
	"~unbiased/pareto-latest":                            0.4,     // $0.8 in / $3.2 out per 1M, openrouter
	"~x-ai/grok-latest":                                  1,       // $2 in / $6 out per 1M, openrouter
	"~z-ai/glm-flash-latest":                             0.009,   // $0.018 in / $0.1402 out per 1M, openrouter
	"~z-ai/glm-latest":                                   0.03495, // $0.0699 in / $1.4 out per 1M, openrouter
}

var CompletionRatio = map[string]float64{
	// aws llama3
	"llama3-8b-8192(33)":  0.0006 / 0.0003,
	"llama3-70b-8192(33)": 0.0035 / 0.00265,
	// whisper
	"whisper-1": 0, // only count input tokens
	// deepseek
	"deepseek-chat":     1.5, // V3.2: $0.42/$0.28 per 1M (off-peak)
	"deepseek-reasoner": 1.5, // V3.2: $0.42/$0.28 per 1M (off-peak)
	// ===================================================================
	// one-api-next price refresh 2026-10-31: explicit completion ratios
	// (= output $ / input $ per 1M) for all newly added models above.
	// ===================================================================
	"deepseek-flash":                    4,        // $0.6/$0.15 per 1M
	"deepseek-v4-flash":                 4,        // $0.6/$0.15 per 1M
	"deepseek-v4-flash-vision-exp":      4,        // $0.6/$0.15 per 1M
	"deepseek-v4-pro":                   3,        // $1.98/$0.66 per 1M
	"glm-4.5":                           3.666667, // $2.2/$0.6 per 1M
	"glm-4.5-air":                       7.499749, // $0.89637/$0.11952 per 1M
	"glm-4.5v":                          3,        // $0.89637/$0.29879 per 1M
	"glm-4.6":                           3.666667, // $2.2/$0.6 per 1M
	"glm-4.6v":                          2.999933, // $0.44819/$0.1494 per 1M
	"glm-4.7":                           4.666615, // $2.09153/$0.44819 per 1M
	"glm-4.7-flashx":                    5.999866, // $0.44819/$0.0747 per 1M
	"glm-5":                             4.5,      // $2.68911/$0.59758 per 1M
	"glm-5.1":                           4,        // $3.58548/$0.89637 per 1M
	"glm-5.2":                           3.5,      // $4.18306/$1.19516 per 1M
	"glm-5.3":                           3.5,      // $4.18306/$1.19516 per 1M
	"glm-5.3-flash":                     3.499916, // $0.41831/$0.11952 per 1M
	"glm-5.3-flashx":                    3.500017, // $1.04577/$0.29879 per 1M
	"glm-5v-turbo":                      4.399971, // $3.28669/$0.74698 per 1M
	"kimi-k2.6":                         4.153743, // $4.0337/$0.9711 per 1M
	"kimi-k2.7-code":                    4.153743, // $4.0337/$0.9711 per 1M
	"kimi-k2.7-code-highspeed":          4.153906, // $8.0673/$1.9421 per 1M
	"kimi-k3":                           5,        // $14.9395/$2.9879 per 1M
	"kimi-k3.1-pro":                     5.230797, // $10.1671/$1.9437 per 1M
	"kimi-k3.1-pro-highspeed":           5.230797, // $20.3342/$3.8874 per 1M
	"MiniMax-M2":                        4.000319, // $1.2549/$0.3137 per 1M
	"MiniMax-M2.1":                      4.000319, // $1.2549/$0.3137 per 1M
	"MiniMax-M2.5":                      4.000319, // $1.2549/$0.3137 per 1M
	"MiniMax-M2.5-highspeed":            3.999681, // $2.5098/$0.6275 per 1M
	"MiniMax-M2.7":                      4.000319, // $1.2549/$0.3137 per 1M
	"MiniMax-M2.7-highspeed":            3.999681, // $2.5098/$0.6275 per 1M
	"MiniMax-M3":                        4.000319, // $1.2549/$0.3137 per 1M
	"MiniMax-M3.1-Flash":                4,        // $0.5976/$0.1494 per 1M
	"grok-4.20-0309-non-reasoning":      2,        // $2.5/$1.25 per 1M
	"grok-4.20-0309-reasoning":          2,        // $2.5/$1.25 per 1M
	"grok-4.20-multi-agent-0309":        2,        // $2.5/$1.25 per 1M
	"grok-4.3":                          2,        // $2.5/$1.25 per 1M
	"grok-4.5":                          3,        // $6/$2 per 1M
	"grok-4.6":                          3,        // $6/$2 per 1M
	"grok-4.7":                          3,        // $6/$2 per 1M
	"grok-build-0.1":                    2,        // $2/$1 per 1M
	"gpt-4.1":                           4,        // $8/$2 per 1M
	"gpt-4.1-mini":                      4,        // $1.6/$0.4 per 1M
	"gpt-4.1-nano":                      4,        // $0.4/$0.1 per 1M
	"gpt-5":                             8,        // $10/$1.25 per 1M
	"gpt-5-mini":                        8,        // $2/$0.25 per 1M
	"gpt-5-nano":                        8,        // $0.4/$0.05 per 1M
	"gpt-5-pro":                         8,        // $120/$15 per 1M
	"gpt-5.1":                           8,        // $10/$1.25 per 1M
	"gpt-5.2":                           8,        // $14/$1.75 per 1M
	"gpt-5.2-pro":                       8,        // $168/$21 per 1M
	"gpt-5.3-codex":                     8,        // $14/$1.75 per 1M
	"gpt-5.3-codex-spark":               8,        // $14/$1.75 per 1M
	"gpt-5.4":                           6,        // $15/$2.5 per 1M
	"gpt-5.4-mini":                      6,        // $4.5/$0.75 per 1M
	"gpt-5.4-nano":                      6.25,     // $1.25/$0.2 per 1M
	"gpt-5.4-pro":                       6,        // $180/$30 per 1M
	"gpt-5.5":                           6,        // $30/$5 per 1M
	"gpt-5.5-pro":                       6,        // $180/$30 per 1M
	"gpt-5.6":                           5,        // $20/$4 per 1M
	"gpt-5.6-luna":                      6,        // $1.2/$0.2 per 1M
	"gpt-5.6-sol":                       5,        // $20/$4 per 1M
	"gpt-5.6-terra":                     6,        // $12/$2 per 1M
	"gpt-6-astra":                       5,        // $50/$10 per 1M
	"gpt-6-luna":                        5,        // $0.5/$0.1 per 1M
	"gpt-6-sol":                         5,        // $10/$2 per 1M
	"gpt-6.1-sol":                       5,        // $10/$2 per 1M
	"gpt-6.1-terra":                     5,        // $5/$1 per 1M
	"gpt-daybreak-blue-latest":          5,        // $20/$4 per 1M
	"gpt-daybreak-red-latest":           6,        // $75/$12.5 per 1M
	"gpt-image-2":                       6,        // $30/$5 per 1M
	"gpt-realtime-2.1":                  6,        // $24/$4 per 1M
	"o1-pro":                            4,        // $600/$150 per 1M
	"o3":                                4,        // $8/$2 per 1M
	"o3-pro":                            4,        // $80/$20 per 1M
	"o4-mini":                           4,        // $4.4/$1.1 per 1M
	"claude-fable-5":                    5,        // $50/$10 per 1M
	"claude-fable-5-1":                  5,        // $50/$10 per 1M
	"claude-haiku-4-5":                  5,        // $5/$1 per 1M
	"claude-haiku-4-5-20251001":         5,        // $5/$1 per 1M
	"claude-haiku-5-5":                  5,        // $0.5/$0.1 per 1M
	"claude-opus-4-5":                   5,        // $25/$5 per 1M
	"claude-opus-4-5-20251101":          5,        // $25/$5 per 1M
	"claude-opus-4-6":                   5,        // $25/$5 per 1M
	"claude-opus-4-7":                   5,        // $25/$5 per 1M
	"claude-opus-4-8":                   5,        // $25/$5 per 1M
	"claude-opus-5":                     5,        // $25/$5 per 1M
	"claude-opus-5-5":                   5,        // $20/$4 per 1M
	"claude-sonnet-4-5":                 5,        // $15/$3 per 1M
	"claude-sonnet-4-5-20250929":        5,        // $15/$3 per 1M
	"claude-sonnet-4-6":                 5,        // $15/$3 per 1M
	"claude-sonnet-5":                   5,        // $10/$2 per 1M
	"claude-sonnet-5-5":                 5,        // $10/$2 per 1M
	"deep-research-max-preview-04-2026": 6,        // $12/$2 per 1M
	"deep-research-preview-04-2026":     6,        // $12/$2 per 1M
	"gemini-2.5-computer-use-preview-10-2025": 8,         // $10/$1.25 per 1M
	"gemini-2.5-flash":                        8.333333,  // $2.5/$0.3 per 1M
	"gemini-2.5-flash-image":                  100,       // $30/$0.3 per 1M
	"gemini-2.5-flash-lite":                   4,         // $0.4/$0.1 per 1M
	"gemini-2.5-flash-preview-tts":            20,        // $10/$0.5 per 1M
	"gemini-2.5-pro":                          8,         // $10/$1.25 per 1M
	"gemini-2.5-pro-preview-tts":              20,        // $20/$1 per 1M
	"gemini-3-flash-preview":                  6,         // $3/$0.5 per 1M
	"gemini-3-pro-image":                      60,        // $120/$2 per 1M
	"gemini-3.1-flash-image":                  120,       // $60/$0.5 per 1M
	"gemini-3.1-flash-lite":                   6,         // $1.5/$0.25 per 1M
	"gemini-3.1-flash-lite-image":             120,       // $30/$0.25 per 1M
	"gemini-3.1-flash-lite-preview":           6,         // $1.5/$0.25 per 1M
	"gemini-3.1-flash-live-preview":           6,         // $4.5/$0.75 per 1M
	"gemini-3.1-flash-tts-preview":            20,        // $20/$1 per 1M
	"gemini-3.1-pro-preview":                  6,         // $12/$2 per 1M
	"gemini-3.1-pro-preview-customtools":      6,         // $12/$2 per 1M
	"gemini-3.5-flash":                        6,         // $9/$1.5 per 1M
	"gemini-3.5-flash-lite":                   8.333333,  // $2.5/$0.3 per 1M
	"gemini-3.5-live-translate-preview":       6,         // $21/$3.5 per 1M
	"gemini-3.6-flash":                        5,         // $3.75/$0.75 per 1M
	"gemini-3.7-flash":                        5,         // $3.75/$0.75 per 1M
	"gemini-3.8-flash":                        5,         // $3.75/$0.75 per 1M
	"gemini-4-argon":                          5,         // $10/$2 per 1M
	"gemini-flash-latest":                     5,         // $3.75/$0.75 per 1M
	"gemini-flash-lite-latest":                8.333333,  // $2.5/$0.3 per 1M
	"gemini-omni-flash-preview":               11.666667, // $17.5/$1.5 per 1M
	"gemini-pro-latest":                       6,         // $12/$2 per 1M
	"deepseek-r1-0528":                        3.996516,  // $2.294/$0.574 per 1M
	"deepseek-v3-1":                           2.998258,  // $1.721/$0.574 per 1M
	"deepseek-v3-2-exp":                       1.501742,  // $0.431/$0.287 per 1M
	"deepseek-v4.1-flash":                     3.999966,  // $1.19015/$0.29754 per 1M
	"kimi-k2-thinking":                        3.996516,  // $2.294/$0.574 per 1M
	"kimi-k2.5":                               5.245645,  // $3.011/$0.574 per 1M
	"moonshot-kimi-k2-instruct":               3.996516,  // $2.294/$0.574 per 1M
	"qvq-max":                                 4,         // $4.588/$1.147 per 1M
	"qwen-deep-research":                      3.018212,  // $23.367/$7.742 per 1M
	"qwen-doc-turbo":                          1.655172,  // $0.144/$0.087 per 1M
	"qwen-flash":                              9.818182,  // $0.216/$0.022 per 1M
	"qwen-long":                               3.986111,  // $0.287/$0.072 per 1M
	"qwen-omni-turbo":                         3.965517,  // $0.23/$0.058 per 1M
	"qwen-omni-turbo-realtime":                3.991304,  // $0.918/$0.23 per 1M
	"qwen-plus-character":                     2.495652,  // $0.287/$0.115 per 1M
	"qwen2-5-omni-7b":                         3.965517,  // $0.345/$0.087 per 1M
	"qwen3-14b":                               3.986111,  // $0.574/$0.144 per 1M
	"qwen3-235b-a22b":                         3.996516,  // $1.147/$0.287 per 1M
	"qwen3-32b":                               3.996516,  // $1.147/$0.287 per 1M
	"qwen3-8b":                                3.986111,  // $0.287/$0.072 per 1M
	"qwen3-coder-30b-a3b-instruct":            3.986111,  // $0.861/$0.216 per 1M
	"qwen3-coder-480b-a35b-instruct":          3.996516,  // $3.441/$0.861 per 1M
	"qwen3-coder-flash":                       3.986111,  // $0.574/$0.144 per 1M
	"qwen3-coder-plus":                        4,         // $2.296/$0.574 per 1M
	"qwen3-max":                               3.994429,  // $1.434/$0.359 per 1M
	"qwen3-next-80b-a3b-instruct":             3.986111,  // $0.574/$0.144 per 1M
	"qwen3-next-80b-a3b-thinking":             9.958333,  // $1.434/$0.144 per 1M
	"qwen3-omni-flash":                        3.833333,  // $0.989/$0.258 per 1M
	"qwen3-omni-flash-realtime":               3.991304,  // $0.918/$0.23 per 1M
	"qwen3-vl-235b-a22b":                      4,         // $1.14682/$0.286705 per 1M
	"qwen3-vl-30b-a3b":                        3.990741,  // $0.431/$0.108 per 1M
	"qwen3-vl-plus":                           9.999965,  // $1.433525/$0.143353 per 1M
	"qwen3.5-397b-a17b":                       6,         // $1.032/$0.172 per 1M
	"qwen3.5-flash":                           9.896552,  // $0.287/$0.029 per 1M
	"qwen3.5-plus":                            5.982609,  // $0.688/$0.115 per 1M
	"qwen3.6-flash":                           6,         // $0.99/$0.165 per 1M
	"qwen3.6-max-preview":                     5.984848,  // $7.9/$1.32 per 1M
	"qwen3.6-plus":                            5.981884,  // $1.651/$0.276 per 1M
	"qwen3.7-flash":                           4.000675,  // $0.1185/$0.02962 per 1M
	"qwen3.7-max":                             3.000606,  // $4.951/$1.65 per 1M
	"qwen3.7-plus":                            3.98913,   // $1.101/$0.276 per 1M
	"qwen3.8-flash":                           3.374568,  // $0.40073/$0.11875 per 1M
	"qwen3.8-max":                             2.999994,  // $5.33231/$1.77744 per 1M
	"qwen3.8-omni-flash":                      3.380531,  // $0.382/$0.113 per 1M
	"qwq-plus":                                2.495652,  // $0.574/$0.23 per 1M
	"tongyi-intent-detect-v3":                 2.482759,  // $0.144/$0.058 per 1M
	"deepseek-v4-flash-0731":                  3,         // $1.32/$0.44 per 1M
	"qwen-plus-character-ja":                  2.8,       // $1.4/$0.5 per 1M
	"qwen3-livetranslate-flash-realtime":      1,         // $10/$10 per 1M
	"qwen3.5-122b-a10b":                       8,         // $3.2/$0.4 per 1M
	"qwen3.5-27b":                             8,         // $2.4/$0.3 per 1M
	"qwen3.5-35b-a3b":                         8,         // $2/$0.25 per 1M
	"qwen3.6-27b":                             6,         // $3.6/$0.6 per 1M
	"qwen3.6-35b-a3b":                         6,         // $2.25/$0.375 per 1M
	"deepseek-v4-flash-ga-260731":             3,         // $1.3359/$0.4453 per 1M
	"deepseek-v4-pro-ga-260813":               3.000007,  // $4.00771/$1.3359 per 1M
	"doubao-seed-1-6-251015":                  9.999747,  // $1.18747/$0.11875 per 1M
	"doubao-seed-1.6-251015":                  9.999747,  // $1.18747/$0.11875 per 1M
	"doubao-seed-1-6-flash-250828":            9.997755,  // $0.22265/$0.02227 per 1M
	"doubao-seed-1-6-vision-250815":           9.999747,  // $1.18747/$0.11875 per 1M
	"doubao-seed-1.8-251228":                  9.999747,  // $1.18747/$0.11875 per 1M
	"doubao-seed-1-8-251228":                  9.999747,  // $1.18747/$0.11875 per 1M
	"doubao-seed-2-0-code-preview-260215":     4.999979,  // $2.37494/$0.47499 per 1M
	"doubao-seed-2-0-lite-260428":             6,         // $0.53436/$0.08906 per 1M
	"doubao-seed-2-0-mini-260428":             9.99899,   // $0.29687/$0.02969 per 1M
	"doubao-seed-2-0-pro-260215":              4.999979,  // $2.37494/$0.47499 per 1M
	"doubao-seed-2-1-pro-260628":              5.000011,  // $4.45301/$0.8906 per 1M
	"doubao-seed-2-1-turbo-260628":            5.000022,  // $2.22651/$0.4453 per 1M
	"doubao-seed-character-260628":            2.499958,  // $0.29687/$0.11875 per 1M
	"doubao-seed-evolving":                    5.000011,  // $4.45301/$0.8906 per 1M
	"glm-5.2-260617":                          3.500004,  // $4.15615/$1.18747 per 1M
	"glm-5-2-260617":                          3.500004,  // $4.15615/$1.18747 per 1M
	"glm-5-3-flash-260828":                    3.500042,  // $0.41563/$0.11875 per 1M
	"step-3.5-flash":                          3.000096,  // $0.31384/$0.10461 per 1M
	"step-3.5-flash-2603":                     3.000096,  // $0.31384/$0.10461 per 1M
	"step-3.7-flash":                          6.000099,  // $1.21052/$0.20175 per 1M
	"step-5-preview":                          2.85715,   // $2.98895/$1.04613 per 1M
	"codestral-2508":                          3,         // $0.9/$0.3 per 1M
	"codestral-latest":                        3,         // $0.9/$0.3 per 1M
	"deepseek-v4-1-flash-1017":                4,         // $1.2/$0.3 per 1M
	"devstral-2512":                           5,         // $2/$0.4 per 1M
	"devstral-latest":                         5,         // $2/$0.4 per 1M
	"devstral-medium-latest":                  5,         // $2/$0.4 per 1M
	"glm-5-2":                                 3.142857,  // $4.4/$1.4 per 1M
	"magistral-medium-latest":                 2.5,       // $5/$2 per 1M
	"ministral-14b-2512":                      1,         // $0.2/$0.2 per 1M
	"ministral-3b-2512":                       1,         // $0.1/$0.1 per 1M
	"ministral-3b-latest":                     1,         // $0.1/$0.1 per 1M
	"ministral-8b-2512":                       1,         // $0.15/$0.15 per 1M
	"ministral-8b-latest":                     1,         // $0.15/$0.15 per 1M
	"mistral-large-2512":                      3,         // $1.5/$0.5 per 1M
	"mistral-large-4":                         3,         // $2.1/$0.7 per 1M
	"mistral-large-4-0-8":                     3,         // $2.1/$0.7 per 1M
	"mistral-medium-2505":                     5,         // $2/$0.4 per 1M
	"mistral-medium-2508":                     5,         // $2/$0.4 per 1M
	"mistral-medium-2604":                     5,         // $7.5/$1.5 per 1M
	"mistral-small-2506":                      3,         // $0.3/$0.1 per 1M
	"mistral-small-2603":                      4,         // $0.6/$0.15 per 1M
	"open-mistral-nemo":                       1,         // $0.15/$0.15 per 1M
	"pixtral-12b":                             1,         // $0.15/$0.15 per 1M
	"voxtral-small-2507":                      4,         // $0.4/$0.1 per 1M
	"voxtral-small-latest":                    4,         // $0.4/$0.1 per 1M
	"zai-glm-5-2":                             3.142857,  // $4.4/$1.4 per 1M
	"zai-glm-5-3":                             3.142857,  // $4.4/$1.4 per 1M
	"zai-glm-5-3-flash":                       3.333333,  // $0.5/$0.15 per 1M
	"c4ai-aya-expanse-32b":                    3,         // $1.5/$0.5 per 1M
	"command-a-03-2025":                       4,         // $10/$2.5 per 1M
	"command-a-plus-05-2026":                  4,         // $10/$2.5 per 1M
	"command-a-reasoning-08-2025":             4,         // $10/$2.5 per 1M
	"command-a-translate-08-2025":             4,         // $10/$2.5 per 1M
	"command-a-vision-07-2025":                4,         // $10/$2.5 per 1M
	"command-r-08-2024":                       4,         // $0.6/$0.15 per 1M
	"command-r-plus-08-2024":                  4,         // $10/$2.5 per 1M
	"command-r7b-12-2024":                     4,         // $0.15/$0.0375 per 1M
	"command-r7b-arabic-02-2025":              4,         // $0.15/$0.0375 per 1M
	"sonar":                                   1,         // $1/$1 per 1M
	"sonar-deep-research":                     4,         // $8/$2 per 1M
	"sonar-pro":                               5,         // $15/$3 per 1M
	"sonar-reasoning-pro":                     4,         // $8/$2 per 1M
	"llama-3.3-70b-versatile":                 1.338983,  // $0.79/$0.59 per 1M
	"aion-labs/aion-2.0":                      2,         // $1.6/$0.8 per 1M
	"aion-labs/aion-3.0":                      2,         // $6/$3 per 1M
	"aion-labs/aion-3.0-mini":                 2,         // $1.4/$0.7 per 1M
	"aion-labs/aion-3.5":                      2,         // $6/$3 per 1M
	"aion-labs/aion-3.5-mini":                 2,         // $1.4/$0.7 per 1M
	"aion-labs/aion-rp-llama-3.1-8b":          2,         // $1.6/$0.8 per 1M
	"amazon/nova-2-lite-v1":                   8.333333,  // $2.5/$0.3 per 1M
	"amazon/nova-lite-v1":                     4,         // $0.24/$0.06 per 1M
	"amazon/nova-micro-v1":                    4,         // $0.14/$0.035 per 1M
	"amazon/nova-premier-v1":                  5,         // $12.5/$2.5 per 1M
	"amazon/nova-pro-v1":                      4,         // $3.2/$0.8 per 1M
	"anthracite-org/magnum-v4-72b":            2,         // $5/$2.5 per 1M
	"anthropic/claude-fable-5":                5,         // $50/$10 per 1M
	"anthropic/claude-fable-5.1":              5,         // $50/$10 per 1M
	"anthropic/claude-haiku-4.5":              5,         // $5/$1 per 1M
	"anthropic/claude-haiku-5.5":              5,         // $0.5/$0.1 per 1M
	"anthropic/claude-opus-4.1":               5,         // $75/$15 per 1M
	"anthropic/claude-opus-4.5":               5,         // $25/$5 per 1M
	"anthropic/claude-opus-4.6":               5,         // $25/$5 per 1M
	"anthropic/claude-opus-4.7":               5,         // $25/$5 per 1M
	"anthropic/claude-opus-4.8":               5,         // $25/$5 per 1M
	"anthropic/claude-opus-5":                 5,         // $25/$5 per 1M
	"anthropic/claude-opus-5.5":               5,         // $20/$4 per 1M
	"anthropic/claude-sonnet-4.5":             5,         // $15/$3 per 1M
	"anthropic/claude-sonnet-4.6":             5,         // $15/$3 per 1M
	"anthropic/claude-sonnet-5":               5,         // $10/$2 per 1M
	"anthropic/claude-sonnet-5.5":             5,         // $10/$2 per 1M
	"bytedance-seed/seed-1.6":                 8,         // $2/$0.25 per 1M
	"bytedance-seed/seed-1.6-flash":           4,         // $0.3/$0.075 per 1M
	"bytedance-seed/seed-2-1-turbo":           5,         // $2.5/$0.5 per 1M
	"bytedance-seed/seed-2.0-code":            6,         // $3/$0.5 per 1M
	"bytedance-seed/seed-2.0-lite":            8,         // $2/$0.25 per 1M
	"bytedance-seed/seed-2.0-mini":            4,         // $0.4/$0.1 per 1M
	"bytedance/ui-tars-1.5-7b":                2,         // $0.2/$0.1 per 1M
	"cognitivecomputations/dolphin-mistral-24b-venice-edition": 4.5, // $0.9/$0.2 per 1M
	"cohere/command-a":                          4,           // $10/$2.5 per 1M
	"cohere/command-a-plus":                     5,           // $1.5/$0.3 per 1M
	"cohere/command-r-08-2024":                  4,           // $0.6/$0.15 per 1M
	"cohere/command-r-plus-08-2024":             4,           // $10/$2.5 per 1M
	"cohere/command-r7b-12-2024":                4,           // $0.15/$0.0375 per 1M
	"deepseek/deepseek-chat":                    3.996503,    // $1.0287/$0.2574 per 1M
	"deepseek/deepseek-chat-v3-0324":            3.931034,    // $1.14/$0.29 per 1M
	"deepseek/deepseek-chat-v3.1":               3.8,         // $0.95/$0.25 per 1M
	"deepseek/deepseek-r1-0528":                 4.3,         // $2.15/$0.5 per 1M
	"deepseek/deepseek-v3.1-terminus":           3.703704,    // $1/$0.27 per 1M
	"deepseek/deepseek-v3.2":                    1.461538,    // $0.38/$0.26 per 1M
	"deepseek/deepseek-v3.2-exp":                1.518519,    // $0.41/$0.27 per 1M
	"deepseek/deepseek-v4-flash":                2,           // $0.182/$0.091 per 1M
	"deepseek/deepseek-v4-flash-0731":           2,           // $0.238/$0.119 per 1M
	"deepseek/deepseek-v4-flash-vision-exp":     3,           // $0.6468/$0.2156 per 1M
	"deepseek/deepseek-v4-pro":                  2,           // $1.8966/$0.9483 per 1M
	"deepseek/deepseek-v4-pro-0813":             3,           // $1.98/$0.66 per 1M
	"deepseek/deepseek-v4.1-flash":              11.111111,   // $0.5/$0.045 per 1M
	"deepseek/deepseek-v4.1-flash-1017":         4,           // $0.6/$0.15 per 1M
	"elorian/glimpse":                           5,           // $3/$0.6 per 1M
	"fireworks/ember-1":                         5,           // $15/$3 per 1M
	"google/gemini-2.5-flash":                   8.333333,    // $2.5/$0.3 per 1M
	"google/gemini-2.5-flash-image":             8.333333,    // $2.5/$0.3 per 1M
	"google/gemini-2.5-flash-lite":              4,           // $0.4/$0.1 per 1M
	"google/gemini-2.5-pro":                     8,           // $10/$1.25 per 1M
	"google/gemini-2.5-pro-preview":             8,           // $10/$1.25 per 1M
	"google/gemini-3-flash-preview":             6,           // $3/$0.5 per 1M
	"google/gemini-3-pro-image":                 6,           // $12/$2 per 1M
	"google/gemini-3-pro-image-preview":         6,           // $12/$2 per 1M
	"google/gemini-3.1-flash-image":             6,           // $3/$0.5 per 1M
	"google/gemini-3.1-flash-image-preview":     6,           // $3/$0.5 per 1M
	"google/gemini-3.1-flash-lite":              6,           // $1.5/$0.25 per 1M
	"google/gemini-3.1-flash-lite-image":        6,           // $1.5/$0.25 per 1M
	"google/gemini-3.1-flash-lite-preview":      6,           // $1.5/$0.25 per 1M
	"google/gemini-3.1-pro-preview":             6,           // $12/$2 per 1M
	"google/gemini-3.1-pro-preview-customtools": 6,           // $12/$2 per 1M
	"google/gemini-3.5-flash":                   6,           // $9/$1.5 per 1M
	"google/gemini-3.5-flash-lite":              8.333333,    // $2.5/$0.3 per 1M
	"google/gemini-3.6-flash":                   5,           // $3.75/$0.75 per 1M
	"google/gemini-3.7-flash":                   5,           // $3.75/$0.75 per 1M
	"google/gemini-3.8-flash":                   5,           // $3.75/$0.75 per 1M
	"google/gemini-4-argon":                     5,           // $10/$2 per 1M
	"google/gemini-nano-banana-2.1":             5,           // $7.5/$1.5 per 1M
	"google/gemma-2-27b-it":                     1,           // $0.65/$0.65 per 1M
	"google/gemma-3-12b-it":                     3,           // $0.15/$0.05 per 1M
	"google/gemma-3-27b-it":                     4.375,       // $0.35/$0.08 per 1M
	"google/gemma-3-4b-it":                      2,           // $0.1/$0.05 per 1M
	"google/gemma-4-26b-a4b-it":                 3.333333,    // $0.3/$0.09 per 1M
	"google/gemma-4-31b-it":                     2,           // $0.4/$0.2 per 1M
	"gryphe/mythomax-l2-13b":                    1.375,       // $0.11/$0.08 per 1M
	"ibm-granite/granite-4.2-8b":                2.5,         // $0.15/$0.06 per 1M
	"inception/mercury-2":                       3,           // $0.75/$0.25 per 1M
	"inception/mercury-2.5":                     3.75,        // $0.15/$0.04 per 1M
	"inception/mercury-3":                       3.5,         // $0.35/$0.1 per 1M
	"inclusionai/ling-3.0-flash":                3,           // $0.063/$0.021 per 1M
	"inclusionai/ling-3.0-flash-vl":             3,           // $0.18/$0.06 per 1M
	"inclusionai/ling-3.1-flash":                3,           // $0.225/$0.075 per 1M
	"inference-net/schematron-v2-small":         4.6,         // $0.23/$0.05 per 1M
	"inference-net/schematron-v2-turbo":         5,           // $0.15/$0.03 per 1M
	"liquid/lfm-2.5-2.6b":                       1.428571,    // $0.5/$0.35 per 1M
	"liquid/lfm-2.5-24b-a2b":                    14,          // $0.7/$0.05 per 1M
	"mancer/weaver":                             1.875,       // $0.75/$0.4 per 1M
	"meituan/longcat-2.0":                       4,           // $1.2/$0.3 per 1M
	"meta-llama/llama-3.1-70b-instruct":         1,           // $0.4/$0.4 per 1M
	"meta-llama/llama-3.1-8b-instruct":          1.6,         // $0.08/$0.05 per 1M
	"meta-llama/llama-3.2-1b-instruct":          7.444444,    // $0.201/$0.027 per 1M
	"meta-llama/llama-3.2-3b-instruct":          6.6,         // $0.33/$0.05 per 1M
	"meta-llama/llama-3.3-70b-instruct":         3.2,         // $0.32/$0.1 per 1M
	"meta-llama/llama-4-maverick":               3.48,        // $0.6525/$0.1875 per 1M
	"meta-llama/llama-4-scout":                  3,           // $0.3/$0.1 per 1M
	"meta-llama/llama-guard-4-12b":              1,           // $0.18/$0.18 per 1M
	"meta/muse-glimmer-30b":                     4,           // $1.2/$0.3 per 1M
	"meta/muse-spark-1.1":                       3.4,         // $4.25/$1.25 per 1M
	"meta/muse-spark-1.2":                       3.4,         // $4.25/$1.25 per 1M
	"meta/muse-spark-1.2-contributor":           2,           // $0.2/$0.1 per 1M
	"meta/muse-spark-1.3":                       3.4,         // $4.25/$1.25 per 1M
	"meta/muse-spark-1.3-contributor":           2,           // $0.2/$0.1 per 1M
	"meta/muse-spark-open":                      3.4,         // $4.25/$1.25 per 1M
	"meta/muse-wonder":                          5,           // $20/$4 per 1M
	"microsoft/phi-4":                           2,           // $0.14/$0.07 per 1M
	"minimax/minimax-01":                        5.5,         // $1.1/$0.2 per 1M
	"minimax/minimax-m1":                        4,           // $2.2/$0.55 per 1M
	"minimax/minimax-m2":                        4,           // $1.02/$0.255 per 1M
	"minimax/minimax-m2-her":                    4,           // $1.2/$0.3 per 1M
	"minimax/minimax-m2.1":                      4,           // $1.2/$0.3 per 1M
	"minimax/minimax-m2.5":                      4,           // $1.2/$0.3 per 1M
	"minimax/minimax-m2.7":                      4,           // $1.2/$0.3 per 1M
	"minimax/minimax-m3":                        4,           // $1.2/$0.3 per 1M
	"minimax/minimax-m3.1-flash":                4,           // $0.3/$0.075 per 1M
	"mistralai/codestral-2508":                  3,           // $0.9/$0.3 per 1M
	"mistralai/devstral-2512":                   5,           // $2/$0.4 per 1M
	"mistralai/ministral-14b-2512":              1,           // $0.2/$0.2 per 1M
	"mistralai/ministral-3b-2512":               1,           // $0.1/$0.1 per 1M
	"mistralai/ministral-8b-2512":               1,           // $0.15/$0.15 per 1M
	"mistralai/mistral-large":                   3,           // $6/$2 per 1M
	"mistralai/mistral-large-2407":              3,           // $6/$2 per 1M
	"mistralai/mistral-large-2512":              3,           // $1.5/$0.5 per 1M
	"mistralai/mistral-large-4-0":               3,           // $2.1/$0.7 per 1M
	"mistralai/mistral-medium-3":                5,           // $2/$0.4 per 1M
	"mistralai/mistral-medium-3-5":              5,           // $7.5/$1.5 per 1M
	"mistralai/mistral-medium-3.1":              5,           // $2/$0.4 per 1M
	"mistralai/mistral-nemo":                    1.034483,    // $0.03/$0.029 per 1M
	"mistralai/mistral-saba":                    3,           // $0.6/$0.2 per 1M
	"mistralai/mistral-small-24b-instruct-2501": 1.6,         // $0.08/$0.05 per 1M
	"mistralai/mistral-small-2603":              4,           // $0.6/$0.15 per 1M
	"mistralai/mistral-small-3.1-24b-instruct":  1.581197,    // $0.555/$0.351 per 1M
	"mistralai/mistral-small-3.2-24b-instruct":  2.666667,    // $0.2/$0.075 per 1M
	"mistralai/mixtral-8x22b-instruct":          3,           // $6/$2 per 1M
	"mistralai/voxtral-small-24b-2507":          3,           // $0.3/$0.1 per 1M
	"moondream/moondream3-preview":              8.333333,    // $2.5/$0.3 per 1M
	"moondream/moondream3.1-9b-a2b":             8.333333,    // $2.5/$0.3 per 1M
	"moonshotai/kimi-k2":                        4.035088,    // $2.3/$0.57 per 1M
	"moonshotai/kimi-k2-0905":                   4.166667,    // $2.5/$0.6 per 1M
	"moonshotai/kimi-k2-thinking":               4.166667,    // $2.5/$0.6 per 1M
	"moonshotai/kimi-k2.5":                      5,           // $2.5/$0.5 per 1M
	"moonshotai/kimi-k2.6":                      5.246154,    // $3.41/$0.65 per 1M
	"moonshotai/kimi-k2.7-code":                 5.833333,    // $3.5/$0.6 per 1M
	"moonshotai/kimi-k3":                        78.125,      // $15/$0.192 per 1M
	"moonshotai/kimi-k3.1-pro":                  5,           // $10/$2 per 1M
	"morph/morph-v3-fast":                       1.5,         // $1.2/$0.8 per 1M
	"morph/morph-v3-large":                      2.111111,    // $1.9/$0.9 per 1M
	"nex-agi/nex-n2.5-mini":                     4,           // $0.1/$0.025 per 1M
	"nex-agi/nex-n2.5-pro":                      3.333333,    // $0.25/$0.075 per 1M
	"nousresearch/hermes-3-llama-3.1-405b":      1,           // $1/$1 per 1M
	"nousresearch/hermes-3-llama-3.1-70b":       1,           // $0.7/$0.7 per 1M
	"nousresearch/hermes-4-405b":                3,           // $3/$1 per 1M
	"nvidia/nemotron-3-nano-30b-a3b":            4,           // $0.24/$0.06 per 1M
	"nvidia/nemotron-3-super-120b-a12b":         8,           // $2.4/$0.3 per 1M
	"nvidia/nemotron-3-ultra-550b-a55b":         4,           // $2.4/$0.6 per 1M
	"nvidia/nemotron-3.5-content-safety":        1,           // $0.2/$0.2 per 1M
	"nvidia/nemotron-3.5-lightning":             2.857143,    // $0.2/$0.07 per 1M
	"nvidia/nemotron-3.5-super-vl":              3.333333,    // $0.5/$0.15 per 1M
	"openai/gpt-3.5-turbo-0613":                 2,           // $2/$1 per 1M
	"openai/gpt-3.5-turbo-16k":                  1.333333,    // $4/$3 per 1M
	"openai/gpt-4":                              2,           // $60/$30 per 1M
	"openai/gpt-4.1":                            4,           // $8/$2 per 1M
	"openai/gpt-4.1-mini":                       4,           // $1.6/$0.4 per 1M
	"openai/gpt-4.1-nano":                       4,           // $0.4/$0.1 per 1M
	"openai/gpt-4o":                             4,           // $10/$2.5 per 1M
	"openai/gpt-4o-2024-05-13":                  3,           // $15/$5 per 1M
	"openai/gpt-4o-2024-08-06":                  4,           // $10/$2.5 per 1M
	"openai/gpt-4o-2024-11-20":                  4,           // $10/$2.5 per 1M
	"openai/gpt-4o-mini":                        4,           // $0.6/$0.15 per 1M
	"openai/gpt-4o-mini-2024-07-18":             4,           // $0.6/$0.15 per 1M
	"openai/gpt-5":                              8,           // $10/$1.25 per 1M
	"openai/gpt-5-image":                        1,           // $10/$10 per 1M
	"openai/gpt-5-image-mini":                   0.8,         // $2/$2.5 per 1M
	"openai/gpt-5-mini":                         8,           // $2/$0.25 per 1M
	"openai/gpt-5-nano":                         8,           // $0.4/$0.05 per 1M
	"openai/gpt-5-pro":                          8,           // $120/$15 per 1M
	"openai/gpt-5.1":                            8,           // $10/$1.25 per 1M
	"openai/gpt-5.1-codex":                      8,           // $10/$1.25 per 1M
	"openai/gpt-5.1-codex-max":                  8,           // $10/$1.25 per 1M
	"openai/gpt-5.1-codex-mini":                 8,           // $2/$0.25 per 1M
	"openai/gpt-5.2":                            8,           // $14/$1.75 per 1M
	"openai/gpt-5.2-codex":                      8,           // $14/$1.75 per 1M
	"openai/gpt-5.2-pro":                        8,           // $168/$21 per 1M
	"openai/gpt-5.3-codex":                      8,           // $14/$1.75 per 1M
	"openai/gpt-5.4":                            6,           // $15/$2.5 per 1M
	"openai/gpt-5.4-image-2":                    1.875,       // $15/$8 per 1M
	"openai/gpt-5.4-mini":                       6,           // $4.5/$0.75 per 1M
	"openai/gpt-5.4-nano":                       6.25,        // $1.25/$0.2 per 1M
	"openai/gpt-5.4-pro":                        6,           // $180/$30 per 1M
	"openai/gpt-5.5":                            6,           // $30/$5 per 1M
	"openai/gpt-5.5-pro":                        6,           // $180/$30 per 1M
	"openai/gpt-5.6-luna":                       6,           // $1.2/$0.2 per 1M
	"openai/gpt-5.6-luna-pro":                   6,           // $1.2/$0.2 per 1M
	"openai/gpt-5.6-sol":                        5,           // $20/$4 per 1M
	"openai/gpt-5.6-sol-pro":                    5,           // $20/$4 per 1M
	"openai/gpt-5.6-terra":                      6,           // $12/$2 per 1M
	"openai/gpt-5.6-terra-pro":                  6,           // $12/$2 per 1M
	"openai/gpt-6-astra":                        5,           // $50/$10 per 1M
	"openai/gpt-6-astra-pro":                    5,           // $50/$10 per 1M
	"openai/gpt-6-luna":                         5,           // $0.5/$0.1 per 1M
	"openai/gpt-6-luna-pro":                     5,           // $0.5/$0.1 per 1M
	"openai/gpt-6-sol":                          5,           // $10/$2 per 1M
	"openai/gpt-6-sol-pro":                      5,           // $10/$2 per 1M
	"openai/gpt-6.1-sol":                        5,           // $10/$2 per 1M
	"openai/gpt-6.1-sol-pro":                    5,           // $10/$2 per 1M
	"openai/gpt-6.1-terra":                      5,           // $5/$1 per 1M
	"openai/gpt-6.1-terra-pro":                  5,           // $5/$1 per 1M
	"openai/gpt-audio":                          4,           // $10/$2.5 per 1M
	"openai/gpt-audio-mini":                     4,           // $2.4/$0.6 per 1M
	"openai/gpt-chat-latest":                    6,           // $30/$5 per 1M
	"openai/gpt-oss-120b":                       4,           // $0.6/$0.15 per 1M
	"openai/gpt-oss-20b":                        5,           // $0.09/$0.018 per 1M
	"openai/gpt-oss-safeguard-20b":              4,           // $0.3/$0.075 per 1M
	"openai/o3":                                 4,           // $8/$2 per 1M
	"openai/o3-pro":                             4,           // $80/$20 per 1M
	"particle-ai/particle-1":                    10,          // $0.5/$0.05 per 1M
	"perceptron/perceptron-mk1":                 10,          // $1.5/$0.15 per 1M
	"perceptron/perceptron-mk1.5":               10,          // $1.5/$0.15 per 1M
	"perplexity/sonar":                          1,           // $1/$1 per 1M
	"perplexity/sonar-deep-research":            4,           // $8/$2 per 1M
	"perplexity/sonar-pro":                      5,           // $15/$3 per 1M
	"perplexity/sonar-pro-search":               5,           // $15/$3 per 1M
	"perplexity/sonar-reasoning-pro":            4,           // $8/$2 per 1M
	"poolside/laguna-s-2.1":                     2,           // $0.18/$0.09 per 1M
	"poolside/laguna-xs-2.1":                    2,           // $0.12/$0.06 per 1M
	"prism-ml/ternary-bonsai-2-27b":             6.666667,    // $0.5/$0.075 per 1M
	"prism-ml/ternary-bonsai-2-9b":              1.625,       // $0.13/$0.08 per 1M
	"qwen/qwen-2.5-72b-instruct":                1.111111,    // $0.4/$0.36 per 1M
	"qwen/qwen-2.5-7b-instruct":                 2,           // $0.2/$0.1 per 1M
	"qwen/qwen-2.5-coder-32b-instruct":          1.515152,    // $1/$0.66 per 1M
	"qwen/qwen-plus":                            3,           // $0.78/$0.26 per 1M
	"qwen/qwen2.5-vl-72b-instruct":              1.25,        // $1/$0.8 per 1M
	"qwen/qwen3-14b":                            2,           // $0.24/$0.12 per 1M
	"qwen/qwen3-235b-a22b-2507":                 6.111111,    // $0.55/$0.09 per 1M
	"qwen/qwen3-235b-a22b-thinking-2507":        7.777778,    // $3.5/$0.45 per 1M
	"qwen/qwen3-30b-a3b":                        4.166667,    // $0.5/$0.12 per 1M
	"qwen/qwen3-30b-a3b-instruct-2507":          3,           // $0.3/$0.1 per 1M
	"qwen/qwen3-32b":                            3.5,         // $0.28/$0.08 per 1M
	"qwen/qwen3-coder":                          3.333333,    // $1/$0.3 per 1M
	"qwen/qwen3-coder-30b-a3b-instruct":         4,           // $0.28/$0.07 per 1M
	"qwen/qwen3-coder-flash":                    5,           // $0.975/$0.195 per 1M
	"qwen/qwen3-coder-next":                     6.666667,    // $0.8/$0.12 per 1M
	"qwen/qwen3-next-80b-a3b-instruct":          12.222222,   // $1.1/$0.09 per 1M
	"qwen/qwen3-vl-235b-a22b-instruct":          9.047619,    // $1.9/$0.21 per 1M
	"qwen/qwen3-vl-30b-a3b-instruct":            4,           // $0.6/$0.15 per 1M
	"qwen/qwen3-vl-30b-a3b-thinking":            3.448276,    // $1/$0.29 per 1M
	"qwen/qwen3-vl-8b-instruct":                 3,           // $0.75/$0.25 per 1M
	"qwen/qwen3.5-122b-a10b":                    8,           // $2.08/$0.26 per 1M
	"qwen/qwen3.5-27b":                          10,          // $2.6/$0.26 per 1M
	"qwen/qwen3.5-35b-a3b":                      7.142857,    // $1/$0.14 per 1M
	"qwen/qwen3.5-397b-a17b":                    7.2,         // $3.6/$0.5 per 1M
	"qwen/qwen3.5-9b":                           1.470588,    // $0.25/$0.17 per 1M
	"qwen/qwen3.5-flash-02-23":                  4,           // $0.26/$0.065 per 1M
	"qwen/qwen3.5-plus-02-15":                   6,           // $1.56/$0.26 per 1M
	"qwen/qwen3.5-plus-20260420":                6,           // $1.8/$0.3 per 1M
	"qwen/qwen3.6-27b":                          10,          // $3.2/$0.32 per 1M
	"qwen/qwen3.6-35b-a3b":                      9,           // $0.9/$0.1 per 1M
	"qwen/qwen3.6-flash":                        6,           // $1.125/$0.1875 per 1M
	"qwen/qwen3.6-plus":                         6,           // $1.95/$0.325 per 1M
	"qwen/qwen3.7-flash":                        4.333333,    // $0.13/$0.03 per 1M
	"qwen/qwen3.7-max":                          3,           // $4.425/$1.475 per 1M
	"qwen/qwen3.7-plus":                         4,           // $1.28/$0.32 per 1M
	"qwen/qwen3.8-2.4t-a95b":                    3,           // $6/$2 per 1M
	"qwen/qwen3.8-27b":                          12.5,        // $1.875/$0.15 per 1M
	"qwen/qwen3.8-flash":                        3.133333,    // $0.47/$0.15 per 1M
	"qwen/qwen3.8-max-0902":                     3,           // $6/$2 per 1M
	"qwen/qwen3.8-max-prime":                    3,           // $12/$4 per 1M
	"qwen/qwen3.8-omni-flash":                   3.133333,    // $0.47/$0.15 per 1M
	"reflection/beam-1.0-501b-a23b":             4,           // $1/$0.25 per 1M
	"rekaai/reka-edge":                          1,           // $0.1/$0.1 per 1M
	"rekaai/reka-flash-3":                       2,           // $0.2/$0.1 per 1M
	"relace/relace-apply-3":                     1.470588,    // $1.25/$0.85 per 1M
	"relace/relace-search":                      3,           // $3/$1 per 1M
	"sakana/fugu-max":                           3,           // $6/$2 per 1M
	"sakana/fugu-ultra":                         6,           // $30/$5 per 1M
	"sakana/fugu-ultra-v2":                      6,           // $30/$5 per 1M
	"sakana/sakana-namazu":                      4.210526,    // $4/$0.95 per 1M
	"sao10k/l3-lunaris-8b":                      1.25,        // $0.05/$0.04 per 1M
	"sao10k/l3.1-euryale-70b":                   1,           // $0.85/$0.85 per 1M
	"stepfun/step-3.5-flash":                    3,           // $0.3/$0.1 per 1M
	"stepfun/step-3.7-flash":                    5.75,        // $1.15/$0.2 per 1M
	"stepfun/step-5-preview":                    2.7,         // $2.7/$1 per 1M
	"tencent/hunyuan-a13b-instruct":             4.071429,    // $0.57/$0.14 per 1M
	"tencent/hy-mt2-1.8b":                       4.022727,    // $0.177/$0.044 per 1M
	"tencent/hy-mt2-30b-a3b":                    3.986486,    // $0.295/$0.074 per 1M
	"tencent/hy-mt2-7b":                         3.986486,    // $0.295/$0.074 per 1M
	"tencent/hy3":                               4,           // $0.33/$0.0825 per 1M
	"tencent/hy3-preview":                       3.333333,    // $0.6/$0.18 per 1M
	"tencent/hy4-preview":                       2.998801,    // $2.2509/$0.7506 per 1M
	"thedrummer/cydonia-24b-v4.1":               1.666667,    // $0.5/$0.3 per 1M
	"thedrummer/skyfall-36b-v2":                 1.454545,    // $0.8/$0.55 per 1M
	"thedrummer/unslopnemo-12b":                 1,           // $0.4/$0.4 per 1M
	"thinkingmachines/inkling":                  4.05,        // $4.05/$1 per 1M
	"thinkingmachines/inkling-small":            2.666667,    // $1.2/$0.45 per 1M
	"unbiased/pareto-26.10":                     4,           // $3.2/$0.8 per 1M
	"unbiased/pareto-26.9":                      3,           // $7.5/$2.5 per 1M
	"undi95/remm-slerp-l2-13b":                  1.857143,    // $0.65/$0.35 per 1M
	"upstage/solar-mini4":                       4,           // $0.2/$0.05 per 1M
	"upstage/solar-pro4":                        4,           // $0.6/$0.15 per 1M
	"writer/palmyra-x5":                         10,          // $6/$0.6 per 1M
	"x-ai/grok-4.20":                            2,           // $2.5/$1.25 per 1M
	"x-ai/grok-4.20-multi-agent":                2,           // $2.5/$1.25 per 1M
	"x-ai/grok-4.3":                             2,           // $2.5/$1.25 per 1M
	"x-ai/grok-4.5":                             3,           // $6/$2 per 1M
	"x-ai/grok-4.6":                             3,           // $6/$2 per 1M
	"x-ai/grok-4.7":                             3,           // $6/$2 per 1M
	"x-ai/grok-build-0.1":                       2,           // $2/$1 per 1M
	"xiaomi/mimo-v2.5":                          5,           // $2/$0.4 per 1M
	"xiaomi/mimo-v2.5-pro":                      3.75,        // $1.8/$0.48 per 1M
	"xiaomi/mimo-v2.6-flash":                    2,           // $0.28/$0.14 per 1M
	"xiaomi/mimo-v2.6-pro":                      2,           // $0.87/$0.435 per 1M
	"xiaomi/mimo-v2.6-pro-ultraspeed":           2,           // $8.7/$4.35 per 1M
	"z-ai/glm-4.5":                              3.666667,    // $2.2/$0.6 per 1M
	"z-ai/glm-4.5-air":                          6.538462,    // $0.85/$0.13 per 1M
	"z-ai/glm-4.5v":                             3,           // $1.8/$0.6 per 1M
	"z-ai/glm-4.6":                              4,           // $2/$0.5 per 1M
	"z-ai/glm-4.6v":                             3,           // $0.9/$0.3 per 1M
	"z-ai/glm-4.7":                              3.666667,    // $1.98/$0.54 per 1M
	"z-ai/glm-4.7-flash":                        6.61157,     // $0.4/$0.0605 per 1M
	"z-ai/glm-5":                                3.2,         // $3.2/$1 per 1M
	"z-ai/glm-5-turbo":                          3.333333,    // $4/$1.2 per 1M
	"z-ai/glm-5.1":                              3.142857,    // $4.4/$1.4 per 1M
	"z-ai/glm-5.2":                              1166.666667, // $7/$0.006 per 1M
	"z-ai/glm-5.3":                              45,          // $3.6/$0.08 per 1M
	"z-ai/glm-5.3-flash":                        7.788889,    // $0.1402/$0.018 per 1M
	"z-ai/glm-5.3-flashx":                       3.378378,    // $1.25/$0.37 per 1M
	"z-ai/glm-5.3-prime":                        3.142857,    // $8.8/$2.8 per 1M
	"z-ai/glm-5v-turbo":                         3.333333,    // $4/$1.2 per 1M
	"~anthropic/claude-fable-latest":            5,           // $50/$10 per 1M
	"~anthropic/claude-haiku-latest":            5,           // $0.5/$0.1 per 1M
	"~anthropic/claude-opus-latest":             5,           // $20/$4 per 1M
	"~anthropic/claude-sonnet-latest":           5,           // $10/$2 per 1M
	"~deepseek/deepseek-flash-latest":           8.727273,    // $0.192/$0.022 per 1M
	"~deepseek/deepseek-pro-latest":             38.647343,   // $4/$0.1035 per 1M
	"~deepseek/deepseek-v4-flash-latest":        9.063444,    // $0.3/$0.0331 per 1M
	"~google/gemini-flash-latest":               5,           // $3.75/$0.75 per 1M
	"~google/gemini-pro-latest":                 5,           // $10/$2 per 1M
	"~moonshotai/kimi-latest":                   69.046436,   // $10.5572/$0.1529 per 1M
	"~openai/gpt-astra-latest":                  5,           // $50/$10 per 1M
	"~openai/gpt-luna-latest":                   5,           // $0.5/$0.1 per 1M
	"~openai/gpt-mini-latest":                   6,           // $4.5/$0.75 per 1M
	"~openai/gpt-sol-latest":                    5,           // $10/$2 per 1M
	"~openai/gpt-terra-latest":                  5,           // $5/$1 per 1M
	"~unbiased/pareto-latest":                   4,           // $3.2/$0.8 per 1M
	"~x-ai/grok-latest":                         3,           // $6/$2 per 1M
	"~z-ai/glm-flash-latest":                    7.788889,    // $0.1402/$0.018 per 1M
	"~z-ai/glm-latest":                          20.028612,   // $1.4/$0.0699 per 1M
}

var (
	DefaultModelRatio      map[string]float64
	DefaultCompletionRatio map[string]float64
)

func init() {
	DefaultModelRatio = make(map[string]float64)
	for k, v := range ModelRatio {
		DefaultModelRatio[k] = v
	}
	DefaultCompletionRatio = make(map[string]float64)
	for k, v := range CompletionRatio {
		DefaultCompletionRatio[k] = v
	}
}

func AddNewMissingRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultModelRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioLock.Lock()
	defer modelRatioLock.Unlock()
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

// GetModelRatio 返回模型倍率；模型不在倍率表中时返回兜底值 30（≈$0.06/1K input）。
//
// Deprecated: 兜底值在按量计费场景下会造成严重错账——真实价格可低至
// $0.0005/1K，静默按 30 计费最高可比真实价格高约 1200 倍。计费路径请改用
// GetModelRatioWithKnown，并在未命中时拒绝服务（机制移植自 omnirelays，
// 见 docs/UPSTREAM.md）。
func GetModelRatio(name string, channelType int) float64 {
	ratio, _ := GetModelRatioWithKnown(name, channelType)
	return ratio
}

// GetModelRatioWithKnown 返回模型倍率，并报告该模型是否真的配置过倍率。
// 查找顺序与历史行为一致：
// name(channelType) → Default(name,channelType) → name → Default(name)。
func GetModelRatioWithKnown(name string, channelType int) (float64, bool) {
	modelRatioLock.RLock()
	defer modelRatioLock.RUnlock()
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	if strings.HasPrefix(name, "command-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	if ratio, ok := ModelRatio[model]; ok {
		return ratio, true
	}
	if ratio, ok := DefaultModelRatio[model]; ok {
		return ratio, true
	}
	if ratio, ok := ModelRatio[name]; ok {
		return ratio, true
	}
	if ratio, ok := DefaultModelRatio[name]; ok {
		return ratio, true
	}
	logger.SysError("model ratio not found: " + name)
	return 30, false
}

func CompletionRatio2JSONString() string {
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		logger.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	completionRatioLock.Lock()
	defer completionRatioLock.Unlock()
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

// AddNewMissingCompletionRatio 把内置输出倍率表中存量 JSON 未覆盖的键增量并入，
// 行为与 AddNewMissingRatio 对 ModelRatio 的处理完全一致（one-api-next 修复：
// 上游仅对 ModelRatio 做增量合并，CompletionRatio 的新默认值永远无法下发到
// 已有数据库，导致升级后新模型的输出倍率不生效）。
func AddNewMissingCompletionRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old completion ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultCompletionRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new completion ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func GetCompletionRatio(name string, channelType int) float64 {
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	completionRatioLock.RLock()
	defer completionRatioLock.RUnlock()
	if ratio, ok := CompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := CompletionRatio[name]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[name]; ok {
		return ratio
	}
	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3
		}
		if strings.HasSuffix(name, "1106") {
			return 2
		}
		return 4.0 / 3.0
	}
	if strings.HasPrefix(name, "gpt-4") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3
			}
			return 4
		}
		if strings.HasPrefix(name, "gpt-4-turbo") ||
			strings.HasSuffix(name, "preview") {
			return 3
		}
		return 2
	}
	// including o1, o1-preview, o1-mini
	if strings.HasPrefix(name, "o1") {
		return 4
	}
	if name == "chatgpt-4o-latest" {
		return 3
	}
	if strings.HasPrefix(name, "claude-3") {
		return 5
	}
	if strings.HasPrefix(name, "claude-") {
		return 3
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3
	}
	if strings.HasPrefix(name, "gemini-") {
		return 3
	}
	if strings.HasPrefix(name, "deepseek-") {
		return 2
	}

	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64
	case "llama3-8b-8192":
		return 2
	case "llama3-70b-8192":
		return 0.79 / 0.59
	case "command", "command-light", "command-nightly", "command-light-nightly":
		return 2
	case "command-r":
		return 3
	case "command-r-plus":
		return 5
	case "grok-beta":
		return 3
	// Replicate Models
	// https://replicate.com/pricing
	case "ibm-granite/granite-20b-code-instruct-8k":
		return 5
	case "ibm-granite/granite-3.0-2b-instruct":
		return 8.333333333333334
	case "ibm-granite/granite-3.0-8b-instruct",
		"ibm-granite/granite-8b-code-instruct-128k":
		return 5
	case "meta/llama-2-13b",
		"meta/llama-2-13b-chat",
		"meta/llama-2-7b",
		"meta/llama-2-7b-chat",
		"meta/meta-llama-3-8b",
		"meta/meta-llama-3-8b-instruct":
		return 5
	case "meta/llama-2-70b",
		"meta/llama-2-70b-chat",
		"meta/meta-llama-3-70b",
		"meta/meta-llama-3-70b-instruct":
		return 2.750 / 0.650 // ≈4.230769
	case "meta/meta-llama-3.1-405b-instruct":
		return 1
	case "mistralai/mistral-7b-instruct-v0.2",
		"mistralai/mistral-7b-v0.1":
		return 5
	case "mistralai/mixtral-8x7b-instruct-v0.1":
		return 1.000 / 0.300 // ≈3.333333
	}

	return 1
}
