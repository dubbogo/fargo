package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestGetNXDomain(t *testing.T) {
	convey.Convey("Given nonexistent domain nxd.local.", t, func() {
		resp, _, err := findTXT("nxd.local.")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(len(resp), convey.ShouldEqual, 0)
	})
}

func TestGetNetflixTestDomain(t *testing.T) {
	convey.Convey("Given domain txt.us-east-1.discoverytest.netflix.net.", t, func() {
		// TODO: use a mock DNS server to eliminate dependency on netflix
		// keeping their discoverytest domain up
		resp, ttl, err := findTXT("txt.us-east-1.discoverytest.netflix.net.")
		convey.So(err, convey.ShouldBeNil)
		convey.So(ttl, convey.ShouldEqual, 60*time.Second)
		convey.So(len(resp), convey.ShouldEqual, 3)
		convey.Convey("And the contents are zone records", func() {
			expected := map[string]bool{
				"us-east-1c.us-east-1.discoverytest.netflix.net": true,
				"us-east-1d.us-east-1.discoverytest.netflix.net": true,
				"us-east-1e.us-east-1.discoverytest.netflix.net": true,
			}
			for _, item := range resp {
				_, ok := expected[item]
				convey.So(ok, convey.ShouldEqual, true)
			}
			convey.Convey("And the zone records contain instances", func() {
				for _, record := range resp {
					servers, _, err := findTXT("txt." + record + ".")
					convey.So(err, convey.ShouldBeNil)
					convey.So(len(servers) >= 1, convey.ShouldEqual, true)
					// servers should be EC2 DNS names
					convey.So(servers[0][0:4], convey.ShouldEqual, "ec2-")
				}
			})
		})
	})
	convey.Convey("Autodiscover discoverytest.netflix.net.", t, func() {
		servers, ttl, err := discoverDNS("discoverytest.netflix.net", 7001, "")
		convey.So(ttl, convey.ShouldEqual, 60*time.Second)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(servers), convey.ShouldEqual, 6)
		convey.Convey("Servers discovered should all be EC2 DNS names", func() {
			for _, s := range servers {
				convey.So(s[0:11], convey.ShouldEqual, "http://ec2-")
			}
		})
	})
}
