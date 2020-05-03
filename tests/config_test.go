package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"github.com/dubbogo/fargo"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestConfigs(t *testing.T) {
	convey.Convey("Reading a blank config to test defaults.", t, func() {
		conf, err := fargo.ReadConfig("./config_sample/blank.gcfg")
		convey.So(err, convey.ShouldBeNil)
		convey.So(conf.Eureka.InTheCloud, convey.ShouldEqual, false)
		convey.So(conf.Eureka.ConnectTimeoutSeconds, convey.ShouldEqual, 10)
		convey.So(conf.Eureka.UseDNSForServiceUrls, convey.ShouldEqual, false)
		convey.So(conf.Eureka.ServerDNSName, convey.ShouldEqual, "")
		convey.So(len(conf.Eureka.ServiceUrls), convey.ShouldEqual, 0)
		convey.So(conf.Eureka.ServerPort, convey.ShouldEqual, 7001)
		convey.So(conf.Eureka.PollIntervalSeconds, convey.ShouldEqual, 30)
		convey.So(conf.Eureka.EnableDelta, convey.ShouldEqual, false)
		convey.So(conf.Eureka.PreferSameZone, convey.ShouldEqual, false)
		convey.So(conf.Eureka.RegisterWithEureka, convey.ShouldEqual, false)
	})

	convey.Convey("Testing a config that connects to local eureka instances", t, func() {
		conf, err := fargo.ReadConfig("./config_sample/local.gcfg")
		convey.So(err, convey.ShouldBeNil)
		convey.So(conf.Eureka.InTheCloud, convey.ShouldEqual, false)
		convey.So(conf.Eureka.ConnectTimeoutSeconds, convey.ShouldEqual, 2)
		convey.Convey("Both test servers should be in the service URL list", func() {
			convey.So(conf.Eureka.ServiceUrls, convey.ShouldContain, "http://172.17.0.2:8080/eureka/v2")
			convey.So(conf.Eureka.ServiceUrls, convey.ShouldContain, "http://172.17.0.3:8080/eureka/v2")
		})
		convey.So(conf.Eureka.UseDNSForServiceUrls, convey.ShouldEqual, false)
	})
}
