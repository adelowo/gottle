package gottle

import (
	"time"

	"github.com/adelowo/onecache"
)

//Option provides configuration of the throttler from client code
type Option func(*OnecacheThrottler)

//IP is a configuration Option that sets the provider of the HTTP request
func IP(ip IPProvider) Option {
	return func(t *OnecacheThrottler) {
		t.ipProvider = ip
	}
}

//Store is a configuration Option that allows client code choose one
//of the many adapters supported by onecache
func Store(store onecache.Store) Option {
	return func(t *OnecacheThrottler) {
		t.store = store
	}
}

//KeyGenerator is a configuration Option that allows client code
//choose the way they'd like the ip to be used as a cache key
//Keep in mind that the cache store in use might still
//perform some operation (based on it's own keyGenerator) on the key generated
func KeyGenerator(gen KeyFunc) Option {
	return func(t *OnecacheThrottler) {
		t.keyGenerator = gen
	}
}

//ThrottleCondition provides an Option for configuring the maxRequestss and timeframe
//before a client can be rate limited
func ThrottleCondition(interval time.Duration, maxRequests int) Option {
	return func(t *OnecacheThrottler) {
		t.interval = interval
		t.maxRequests = maxRequests
	}
}
