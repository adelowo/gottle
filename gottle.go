//Gottle is an HTTP ratelimiter built ontop of the onecache library.
package gottle

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/adelowo/onecache"
	"github.com/adelowo/onecache/memory"
)

const (
	defaultThrottledItemIncrement = 1
	defaultMaxRequests            = 10
	defaultInterval               = time.Minute * 10
)

//ErrClientIsRateLimited is an error value that signifies a client has been ratelimited
var ErrClientIsRateLimited = errors.New(
	`gottle: The client is currently rate limited`)

//KeyFunc is a function type for setting the key in the cache
type KeyFunc func(ip string) string

//IPProvider provides an interface for fetching the IP of an HTTP request
type IPProvider interface {
	IP(r *http.Request) string
}

//Throttler defines the operation needed to limit clients
//and check if an HTTP request is currently rate limited
type Throttler interface {
	Throttle(r *http.Request) error
	Clear(r *http.Request) error
}

//ThrottlerAttempts provides access to stats about the current request
type ThrottlerAttempts interface {
	Attempts(r *http.Request) (int, error)
	IsRateLimited(r *http.Request) bool
}

//OnecacheThrottler provides an implementation of Throttler by
//making use of onecache's cache implementation
type OnecacheThrottler struct {
	ipProvider   IPProvider
	store        onecache.Store
	keyGenerator KeyFunc
	maxRequests  int
	interval     time.Duration
}

//NewOneCacheThrottler returns an instance of OnecacheThrottler
func NewOneCacheThrottler(opts ...Option) *OnecacheThrottler {

	throttler := &OnecacheThrottler{
		maxRequests: defaultMaxRequests,
		interval:    defaultInterval}

	for _, opt := range opts {
		if opt != nil {
			opt(throttler)
		}
	}

	setDefaultsForEmptyFields(throttler)

	return throttler
}

func setDefaultsForEmptyFields(throttler *OnecacheThrottler) {
	if throttler.ipProvider == nil {
		throttler.ipProvider = NewRealIP()
	}

	if throttler.keyGenerator == nil {
		throttler.keyGenerator = throttleKey
	}

	if throttler.store == nil {
		throttler.store = memory.NewInMemoryStore(time.Minute * 20)
	}
}

type throttledItem struct {
	LastThrottledAt time.Time //The most recent throttle time, so we can diff to lockout or not
	Hits            int
}

//IsRateLimited checks if a client has reached his/her maximum number of tries
func (t *OnecacheThrottler) IsRateLimited(r *http.Request) bool {
	key := t.keyGenerator(t.ipProvider.IP(r))

	if ok := t.store.Has(key); !ok {
		return false
	}

	buf, err := t.store.Get(key)

	//--->
	//Callers of this method expect a bool.
	//So we discard errors (or "convert them to booleans")
	//On encontering a non nil error, a falsy value is returned
	//A nil value is converted to a truthy value
	//Not too sure if this is right
	//but converting the return type to (bool, error) seem weird enough

	if err != nil {
		return false
	}

	item := new(throttledItem)

	if err = DecodeGob(buf, item); err != nil {
		return false
	}

	//The user must have made X requests in Y timeframe
	if item.Hits >= t.maxRequests &&
		time.Now().Sub(item.LastThrottledAt) <= t.interval {
		return true
	}

	return false
}

//Throttle throttles an HTTP request
func (t *OnecacheThrottler) Throttle(r *http.Request) error {

	if t.IsRateLimited(r) {
		return ErrClientIsRateLimited
	}

	key := t.keyGenerator(t.ipProvider.IP(r))

	if ok := t.store.Has(key); ok {

		buf, err := t.store.Get(key)

		if err != nil {
			return err
		}

		item := new(throttledItem)

		if err := DecodeGob(buf, item); err != nil {
			return err
		}

		item.LastThrottledAt = time.Now()
		item.Hits += defaultThrottledItemIncrement

		buf, err = EncodeGob(item)

		if err != nil {
			return err
		}

		if err := t.store.Set(key, buf, t.interval); err != nil {
			return err
		}

		return nil
	}

	item := &throttledItem{
		Hits: 1, LastThrottledAt: time.Now()}

	byt, err := EncodeGob(item)

	if err != nil {
		return err
	}

	if err = t.store.Set(key, byt, t.interval); err != nil {
		return err
	}

	return nil
}

//Clear resets the throttle on the request
func (t *OnecacheThrottler) Clear(r *http.Request) error {

	key := t.keyGenerator(t.ipProvider.IP(r))

	//It should be a no-op for requests that have not been throttled before
	if !t.store.Has(key) {
		return nil
	}

	if err := t.store.Delete(key); err != nil {
		return err
	}

	return nil
}

//Attempts returns the number of times the request have been throttled
func (t *OnecacheThrottler) Attempts(r *http.Request) (int, error) {

	key := t.keyGenerator(t.ipProvider.IP(r))

	if !t.store.Has(key) {
		return -1, errors.New(`
			gottle: Cannot get the number of attempts left as the current
			request has not been throttled or it has previously been cleared out`)
	}

	buf, err := t.store.Get(key)

	if err != nil {
		return -1, err
	}

	item := new(throttledItem)

	if err := DecodeGob(buf, item); err != nil {
		return -1, err
	}

	return item.Hits, nil
}

//Default implementation of KeyFunc
//Returns the ip as is...
//Library users might have a different implementation of this
func throttleKey(ip string) string {
	return ip
}

func EncodeGob(val *throttledItem) ([]byte, error) {

	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeGob(buf []byte, val *throttledItem) error {
	return gob.NewDecoder(bytes.NewBuffer(buf)).Decode(val)
}
