package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dubbogo/fargo"
	"github.com/smartystreets/goconvey/convey"
)

func shouldNotBearAnHTTPStatusCode(actual interface{}, expected ...interface{}) string {
	if code, present := fargo.HTTPResponseStatusCode(actual.(error)); present {
		return fmt.Sprintf("Expected: no HTTP status code\nActual:   %d", code)
	}
	return ""
}

func shouldBearHTTPStatusCode(actual interface{}, expected ...interface{}) string {
	expectedCode := expected[0]
	code, present := fargo.HTTPResponseStatusCode(actual.(error))
	if !present {
		return fmt.Sprintf("Expected: %d\nActual:   no HTTP status code", expectedCode)
	}
	if code != expectedCode {
		return fmt.Sprintf("Expected: %d\nActual:   %d", expectedCode, code)
	}
	return ""
}

func withCustomRegisteredInstance(e *fargo.EurekaConnection, application string, hostName string, f func(i *fargo.Instance)) func() {
	return func() {
		vipAddress := "app"
		i := &fargo.Instance{
			HostName:         hostName,
			Port:             9090,
			App:              application,
			IPAddr:           "127.0.0.10",
			VipAddress:       vipAddress,
			DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
			SecureVipAddress: vipAddress,
			Status:           fargo.UP,
			LeaseInfo: fargo.LeaseInfo{
				DurationInSecs: 90,
			},
		}
		convey.So(e.ReregisterInstance(i), convey.ShouldBeNil)

		var wg sync.WaitGroup
		stop := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					if err := e.HeartBeatInstance(i); err != nil {
						if code, ok := fargo.HTTPResponseStatusCode(err); ok && code == http.StatusNotFound {
							e.ReregisterInstance(i)
						}
					}
				}
			}
		}()

		convey.Reset(func() {
			close(stop)
			wg.Wait()
			convey.So(e.DeregisterInstance(i), convey.ShouldBeNil)
		})

		f(i)
	}
}

func withRegisteredInstance(e *fargo.EurekaConnection, f func(i *fargo.Instance)) func() {
	return withCustomRegisteredInstance(e, "TESTAPP", "i-123456", f)
}

func TestConnectionCreation(t *testing.T) {
	convey.Convey("Pull applications", t, func() {
		cfg, err := fargo.ReadConfig("./config_sample/local.gcfg")
		convey.So(err, convey.ShouldBeNil)
		e := fargo.NewConnFromConfig(cfg)
		apps, err := e.GetApps()
		convey.So(err, convey.ShouldBeNil)
		app := apps["EUREKA"]
		convey.So(app,convey.ShouldNotBeNil)
		convey.So(len(app.Instances), convey.ShouldEqual, 2)
	})
}

func TestGetApps(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, j := range []bool{false, true} {
		e.UseJson= j
		convey.Convey("Pull applications", t, func() {
			apps, err := e.GetApps()
			convey.So(err, convey.ShouldBeNil)
			app := apps["EUREKA"]
			convey.So(app, convey.ShouldNotBeNil)
			convey.So(len(app.Instances), convey.ShouldEqual, 2)
		})
		convey.Convey("Pull single application", t, func() {
			app, err := e.GetApp("EUREKA")
			convey.So(err, convey.ShouldBeNil)
			convey.So(app, convey.ShouldNotBeNil)
			convey.So(len(app.Instances), convey.ShouldEqual, 2)
			for _, ins := range app.Instances {
				convey.So(ins.IPAddr, convey.ShouldBeIn, []string{"172.17.0.2", "172.17.0.3"})
			}
		})
	}
}

func TestGetInstancesByNonexistentVIPAddress(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, e.UseJson = range []bool{false, true} {
		convey.Convey("Get instances by VIP address", t, func() {
			convey.Convey("when the VIP address has no instances", func() {
				instances, err := e.GetInstancesByVIPAddress("nonexistent", false)
				convey.So(err, convey.ShouldBeNil)
				convey.So(instances, convey.ShouldBeEmpty)
			})
			convey.Convey("when the secure VIP address has no instances", func() {
				instances, err := e.GetInstancesByVIPAddress("nonexistent", true)
				convey.So(err, convey.ShouldBeNil)
				convey.So(instances, convey.ShouldBeEmpty)
			})
		})
	}
}

