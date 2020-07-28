/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http

import (
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	types "mayadata.io/d-operators/types/http"
)

// InvocableConfig is used to initialise a new instance of
// Invocable
type InvocableConfig struct {
	Username    string
	Password    string
	HTTPMethod  string
	URL         string
	Body        string
	Headers     map[string]string
	QueryParams map[string]string
	PathParams  map[string]string
}

// Invocable helps invoking a http request
type Invocable struct {
	Username    string
	Password    string
	HTTPMethod  string
	URL         string
	Body        string
	Headers     map[string]string
	QueryParams map[string]string
	PathParams  map[string]string
}

// Invoker returns a new instance of Invocable
func Invoker(config InvocableConfig) *Invocable {
	return &Invocable{
		Username:    config.Username,
		Password:    config.Password,
		HTTPMethod:  config.HTTPMethod,
		URL:         config.URL,
		Body:        config.Body,
		Headers:     config.Headers,
		PathParams:  config.PathParams,
		QueryParams: config.QueryParams,
	}
}

func (i *Invocable) buildStatus(response *resty.Response) types.HTTPResponse {
	var resp = &types.HTTPResponse{}
	// set http response details
	if response != nil {
		resp.Body = response.Result()
		resp.HTTPStatusCode = response.StatusCode()
		resp.HTTPStatus = response.Status()
		resp.HTTPError = response.Error()
		resp.IsError = response.IsError()
	}
	return *resp
}

// Invoke executes the http request
func (i *Invocable) Invoke() (types.HTTPResponse, error) {
	req := resty.New().R().
		SetBody(i.Body).
		SetHeaders(i.Headers).
		SetQueryParams(i.QueryParams).
		SetPathParams(i.PathParams)

	// set credentials only if it was provided
	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	var response *resty.Response
	var err error

	switch strings.ToUpper(i.HTTPMethod) {
	case types.POST:
		response, err = req.Post(i.URL)
	case types.GET:
		response, err = req.Get(i.URL)
	default:
		err = errors.Errorf(
			"HTTP method not supported: URL %q: Method %q",
			i.URL,
			i.HTTPMethod,
		)
	}

	if err != nil {
		return types.HTTPResponse{}, err
	}
	return i.buildStatus(response), nil
}
