package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"math/rand"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func instancePredicateFrom(t *testing.T, opts ...InstanceQueryOption) func(*Instance) bool {
	var mergedOptions instanceQueryOptions
	for _, o := range opts {
		if err := o(&mergedOptions); err != nil {
			t.Fatal(err)
		}
	}
	if pred := mergedOptions.predicate; pred != nil {
		return pred
	}
	t.Fatal("no predicate available")
	panic("unreachable")
}

type countingSource struct {
	callCount uint
	seed      int64
}

func (s *countingSource) Int63() int64 {
	s.callCount++
	return s.seed
}

func (s *countingSource) Seed(seed int64) {
	s.seed = seed
}

func (s *countingSource) Reset() {
	s.callCount = 0
}

func TestInstanceQueryOptions(t *testing.T) {
	convey.Convey("A status predicate", t, func() {
		convey.Convey("mandates a nonempty status", func() {
			var opts instanceQueryOptions
			err := WithStatus("")(&opts)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(opts.predicate, convey.ShouldBeNil)
		})
		matchesStatus := func(pred func(*Instance) bool, status StatusType) bool {
			return pred(&Instance{Status: status})
		}
		convey.Convey("matches a single status", func() {
			var opts instanceQueryOptions
			desiredStatus := UNKNOWN
			err := WithStatus(desiredStatus)(&opts)
			convey.So(err, convey.ShouldBeNil)
			pred := opts.predicate
			convey.So(pred, convey.ShouldNotBeNil)
			convey.So(matchesStatus(pred, desiredStatus), convey.ShouldBeTrue)
			for _, status := range []StatusType{UP, DOWN, STARTING, OUTOFSERVICE} {
				convey.So(status, convey.ShouldNotEqual, desiredStatus)
				convey.So(matchesStatus(pred, status), convey.ShouldBeFalse)
			}
		})
		convey.Convey("matches a set of states", func() {
			var opts instanceQueryOptions
			desiredStates := []StatusType{DOWN, OUTOFSERVICE}
			for _, status := range desiredStates {
				err := WithStatus(status)(&opts)
				convey.So(err, convey.ShouldBeNil)
			}
			pred := opts.predicate
			convey.So(pred, convey.ShouldNotBeNil)
			for _, status := range desiredStates {
				convey.So(matchesStatus(pred, status), convey.ShouldBeTrue)
			}
			for _, status := range []StatusType{UP, STARTING, UNKNOWN} {
				convey.So(desiredStates, convey.ShouldNotContain, status)
				convey.So(matchesStatus(pred, status), convey.ShouldBeFalse)
			}
		})
	})
	convey.Convey("A shuffling directive", t, func() {
		convey.Convey("using the global Rand instance", func() {
			var opts instanceQueryOptions
			err := Shuffled(&opts)
			convey.So(err, convey.ShouldBeNil)
			convey.So(opts.intn, convey.ShouldNotBeNil)
			convey.So(opts.intn(1), convey.ShouldEqual, 0)
		})
		convey.Convey("using a specific Rand instance", func() {
			source := countingSource{}
			var opts instanceQueryOptions
			err := ShuffledWith(rand.New(&source))(&opts)
			convey.So(err, convey.ShouldBeNil)
			convey.So(opts.intn, convey.ShouldNotBeNil)
			convey.So(source.callCount, convey.ShouldEqual, 0)
			convey.So(opts.intn(2), convey.ShouldEqual, 0)
			convey.So(source.callCount, convey.ShouldEqual, 1)
		})
	})
}

