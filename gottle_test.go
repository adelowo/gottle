package gottle

import (
	"testing"
	"time"

	"github.com/adelowo/onecache/memory"
)

var _ Throttler = NewOneCacheThrottler(nil)

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
		maxRequests:  5,
		interval:     time.Minute * 10,
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

func TestOnecacheThrottler_Throttle_multipleTimes(t *testing.T) {

	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Error(err)
	}

	r.Header.Set(xRealIP, "123.456.789.000")

	throttler := &OnecacheThrottler{
		ipProvider:   NewRealIP(),
		keyGenerator: throttleKey,
		store:        memory.NewInMemoryStore(time.Minute * 10),
		maxRequests:  5,
		interval:     time.Minute * 10,
	}

	if err := throttler.Throttle(r); err != nil {
		t.Fatalf(`An error occurred while throttling the request .. %v`, err)
	}

	//Throttle the request 3 more times
	for it := 0; it < 3; it++ {
		if err := throttler.Throttle(r); err != nil {
			t.Fatalf(`An error occurred while throttling the request .. %v`, err)
		}
	}

	//Then check if the item was actually updated
	key := throttler.keyGenerator(throttler.ipProvider.IP(r))

	buf, err := throttler.store.Get(key)

	if err != nil {
		t.Fatalf(`
      An error occured while trying to fetch the data from the
       cache store... %v`, err)
	}

	item := new(throttledItem)

	if err := DecodeGob(buf, item); err != nil {
		t.Fatalf(`
      An error occured while decoding the bytes slice into an item..%v`, err)
	}

	//since the request was throttled four times and is incremented by 1
	//on every throttle attempt
	expectedHits := 4

	if expectedHits != item.Hits {
		t.Fatalf(`
      The number of hits were not properly recorded after throttling the
      request twice.. \n
      Expected %d.. Got %d`, expectedHits, item.Hits)
	}
}

func TestOnecacheThrottler_IsRateLimited(t *testing.T) {
	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Error(err)
	}

	r.Header.Set(xForwardedFor, "123.456.789.000")

	throttler := &OnecacheThrottler{
		ipProvider:   NewRealIP(),
		keyGenerator: throttleKey,
		store:        memory.NewInMemoryStore(time.Minute * 10),
		maxRequests:  2,
		interval:     time.Second * 5,
	}

	for i := 0; i < 2; i++ { //Throttle the client twice
		throttler.Throttle(r)
	}

	if ok := throttler.IsRateLimited(r); !ok {
		t.Fatalf(`
			The request is supposed to be ratelimited since it has surpassed
			it's max requests condition(%d) in the timeframe allocated
			to it..Expected %v.. Got %v`, throttler.maxRequests, true, ok)
	}

	//Throttling the request after it has surpassed it's throttling conditions
	//should be a no-op and a rate limited error is returned
	if err := throttler.Throttle(r); err != ErrClientIsRateLimited {
		t.Fatalf(`
			The http request is supposed to be rate limited..
			Expected %v. \n Got %v`, ErrClientIsRateLimited, err)
	}
}

func TestOnecacheThrottler_Clear(t *testing.T) {
	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Error(err)
	}

	r.Header.Set(xForwardedFor, "123.456.789.000")

	throttler := &OnecacheThrottler{
		ipProvider:   NewRealIP(),
		keyGenerator: throttleKey,
		store:        memory.NewInMemoryStore(time.Minute * 10),
		maxRequests:  2,
		interval:     time.Second * 5,
	}

	for i := 0; i < 2; i++ { //Throttle the client twice
		throttler.Throttle(r)
	}

	//We have to make sure the client has been rate limited
	//After which we clear out the ratelimit set on the client
	//Then check if the client has been frred of the limits

	if ok := throttler.IsRateLimited(r); !ok {
		t.Fatalf(`
			The request is supposed to be ratelimited since it has surpassed
			it's max requests condition(%d) in the timeframe allocated
			to it..Expected %v.. Got %v`, throttler.maxRequests, true, ok)
	}

	if err := throttler.Clear(r); err != nil {
		t.Fatalf(`
			An error occurred while trying to clear the rate limit
			off the client ... %v`, err)
	}

	//The client must have been freed of the ratelimit
	if ok := throttler.IsRateLimited(r); ok {
		t.Fatal(`
			The request is not supposed to be ratelimited since the ratelimit
			on it has been cleared`)
	}
}

func TestOnecacheThrottler_Clear_forunthrottledrequest(t *testing.T) {
	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Error(err)
	}

	r.Header.Set(xForwardedFor, "123.456.789.000")

	throttler := &OnecacheThrottler{
		ipProvider:   NewRealIP(),
		keyGenerator: throttleKey,
		store:        memory.NewInMemoryStore(time.Minute * 10),
		maxRequests:  2,
		interval:     time.Second * 5,
	}

	if err := throttler.Clear(r); err != nil {
		t.Fatalf(`
			An error occurred while trying to clear the throttle off a client
			.This isn't supposed to have occurred as clearing a ratelimit off an
			 "un throttled" request should be a no-op ... %v`, err)
	}

	//The client must not be ratelimited
	if ok := throttler.IsRateLimited(r); ok {
		t.Fatal(`
			The request is not supposed to be ratelimited since it
			has not been throttled previously`)
	}
}
