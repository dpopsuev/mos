package network_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dpopsuev/mos/testkit/network"
)

var _ = Describe("Bus event propagation", func() {
	var bus *network.Bus

	BeforeEach(func() {
		bus = network.NewBus()
	})

	Context("when subscribers are connected", func() {
		var aliceCh, bobCh <-chan network.Event

		BeforeEach(func() {
			aliceCh = bus.Subscribe("alice")
			bobCh = bus.Subscribe("bob")
		})

		It("delivers events to all subscribers", func() {
			bus.Publish(network.Event{Type: "rule-created", Source: "alice"})

			Eventually(aliceCh).Should(Receive(WithTransform(
				func(e network.Event) string { return e.Type },
				Equal("rule-created"),
			)))
			Eventually(bobCh).Should(Receive(WithTransform(
				func(e network.Event) string { return e.Type },
				Equal("rule-created"),
			)))
		})

		It("records delivered events for assertion", func() {
			bus.Publish(network.Event{Type: "e1", Source: "sys"})
			bus.Publish(network.Event{Type: "e2", Source: "sys"})

			Eventually(aliceCh).Should(Receive())
			Eventually(aliceCh).Should(Receive())
			Eventually(bobCh).Should(Receive())
			Eventually(bobCh).Should(Receive())

			Expect(bus.Recorder().Events("alice")).To(HaveLen(2))
			Expect(bus.Recorder().Events("bob")).To(HaveLen(2))
		})
	})

	Context("when a subscriber is partitioned", func() {
		var aliceCh, bobCh <-chan network.Event

		BeforeEach(func() {
			aliceCh = bus.Subscribe("alice")
			bobCh = bus.Subscribe("bob")
			bus.Partition("bob")
		})

		It("drops messages for the partitioned subscriber", func() {
			bus.Publish(network.Event{Type: "rule-amended", Source: "alice"})

			Eventually(aliceCh).Should(Receive())
			Consistently(bobCh, 50*time.Millisecond).ShouldNot(Receive())
		})

		It("resumes delivery after healing", func() {
			bus.Publish(network.Event{Type: "event-1", Source: "sys"})
			Eventually(aliceCh).Should(Receive())

			bus.Heal("bob")
			bus.Publish(network.Event{Type: "event-2", Source: "sys"})

			Eventually(bobCh).Should(Receive(WithTransform(
				func(e network.Event) string { return e.Type },
				Equal("event-2"),
			)))
		})
	})

	Context("when delay is injected", func() {
		It("adds latency to message delivery", func() {
			bobCh := bus.Subscribe("bob")
			bus.Delay("bob", 100*time.Millisecond)

			start := time.Now()
			bus.Publish(network.Event{Type: "delayed", Source: "sys"})

			Eventually(bobCh, time.Second).Should(Receive())
			Expect(time.Since(start)).To(BeNumerically(">=", 80*time.Millisecond))
		})

		It("removes latency after clearing delay", func() {
			bobCh := bus.Subscribe("bob")
			bus.Delay("bob", 500*time.Millisecond)
			bus.ClearDelay("bob")

			bus.Publish(network.Event{Type: "fast", Source: "sys"})

			Eventually(bobCh, 100*time.Millisecond).Should(Receive(WithTransform(
				func(e network.Event) string { return e.Type },
				Equal("fast"),
			)))
		})
	})
})
