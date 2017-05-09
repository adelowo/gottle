package gottle

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/adelowo/onecache"
)

const (
	//Set a long enough time for the data to expire
	//TODO (adelowo) : make use of `time.Duration(-1)` instead ?
	expirationTime                = time.Hour
	defaultThrottledItemIncrement = 1
)

//KeyFunc is a function type for setting the key in the cache
type KeyFunc func(ip string) string

//IPProvider provides an interface for fetching the IP of an HTTP request
type IPProvider interface {
	IP(r *http.Request) string
}

//Throttler defines the operation needed to limit clients
type Throttler interface {
	Throttle(r *http.Request) error
}

//OnecacheThrottler provides an implementation of Throttler by
//making use of onecache's cache implementation
type OnecacheThrottler struct {
	ipProvider   IPProvider
	store        onecache.Store
	keyGenerator KeyFunc
}

type throttledItem struct {
	ThrottledAt     time.Time //The first time the request was throttled
	LastThrottledAt time.Time //The most recent throttle time, so we can diff to lockout or not
	Hits            int
}

//Throttle throttles an HTTP request
func (t *OnecacheThrottler) Throttle(r *http.Request) error {

	ip := t.ipProvider.IP(r)

	key := t.keyGenerator(ip)

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
		item.Hits = item.Hits + defaultThrottledItemIncrement

		buf, err = EncodeGob(item)

		if err != nil {
			return err
		}

		if err := t.store.Set(key, buf, expirationTime); err != nil {
			return err
		}

		return nil
	}

	item := &throttledItem{
		Hits: 1, ThrottledAt: time.Now(), LastThrottledAt: time.Now()}

	byt, err := EncodeGob(item)

	if err != nil {
		return err
	}

	if err = t.store.Set(key, byt, expirationTime); err != nil {
		return err
	}

	return nil
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
