package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/xml"
	"github.com/dubbogo/fargo"
	"github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
)

func TestGetInt(t *testing.T) {
	convey.Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		convey.Convey("With metadata", func() {
			metadata := new(fargo.InstanceMetadata)
			instance.Metadata = *metadata
			convey.Convey("That has a single integer value", func() {
				key := "d"
				value := 1
				metadata.Raw = []byte("<" + key + ">" + strconv.Itoa(value) + "</" + key + ">")
				convey.Convey("GetInt should return that value", func() {
					actualValue, err := metadata.GetInt(key)
					convey.So(err, convey.ShouldBeNil)
					convey.So(actualValue, convey.ShouldBeNil, value)
				})
			})
		})
	})
}

func TestGetFloat(t *testing.T) {
	convey.Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		convey.Convey("With metadata", func() {
			metadata := new(fargo.InstanceMetadata)
			instance.Metadata = *metadata
			convey.Convey("That has a float64 value", func() {
				key := "d"
				value := 1.9
				metadata.Raw = []byte("<" + key + ">" + strconv.FormatFloat(value, 'f', -1, 64) + "</" + key + ">")
				convey.Convey("GetFloat64 should return that value", func() {
					actualValue, err := metadata.GetFloat64(key)
					convey.So(err, convey.ShouldBeNil)
					convey.So(actualValue, convey.ShouldBeNil, value)
				})
			})
			convey.Convey("That has a float32 value", func() {
				key := "d"
				value := 1.9
				metadata.Raw = []byte("<" + key + ">" + strconv.FormatFloat(value, 'f', -1, 32) + "</" + key + ">")
				convey.Convey("GetFloat32 should return that value", func() {
					actualValue, err := metadata.GetFloat32(key)
					convey.So(err, convey.ShouldBeNil)
					convey.So(actualValue, convey.ShouldBeNil, float32(1.9))
				})
			})
		})
	})
}

func TestSerializeMeta(t *testing.T) {
	convey.Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		convey.Convey("With metadata", func() {
			instance.SetMetadataString("test", "value")
			convey.Convey("Serializing results in correct Jconvey.SoN", func() {
				b, err := instance.Metadata.MarshalJSON()
				convey.So(err, convey.ShouldBeNil)
				convey.So(string(b), convey.ShouldBeNil, "{\"test\":\"value\"}")
			})
			convey.Convey("Serializing results in correct XML", func() {
				b, err := xml.Marshal(instance.Metadata)
				convey.So(err, convey.ShouldBeNil)
				convey.So(string(b), convey.ShouldBeNil, "<InstanceMetadata><test>value</test></InstanceMetadata>")
			})
			convey.Convey("Blank metadata results in blank XML", func() {
				metadata := new(fargo.InstanceMetadata)
				b, err := xml.Marshal(metadata)
				convey.So(err, convey.ShouldBeNil)
				convey.So(string(b), convey.ShouldBeNil, "<InstanceMetadata></InstanceMetadata>")
			})
		})
	})
}