func TestFilterInstancesInApps(t *testing.T) {
	convey.Convey("A predicate should preserve only those instances", t, func() {
		convey.Convey("with status UP", func() {
			areUp := instancePredicateFrom(t, ThatAreUp)
			convey.Convey("from an empty set of applications", func() {
				convey.So(filterInstancesInApps(nil, areUp), convey.ShouldBeEmpty)
			})
			convey.Convey("from a single application with no instances", func() {
				convey.So(filterInstancesInApps([]*Application{
					&Application{},
				}, areUp), convey.ShouldBeEmpty)
			})
			convey.Convey("from a single application with one DOWN instance", func() {
				convey.So(filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{&Instance{Status: DOWN}},
					},
				}, areUp), convey.ShouldBeEmpty)
			})
			convey.Convey("from a single application with one UP instance", func() {
				instance := &Instance{Status: UP}
				filtered := filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{instance},
					},
				}, areUp)
				convey.So(filtered, convey.ShouldHaveLength, 1)
				convey.So(filtered, convey.ShouldContain, instance)
			})
			convey.Convey("from a single application with multiple instances", func() {
				upInstance := &Instance{Status: UP}
				justHasUpInstance := func(instances ...*Instance) {
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: instances,
						},
					}, areUp)
					convey.So(filtered, convey.ShouldHaveLength, 1)
					convey.So(filtered, convey.ShouldContain, upInstance)
				}
				downInstance := &Instance{Status: DOWN}
				convey.Convey("with UP instance first", func() {
					justHasUpInstance(upInstance, downInstance)
				})
				convey.Convey("with UP instance last", func() {
					justHasUpInstance(downInstance, upInstance)
				})
				convey.Convey("with multiple UP instances", func() {
					secondUpInstance := &Instance{Status: UP}
					thirdUpInstance := &Instance{Status: UP}
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, downInstance, secondUpInstance, thirdUpInstance, &Instance{Status: OUTOFSERVICE}},
						},
					}, areUp)
					convey.So(filtered, convey.ShouldHaveLength, 3)
					convey.So(filtered, convey.ShouldContain, upInstance)
					convey.So(filtered, convey.ShouldContain, secondUpInstance)
					convey.So(filtered, convey.ShouldContain, thirdUpInstance)
				})
			})
			convey.Convey("from multiple applications", func() {
				firstUpInstance := &Instance{Status: UP}
				secondUpInstance := &Instance{Status: UP}
				filtered := filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{firstUpInstance, &Instance{Status: OUTOFSERVICE}},
					},
					&Application{},
					&Application{
						Instances: []*Instance{&Instance{Status: DOWN}, secondUpInstance},
					},
					&Application{
						Instances: []*Instance{&Instance{Status: UNKNOWN}},
					},
				}, areUp)
				convey.So(filtered, convey.ShouldHaveLength, 2)
				convey.So(filtered, convey.ShouldContain, firstUpInstance)
				convey.So(filtered, convey.ShouldContain, secondUpInstance)
			})
		})
		convey.Convey("with status matching any of those designated", func() {
			upInstance := &Instance{Status: UP}
			downInstance := &Instance{Status: DOWN}
			startingInstance := &Instance{Status: STARTING}
			outOfServiceInstance := &Instance{Status: OUTOFSERVICE}
			pred := instancePredicateFrom(t, WithStatus(DOWN), WithStatus(OUTOFSERVICE))
			convey.Convey("from a single application", func() {
				convey.Convey("with no matching instances", func() {
					convey.So(filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, startingInstance},
						},
					}, pred), convey.ShouldBeEmpty)
				})
				convey.Convey("with two matching instances", func() {
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, downInstance, startingInstance, outOfServiceInstance},
						},
					}, pred)
					convey.So(filtered, convey.ShouldHaveLength, 2)
					convey.So(filtered, convey.ShouldContain, downInstance)
					convey.So(filtered, convey.ShouldContain, outOfServiceInstance)
				})
			})
		})
	})
}
// Preclude compiler optimization eliding the filter procedure.
var filterBenchmarkResult []*Instance

type filterInstancesFunc func([]*Instance, func(*Instance) bool) []*Instance

func benchmarkFilterInstancesFunc(b *testing.B, f filterInstancesFunc) {
	retainAll := func(*Instance) bool { return true }
	dropAll := func(*Instance) bool { return false }
	thatAreUp := func(instance *Instance) bool { return instance.Status == UP }

	type runLengthByStatus struct {
		Up, Down, Starting, OutOfService, Unknown int
	}
	synthesizeInstances := func(rls ...runLengthByStatus) []*Instance {
		var instances []*Instance
		push := func(status StatusType, n int) {
			if n <= 0 {
				return
			}
			instance := &Instance{Status: status}
			for i := 0; ; {
				instances = append(instances, instance)
				i++
				if i == n {
					break
				}
			}
		}
		for _, rl := range rls {
			push(UP, rl.Up)
			push(DOWN, rl.Down)
			push(STARTING, rl.Starting)
			push(OUTOFSERVICE, rl.OutOfService)
			push(UNKNOWN, rl.OutOfService)
		}
		return instances
	}
	filter := func(n int, f filterInstancesFunc, instances []*Instance, pred func(*Instance) bool) {
		var result []*Instance
		for i := 0; i != n; i++ {
			result = f(instances, pred)
		}
		filterBenchmarkResult = result
	}
	benchAllNoneAndUp := func(b *testing.B, instances []*Instance) {
		b.Run("all", func(b *testing.B) {
			filter(b.N, f, instances, retainAll)
		})
		b.Run("none", func(b *testing.B) {
			filter(b.N, f, instances, dropAll)
		})
		b.Run("up", func(b *testing.B) {
			filter(b.N, f, instances, thatAreUp)
		})
	}

	b.Run("1↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1}))
	})
	b.Run("10↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 10}))
	})
	b.Run("100↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 100}))
	})
	b.Run("1000↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1000}))
	})
	b.Run("1↑1↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1, Down: 1}))
	})
	b.Run("1↑9↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1, Down: 9}))
	})
	b.Run("1↑99↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1, Down: 99}))
	})
	b.Run("1↑999↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 1, Down: 999}))
	})
	b.Run("9↑1↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 9, Down: 1}))
	})
	b.Run("99↑1↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 99, Down: 1}))
	})
	b.Run("999↑1↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 999, Down: 1}))
	})
	b.Run("3↓4↑3↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Down: 3}, runLengthByStatus{Up: 4, Down: 3}))
	})
	b.Run("3↑4↓3↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 3, Down: 4}, runLengthByStatus{Up: 3}))
	})
	b.Run("33↓34↑33↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Down: 33}, runLengthByStatus{Up: 34, Down: 33}))
	})
	b.Run("33↑34↓33↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 33, Down: 34}, runLengthByStatus{Up: 33}))
	})
	b.Run("333↓334↑333↓", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Down: 333}, runLengthByStatus{Up: 334, Down: 333}))
	})
	b.Run("333↑334↓333↑", func(b *testing.B) {
		benchAllNoneAndUp(b, synthesizeInstances(runLengthByStatus{Up: 333, Down: 334}, runLengthByStatus{Up: 333}))
	})
}

func BenchmarkFilterInstances(b *testing.B) {
	benchmarkFilterInstancesFunc(b, filterInstances)
}
