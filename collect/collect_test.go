package collect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/samproxy/collect/cache"
	"github.com/honeycombio/samproxy/config"
	"github.com/honeycombio/samproxy/logger"
	"github.com/honeycombio/samproxy/metrics"
	"github.com/honeycombio/samproxy/sample"
	"github.com/honeycombio/samproxy/transmit"
	"github.com/honeycombio/samproxy/types"
)

// TestAddRootSpan tests that adding a root span winds up with a trace object in
// the cache and that that trace gets sent
func TestAddRootSpan(t *testing.T) {
	transmission := &transmit.MockTransmission{}
	transmission.Start()
	coll := &InMemCollector{
		Config: &config.MockConfig{
			GetSendDelayVal:    0,
			GetTraceTimeoutVal: 60,
		},
		Logger:       &logger.NullLogger{},
		Transmission: transmission,
		Sampler:      &sample.DeterministicSampler{},
		Metrics:      &metrics.NullMetrics{},
	}
	c := &cache.DefaultInMemCache{
		Config: cache.CacheConfig{
			CacheCapacity: 3,
		},
		Metrics: &metrics.NullMetrics{},
	}
	c.Start()
	coll.Cache = c

	var traceID = "mytrace"

	span := &types.Span{
		TraceID: traceID,
		Event: types.Event{
			Dataset: "aoeu",
		},
	}
	coll.AddSpan(span)
	time.Sleep(10 * time.Millisecond)
	// adding one span with no parent ID should:
	// * create the trace in the cache
	// * send the trace
	assert.Equal(t, traceID, coll.Cache.Get(traceID).TraceID, "after adding the span, we should have a trace in the cache with the right trace ID")
	assert.Equal(t, 1, len(transmission.Events), "adding a non-root span should not yet send the span")
	assert.Equal(t, "aoeu", transmission.Events[0].Dataset, "sending a root span should immediately send that span via transmission")
}

// TestAddSpan tests that adding a span winds up with a trace object in the
// cache
func TestAddSpan(t *testing.T) {
	transmission := &transmit.MockTransmission{}
	transmission.Start()
	coll := &InMemCollector{
		Config: &config.MockConfig{
			GetSendDelayVal:    0,
			GetTraceTimeoutVal: 60,
		},
		Logger:       &logger.NullLogger{},
		Transmission: transmission,
		Sampler:      &sample.DeterministicSampler{},
		Metrics:      &metrics.NullMetrics{},
	}
	c := &cache.DefaultInMemCache{
		Config: cache.CacheConfig{
			CacheCapacity: 3,
		},
		Metrics: &metrics.NullMetrics{},
	}
	c.Start()
	coll.Cache = c

	var traceID = "mytrace"

	span := &types.Span{
		TraceID: traceID,
		Event: types.Event{
			Dataset: "aoeu",
			Data: map[string]interface{}{
				"trace.parent_id": "unused",
			},
		},
	}
	coll.AddSpan(span)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, traceID, coll.Cache.Get(traceID).TraceID, "after adding the span, we should have a trace in the cache with the right trace ID")
	assert.Equal(t, 0, len(transmission.Events), "adding a non-root span should not yet send the span")
	// ok now let's add the root span and verify that both got sent
	rootSpan := &types.Span{
		TraceID: traceID,
		Event: types.Event{
			Dataset: "aoeu",
			Data:    map[string]interface{}{},
		},
	}
	coll.AddSpan(rootSpan)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 2, len(coll.Cache.Get(traceID).Spans), "after adding a leaf and root span, we should have a two spans in the cache")
	assert.Equal(t, 2, len(transmission.Events), "adding a root span should send all spans in the trace")

}