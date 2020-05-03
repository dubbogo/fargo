package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"errors"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestHTTPResponseStatusCode(t *testing.T) {
	convey.Convey("An nil error should have no HTTP status code", t, func() {
		_, present := HTTPResponseStatusCode(nil)
		convey.So(present, convey.ShouldBeFalse)
	})
	convey.Convey("A foreign error should have no detectable HTTP status code", t, func() {
		_, present := HTTPResponseStatusCode(errors.New("other"))
		convey.So(present, convey.ShouldBeFalse)
	})
	convey.Convey("A fargo error generated from a response from Eureka", t, func() {
		verify := func(err *unsuccessfulHTTPResponse) {
			convey.Convey("should have the given HTTP status code", func() {
				code, present := HTTPResponseStatusCode(err)
				convey.So(present, convey.ShouldBeTrue)
				convey.So(code, convey.ShouldEqual, err.statusCode)
				convey.Convey("should produce a message", func() {
					msg := err.Error()
					if len(err.messagePrefix) == 0 {
						convey.Convey("that lacks a prefx", func() {
							convey.So(msg, convey.ShouldNotStartWith, ",")
						})
					} else {
						convey.Convey("that starts with the given prefix", func() {
							convey.So(msg, convey.ShouldStartWith, err.messagePrefix)
						})
					}
					convey.Convey("that contains the status code in decimal notation", func() {
						convey.So(msg, convey.ShouldContainSubstring, strconv.Itoa(err.statusCode))
					})
				})
			})
		}
		convey.Convey("with a message prefix", func() {
			verify(&unsuccessfulHTTPResponse{500, "operation failed"})
		})
		convey.Convey("without a message prefix", func() {
			verify(&unsuccessfulHTTPResponse{statusCode: 500})
		})
	})
}
