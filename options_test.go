package gottle

import (
	"reflect"
	"testing"
	"time"

	"github.com/adelowo/onecache/filesystem"
)

func TestIp(t *testing.T) {

	ipProvider := NewRealIP()

	throttler := NewOneCacheThrottler(IP(ipProvider))

	if !reflect.DeepEqual(ipProvider, throttler.ipProvider) {
		t.Fatalf(`
      IP providers differ\n .. Expected %v..\n
      Got %v`, ipProvider, throttler.ipProvider)
	}
}

func TestStore(t *testing.T) {
	store := filesystem.MustNewFSStore("cache", time.Minute*15)

	throttler := NewOneCacheThrottler(Store(store))

	if !reflect.DeepEqual(store, throttler.store) {
		t.Fatalf(`
      Cache store differs...\n Expected %v \n
      Got %v`, store, throttler.store)
	}
}

func TestKeyGenerator(t *testing.T) {
	customKeyGenerator := func(ip string) string {
		return "custom-" + ip
	}

	throttler := NewOneCacheThrottler(KeyGenerator(customKeyGenerator))

	ip := "123.456.789.000"

	expected := customKeyGenerator(ip)

	if actual := throttler.keyGenerator(ip); expected != actual {
		t.Fatalf(`
      Generated key differs.. \n
      Expected %v..\n Got %v`, expected, actual)
	}
}

func TestThrottleCondition(t *testing.T) {
	interval := time.Minute
	maxRequests := 60

	throttler := NewOneCacheThrottler(ThrottleCondition(interval, maxRequests))

	if !reflect.DeepEqual(interval, throttler.interval) {
		t.Fatalf(`
      Interval differs... Expected %v \n Got %v`,
			interval, throttler.interval)
	}

	if !reflect.DeepEqual(maxRequests, throttler.maxRequests) {
		t.Fatalf(`
      Max requests differ.. \n
      Expected %d.. Got %d`, maxRequests, throttler.maxRequests)
	}
}
