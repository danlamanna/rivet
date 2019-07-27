package girder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"io/ioutil"
	"log"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

func addBaseHeaders(ctx *Context, request *retryablehttp.Request) {
	request.Header.Add("User-Agent", "rivet/0.0.1")
	request.Header.Add("Girder-Token", ctx.Auth)
}

func urlFromRequest(request *http.Request) string {
	return fmt.Sprintf("%s://%s%s?%s", request.URL.Scheme,
		request.URL.Host, request.URL.Path, request.URL.RawQuery)

}

func logRequest(ctx *Context) func(retryablehttp.Logger, *http.Request, int) {
	return func(logger retryablehttp.Logger, request *http.Request, tryNumber int) {
		if tryNumber > 0 {
			ctx.Logger.Warnf("retrying (attempt %d/5): %s\n", tryNumber+1, urlFromRequest(request))
		}
	}
}

func decodeResponse(resp *http.Response, success, failure interface{}) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		return json.NewDecoder(resp.Body).Decode(success)
	}
	return json.NewDecoder(resp.Body).Decode(failure)
}

// Get does stuff
func Get(ctx *Context, url string, success interface{}, failure interface{}) (*http.Response, error) {
	client := retryablehttp.NewClient()
	if ctx.Logger.Level == logrus.InfoLevel {
		client.Logger = log.New(ioutil.Discard, "", 0)
		client.RequestLogHook = logRequest(ctx)
	}
	request, err := retryablehttp.NewRequest("GET", fmt.Sprintf("%s/%s", ctx.URL, url), nil)

	if err != nil {
		return nil, err
	}

	addBaseHeaders(ctx, request)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	decodeResponse(response, success, failure)

	return response, nil
}

// GetBasicAuth does stuff
func GetBasicAuth(ctx *Context, auth string, url string, success interface{}, failure interface{}) (*http.Response, error) {
	client := retryablehttp.NewClient()
	if ctx.Logger.Level == logrus.InfoLevel {
		client.Logger = log.New(ioutil.Discard, "", 0)
		client.RequestLogHook = logRequest(ctx)
	}
	request, err := retryablehttp.NewRequest("GET", fmt.Sprintf("%s/%s", ctx.URL, url), nil)

	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", "rivet/0.0.1")
	authHeader := base64.StdEncoding.EncodeToString([]byte(auth))
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", authHeader))

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	decodeResponse(response, success, failure)

	return response, nil
}

// Post does stuff
func Post(ctx *Context, url string, rawBody interface{}, success interface{}, failure *GirderError) (*http.Response, error) {
	client := retryablehttp.NewClient()
	client.RetryWaitMax = time.Millisecond * 100
	if ctx.Logger.Level == logrus.InfoLevel {
		client.Logger = log.New(ioutil.Discard, "", 0)
		client.RequestLogHook = logRequest(ctx)
	}
	request, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/%s", ctx.URL, url), rawBody)

	if err != nil {
		return nil, err
	}

	addBaseHeaders(ctx, request)

	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	decodeResponse(response, success, failure)
	if response.StatusCode != 200 && failure != nil {
		ctx.Logger.Warn(failure.Message)
	}

	return response, nil
}

// Put does stuff
func Put(ctx *Context, url string, rawBody interface{}, success interface{}, failure interface{}) (*http.Response, error) {
	client := retryablehttp.NewClient()
	if ctx.Logger.Level == logrus.InfoLevel {
		client.Logger = log.New(ioutil.Discard, "", 0)
		client.RequestLogHook = logRequest(ctx)
	}
	request, err := retryablehttp.NewRequest("PUT", fmt.Sprintf("%s/%s", ctx.URL, url), rawBody)

	if err != nil {
		return nil, err
	}

	addBaseHeaders(ctx, request)

	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	decodeResponse(response, success, failure)
	return response, nil
}