func TestGetSingleInstanceByVIPAddress(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	cacheDelay := 35 * time.Second
	vipAddress := "app"
	for _, e.UseJson = range []bool{false, true} {
		convey.Convey("When the VIP address has one instance", t, withRegisteredInstance(&e, func(instance *fargo.Instance) {
			time.Sleep(cacheDelay)
			instances, err := e.GetInstancesByVIPAddress(vipAddress, false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(instances, convey.ShouldHaveLength, 1)
			convey.So(instances[0].VipAddress, convey.ShouldEqual, vipAddress)
			convey.Convey("requesting the instances by that VIP address with status UP should provide that one", func() {
				instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.ThatAreUp)
				convey.So(err, convey.ShouldBeNil)
				convey.So(instances, convey.ShouldHaveLength, 1)
				convey.So(instances[0].VipAddress, convey.ShouldEqual, vipAddress)
				convey.Convey("and when the instance has a status other than UP", func() {
					otherStatus := fargo.OUTOFSERVICE
					convey.So(otherStatus, convey.ShouldNotEqual, fargo.UP)
					err := e.UpdateInstanceStatus(instance, otherStatus)
					convey.So(err, convey.ShouldBeNil)
					convey.Convey("selecting instances with that other status should provide that one", func() {
						time.Sleep(cacheDelay)
						instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.WithStatus(otherStatus))
						convey.So(err, convey.ShouldBeNil)
						convey.So(instances, convey.ShouldHaveLength, 1)
						convey.Convey("And selecting instances with status UP should provide none", func() {
							// Ensure that we tolerate a nil option safely.
							instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.ThatAreUp, nil)
							convey.So(err, convey.ShouldBeNil)
							convey.So(instances, convey.ShouldBeEmpty)
						})
					})
				})
			})
		}))
		convey.Convey("When the secure VIP address has one instance", t, withRegisteredInstance(&e, func(_ *fargo.Instance) {
			convey.Convey("requesting the instances by that VIP address should provide that one", func() {
				time.Sleep(cacheDelay)
				// Ensure that we tolerate a nil option safely.
				instances, err := e.GetInstancesByVIPAddress(vipAddress, true, nil)
				convey.So(err, convey.ShouldBeNil)
				convey.So(instances, convey.ShouldHaveLength, 1)
				convey.So(instances[0].SecureVipAddress, convey.ShouldEqual, vipAddress)
			})
		}))
		time.Sleep(cacheDelay)
	}
}

func TestGetMultipleInstancesByVIPAddress(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	cacheDelay := 35 * time.Second
	for _, e.UseJson = range []bool{false, true} {
		convey.Convey("When the VIP address has one instance", t, withRegisteredInstance(&e, func(instance *fargo.Instance) {
			convey.Convey("when the VIP address has two instances", withCustomRegisteredInstance(&e, "TESTAPP2", "i-234567", func(_ *fargo.Instance) {
				convey.Convey("requesting the instances by that VIP address should provide them", func() {
					time.Sleep(cacheDelay)
					vipAddress := "app"
					instances, err := e.GetInstancesByVIPAddress(vipAddress, false)
					convey.So(err, convey.ShouldBeNil)
					convey.So(instances, convey.ShouldHaveLength, 2)
					for _, ins := range instances {
						convey.So(ins.VipAddress, convey.ShouldEqual, vipAddress)
					}
					convey.So(instances[0], convey.ShouldNotEqual, instances[1])
				})
			}))
		}))
		time.Sleep(cacheDelay)
	}
}

func TestRegistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	for _, j := range []bool{false, true} {
		e.UseJson = j
		convey.Convey("Fail to heartbeat a non-existent instance", t, func() {
			j := fargo.Instance{
				HostName:         "i-6543",
				Port:             9090,
				PortEnabled:      true,
				App:              "TESTAPP",
				IPAddr:           "127.0.0.10",
				VipAddress:       "127.0.0.10",
				DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
				SecureVipAddress: "127.0.0.10",
				Status:           fargo.UP,
			}
			err := e.HeartBeatInstance(&j)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err, shouldBearHTTPStatusCode, http.StatusNotFound)
		})
		convey.Convey("Register an instance to TESTAPP", t, func() {
			convey.Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("Instance can check in", func() {
				err := e.HeartBeatInstance(&i)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	}
}

func TestReregistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")

	for _, j := range []bool{false, true} {
		e.UseJson = j

		i := fargo.Instance{
			HostName:         "i-123456",
			Port:             9090,
			PortEnabled:      true,
			App:              "TESTAPP",
			IPAddr:           "127.0.0.10",
			VipAddress:       "127.0.0.10",
			DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
			SecureVipAddress: "127.0.0.10",
			Status:           fargo.UP,
		}

		convey.Convey("Register a TESTAPP instance", t, func() {
			convey.Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				convey.So(err, convey.ShouldBeNil)

				convey.Convey("Reregister the TESTAPP instance", func() {
					convey.Convey("Instance reregisters correctly", func() {
						err := e.ReregisterInstance(&i)
						convey.So(err, convey.ShouldBeNil)

						convey.Convey("Instance can check in", func() {
							err := e.HeartBeatInstance(&i)
							convey.So(err, convey.ShouldBeNil)
						})

						convey.Convey("Instance can be gotten correctly", func() {
							ii, err := e.GetInstance(i.App, i.HostName)
							convey.So(err, convey.ShouldBeNil)
							convey.So(ii.App, convey.ShouldEqual, i.App)
							convey.So(ii.HostName, convey.ShouldEqual, i.HostName)
						})
					})
				})
			})
		})
	}
}

func DontTestDeregistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	convey.Convey("Register a TESTAPP instance", t, func() {
		convey.Convey("Instance registers correctly", func() {
			err := e.RegisterInstance(&i)
			convey.So(err, convey.ShouldBeNil)
		})
	})
	convey.Convey("Deregister the TESTAPP instance", t, func() {
		convey.Convey("Instance deregisters correctly", func() {
			err := e.DeregisterInstance(&i)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Instance cannot check in", func() {
			err := e.HeartBeatInstance(&i)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err, shouldBearHTTPStatusCode, http.StatusNotFound)
		})
	})
}

func TestUpdateStatus(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	for _, j := range []bool{false, true} {
		e.UseJson = j
		convey.Convey("Register an instance to TESTAPP", t, func() {
			convey.Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				convey.So(err, convey.ShouldBeNil)
			})
		})
		convey.Convey("Update an instance status", t, func() {
			convey.Convey("Instance updates to OUT_OF_SERVICE correctly", func() {
				err := e.UpdateInstanceStatus(&i, fargo.OUTOFSERVICE)
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("Instance updates to UP corectly", func() {
				err := e.UpdateInstanceStatus(&i, fargo.UP)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	}
}

func TestMetadataReading(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	for _, j := range []bool{false, true} {
		e.UseJson = j
		convey.Convey("Read empty instance metadata", t, func() {
			a, err := e.GetApp("EUREKA")
			convey.So(err, convey.ShouldBeNil)
			i := a.Instances[0]
			_, err = i.Metadata.GetString("convey.So(meProp")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Register an instance to TESTAPP", t, func() {
			convey.Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				convey.So(err, convey.ShouldBeNil)

				convey.Convey("Read valid instance metadata", func() {
					a, err := e.GetApp("TESTAPP")
					convey.So(err, convey.ShouldBeNil)
					convey.So(len(a.Instances), convey.ShouldBeGreaterThan, 0)
					i := a.Instances[0]
					err = e.AddMetadataString(i, "convey.So(meProp", "AValue")
					convey.So(err, convey.ShouldBeNil)
					v, err := i.Metadata.GetString("convey.So(meProp")
					convey.So(err, convey.ShouldBeNil)
					convey.So(v, convey.ShouldEqual, "AValue")
				})
			})
		})
	}
}
