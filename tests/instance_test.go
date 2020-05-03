package fargo_test

// MIT Licensed (see README.md)

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/dubbogo/fargo"
)

func TestInstanceID(t *testing.T) {
	i := fargo.Instance{
		HostName:         "i-6543",
		Port:             9090,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}

	convey.Convey("Given an instance with DataCenterInfo.Name set to Amazon", t, func() {
		i.DataCenterInfo = fargo.DataCenterInfo{Name: fargo.Amazon}

		convey.Convey("When UniqueID function has NOT been set", func() {
			i.UniqueID = nil

			convey.Convey("And InstanceID has been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = "EXPECTED-ID"

				convey.Convey("Id() should return the provided InstanceID", func() {
					convey.So(i.Id(), convey.ShouldEqual, "EXPECTED-ID")
				})
			})

			convey.Convey("And InstanceID has NOT been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = ""

				convey.Convey("Id() should return an empty string", func() {
					convey.So(i.Id(), convey.ShouldEqual, "")
				})
			})
		})

		convey.Convey("When UniqueID function has been set", func() {
			i.UniqueID = func(i fargo.Instance) string {
				return fmt.Sprintf("%s:%d", i.App, 123)
			}

			convey.Convey("And InstanceID has been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = "UNEXPECTED"

				convey.Convey("Id() should return the ID that is provided by UniqueID", func() {
					convey.So(i.Id(), convey.ShouldEqual, "TESTAPP:123")
				})
			})

			convey.Convey("And InstanceID has not been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = ""

				convey.Convey("Id() should return the ID that is provided by UniqueID", func() {
					convey.So(i.Id(), convey.ShouldEqual, "TESTAPP:123")
				})
			})
		})
	})

	convey.Convey("Given an instance with DataCenterInfo.Name set to MyOwn", t, func() {
		i.DataCenterInfo = fargo.DataCenterInfo{Name: fargo.MyOwn}

		convey.Convey("When UniqueID function has NOT been set", func() {
			i.UniqueID = nil

			convey.Convey("Id() should return the host name", func() {
				convey.So(i.Id(), convey.ShouldEqual, "i-6543")
			})
		})

		convey.Convey("When UniqueID function has been set", func() {
			i.Metadata.Raw = []byte(`{"instanceId": "unique-id"}`)
			i.UniqueID = func(i fargo.Instance) string {
				if id, err := i.Metadata.GetString("instanceId"); err == nil {
					return fmt.Sprintf("%s:%s", i.HostName, id)
				}
				return i.HostName
			}

			convey.Convey("Id() should return the ID that is provided by UniqueID", func() {
				convey.So(i.Id(), convey.ShouldEqual, "i-6543:unique-id")
			})
		})
	})
}
