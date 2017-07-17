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
	if o.Token == "" {
		apiToken, err := getToken(o.ServiceName)
		if err != nil {
			log.Printf("Could not get apiToken: %s", err.Error())
			return
		}

		o.Token = apiToken
	}

	if o.Url == "" {
		apiUrl, err := getUrl(o.ServiceName)
		if err != nil {
			log.Printf("Could not get Url: %s", err.Error())
			return
		}

		o.Url = apiUrl
	}

	if o.AppGuid == "" {
		appGuid, err := getAppGuid()
		if err != nil {
			log.Printf("Could not get app guid: %s", err.Error())
			return
		}

		o.AppGuid = appGuid
	}
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
func WithInstanceIndex(id string) ExporterOption {
	return func(o *Options) {
		o.InstanceId = id
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
