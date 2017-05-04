This is the `go-metrics` exporter for Pivotal Cloud Foundry (PCF) Metrics.

It is designed to work with the [go-metrics](https://github.com/rcrowley/go-metrics) library, which is a Go port of [Dropwizard metrics](https://github.com/dropwizard/metrics).


## 1. Download and install the Go packages

```
go get github.com/rcrowley/go-metrics
go get github.com/pivotal-cf/go-metrics-pcf
```

## 2. Import the metrics libraries into your application

Add these import lines to your `main.go`:

```
import (
    "time"
    
    // add both of these lines to your application
    "github.com/pivotal-cf/go-metrics-pcf"
    "github.com/rcrowley/go-metrics"
)
```

You'll also want to include the `go-metrics` import in any other places where you collect metrics.

## 3. Start the exporter inside your application 

This only needs to be called once, so find a convenient place to run this when your application starts.

```
func main() {
    // ... other code here
    
    // start exporting to PCF every minute
    go pcfmetrics.StartExporter(metrics.DefaultRegistry)

    // ...
}
```

## 4. Add desired instrumentation elsewhere in your code

Remember to import the `github.com/rcrowley/go-metrics` package everywhere you want to add this instrumentation.

```
    timer := metrics.GetOrRegisterTimer("sample-timer", metrics.DefaultRegistry)
    
    timer.Time(func() {
        time.Sleep(time.Second)
    })
```

## 5. Create and bind the CF Metrics service to your application

This will bind your application to the metrics service and make sure credentials are available for the metrics API.

```
$ cf create-service cf-metrics standard my-metrics-service

$ cf bind-service my-app my-metrics-service

$ cf restage my-app
```
