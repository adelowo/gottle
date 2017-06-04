
### Gottle - A rate limiter for Golang which builds on Onecache

[![Coverage Status](https://coveralls.io/repos/github/adelowo/gottle/badge.svg)](https://coveralls.io/github/adelowo/gottle)
[![Build Status](https://img.shields.io/travis/adelowo/gottle/master.svg?style=flat-square)](https://travis-ci.org/adelowo/gottle.svg?branch=master)


- [Installation](#install)
- [Usage](#usage)
- [How it works](#works)

<div id="install"> </div>

```bash
$ go get -u github.com/adelowo/gottle
```

<div id="usage"></div>

> Docs for all available operations -> https://godoc.org/github.com/adelowo/gottle or run `godoc github.com/adelowo/gottle`


```go

package main

func main() {

  func someHandler(w http.ResponseWriter, r *http.Request) {

    var throttler = NewOneCacheThrottler(nil) //can pass in options instead of nil

    //Throttle the request
    if err := throttler.Throttle(r); err != nil {
      sendFailureResponse(500, "An error occurred while trying to throttle the request")
      return
    }

    //Can also check if the current request has been rate limited
    if throttler.IsRateLimited(r) {
      //We only want to throttle requests from other account types
      //aside premium users
      if isPremiumUser(r) {
        //isPremiumUser fetches the account type from the context
        throttler.Clear(r) //Clear the rate limit
        someOperationUserWantsToCarryOut(r)
        return
      }

      sendFailureResponse(400, `
        You have been ratelimited.. Wait for some time before trying again`)
      return
    }

  }

}
```

<div id="works"> </div>

This is a very simple throttler implementation (albeit it works very well). All it does is keep a record of the IP of a request and the number of times a request was received from that IP. Once the request count has passed it's limit, a lockout is obtained

There are IPProviders, which extract out the IP from a request.. There are currently two implementations you can choose from

- `RealIP` - This fetches the IP from the HTTP headers (`X-Forwarded-For` or `X-Real-IP`).. This is suitable when you have a reverse proxy to your go binary.

- `RemoteIP` - extremely basic and not guareented to work as expected because Go sets the `RemoteAddr` of a request to `IP:port` and you are expected to manipulate that yourself in a middleware or something of that sort.

You can also write your own IPProvider by implementing ;

```go
IP(r *http.Request) string
```

You can also configure the amount of requests a client is allowed to make before a lockout is obtained.. By default, it is at 10 requests in a timeframe of 10 minutes. To override that, you simply set an option on `NewOneCacheThrottler` like :

```go

interval := time.Minute
maxRequests := 60

throttler := NewOneCacheThrottler(
  ThrottleCondition(interval, maxRequests))

//Now a lockout would only be obtained after the client has made 60 requests in less than a minute

```

> Do check the  other available options in the [godoc](https://godoc.org/github.com/adelowo/gottle) or the test suites


### License
MIT
