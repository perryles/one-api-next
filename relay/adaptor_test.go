package relay

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/perryles/one-api-next/relay/channeltype"
	"github.com/perryles/one-api-next/relay/registry"
	"testing"
)

func TestGetAdaptor(t *testing.T) {
	Convey("get adaptor", t, func() {
		for i := 0; i < channeltype.Dummy; i++ {
			a := GetAdaptor(i)
			So(a, ShouldNotBeNil)
		}
	})
}

func TestGetAdaptorByChannelID(t *testing.T) {
	Convey("get adaptor by channel id", t, func() {
		for _, id := range registry.AllChannelIDs() {
			a := GetAdaptorByChannelID(id)
			So(a, ShouldNotBeNil)
		}
	})
}
