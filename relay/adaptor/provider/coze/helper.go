package coze

import "github.com/perryles/one-api-next/relay/adaptor/provider/coze/constant/event"

func event2StopReason(e *string) string {
	if e == nil || *e == event.Message {
		return ""
	}
	return "stop"
}
