// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package pcfmetrics

import (
	"time"
	"log"
)

// Options is used when starting an exporter.
type Options struct {
	Frequency           time.Duration
	InstanceId          string
	InstanceIndex       string
	Token               string
	Url                 string
	AppGuid             string
	TimeUnit            time.Duration
	ServiceName         string
	SkipSSLVerification bool
}

func (o *Options) fillDefaults() {
	if o.Token == "" || o.Url == "" {
		o.fillCredentialDefaults()
	}

	if o.AppGuid == "" {
		o.fillAppGuidDefault()
	}

	if o.Frequency == time.Duration(0) {
		o.Frequency = time.Minute
	}
}

func (o *Options) fillCredentialDefaults() {
	creds, err := getCredentials(o.ServiceName)
	if err != nil {
		log.Printf("Could not get metrics forwarder credentials: %s", err.Error())
		return
	}

	if o.Token == "" {
		o.Token = creds.AccessToken
	}

	if o.Url == "" {
		o.Url = creds.Url
	}
}

func (o *Options) fillAppGuidDefault() {
	appGuid, err := getAppGuid()
	if err != nil {
		log.Printf("Could not get app guid: %s", err.Error())
		return
	}

	o.AppGuid = appGuid
}

// ExporterOption is used to configure an exporter.
type ExporterOption func(*Options)

// WithFrequency sets the frequency. The default is a minute.
func WithFrequency(f time.Duration) ExporterOption {
	return func(o *Options) {
		o.Frequency = f
	}
}

// WithInstanceId sets the instance ID. The default is read from the
// environment variable INSTANCE_GUID
func WithInstanceId(guid string) ExporterOption {
	return func(o *Options) {
		o.InstanceId = guid
	}
}

// WithInstanceIndex sets the instance index. The default is read from the
// environment variable INSTANCE_INDEX.
func WithInstanceIndex(index string) ExporterOption {
	return func(o *Options) {
		o.InstanceIndex = index
	}
}

// WithToken sets the token.
func WithToken(token string) ExporterOption {
	return func(o *Options) {
		o.Token = token
	}
}

// WithURL sets the URL.
func WithURL(URL string) ExporterOption {
	return func(o *Options) {
		o.Url = URL
	}
}

// WithAppGuid sets the AppGuid.
func WithAppGuid(guid string) ExporterOption {
	return func(o *Options) {
		o.AppGuid = guid
	}
}

// WithTimeUnit sets the TimeUnit.
func WithTimeUnit(u time.Duration) ExporterOption {
	return func(o *Options) {
		o.TimeUnit = u
	}
}

// WithServiceName sets the ServiceName.
func WithServiceName(n string) ExporterOption {
	return func(o *Options) {
		o.ServiceName = n
	}
}

// WithSkipSSL sets the InsecureSkipVerify flag on the HTTP transport.
func WithSkipSSL(skip bool) ExporterOption {
	return func(o *Options) {
		o.SkipSSLVerification = skip
	}
}
