package gottle

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ IPProvider = &RealIP{}
var _ IPProvider = &RemoteIP{}

func setUp(t *testing.T) (*http.Request, func(), error) {

	r := httptest.NewRequest(http.MethodGet, "/oops", nil)

	return r, func() { r.Body.Close() }, nil
}

func TestRealIP_IP(t *testing.T) {
	r, tearDown, err := setUp(t)

	defer tearDown()

	if err != nil {
		t.Fatalf("An error occurred while setting up the test... %v", err)
	}

	cases := []struct {
		Header, IP string
	}{
		{xForwardedFor, "111.222.333.444"},
		{xRealIP, "555.666.777.888"},
	}

	provider := NewRealIP()

	for _, v := range cases {
		r.Header.Set(v.Header, v.IP)

		if actual := provider.IP(r); actual != v.IP {
			t.Fatalf(`IPs don't match for %s...\n
				Expected %s, Got %s`, v.Header, v.IP, actual)
		}

		//Get rid of the header set.. This is as the implementation
		//would bail out on finding one of the two headers
		//This is to allow us test both cases
		r.Header.Del(v.Header)
	}
}

func TestRemoteIP(t *testing.T) {
	r, teardown, err := setUp(t)
	defer teardown()

	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		IP string
	}{
		{"111.222.333.444"},
	}

	provider := NewRemoteIP()

	for _, v := range cases {
		r.RemoteAddr = v.IP

		if actual := provider.IP(r); v.IP != actual {
			t.Fatalf(`
				IP fetched from the RemoteAddr differ\n
				Expected %s.. Got %s`, v.IP, actual)
		}
	}
}
