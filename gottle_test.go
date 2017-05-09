package gottle

import (
	"testing"
	"time"

	"github.com/adelowo/onecache/memory"
)

var _ Throttler = &OnecacheThrottler{}

func TestOnecacheThrottler_Throttle(t *testing.T) {

	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Fatalf("An error occurred while setting up the test ..%v", err)
	}

	r.Header.Set(xForwardedFor, "123.456.789.000")

	throttler := &OnecacheThrottler{
		ipProvider:   NewRealIP(),
		keyGenerator: throttleKey,
		store:        memory.NewInMemoryStore(time.Minute * 10),
	}

	if err := throttler.Throttle(r); err != nil {
		t.Fatalf(`An error occurred while throttling the request .. %v`, err)
	}

	//Manually inspect the cache to make sure it has been stored there
	ip := throttler.ipProvider.IP(r)

	if ok := throttler.store.Has(throttler.keyGenerator(ip)); !ok {
		t.Error(`Expected the request to have been throttled..\n
      Could not find the throttled item in the cache store`)
	}
}
