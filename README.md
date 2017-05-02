This is the `go-metrics` exporter for Pivotal Cloud Foundry.

It is designed to work with the [go-metrics](https://github.com/rcrowley/go-metrics) library, which is a Go port of [Dropwizard metrics](https://github.com/dropwizard/metrics).

Adding the exporter to your app
-------------------------------

Add to your main.go:

```
import (
    "time"
    
    "github.com/pivotal-cf/go-metrics-pcf"
    "github.com/rcrowley/go-metrics"
)

func main() {
    // start exporting to PCF every minute
    go pcf.Pcf(metrics.DefaultRegistry)
    ...

    timer := metrics.GetOrRegisterTimer("sample-timer", metrics.DefaultRegistry)
    
    timer.Time(func() {
        time.Sleep(time.Second)
    })
}
```

Creating and Binding the service
--------------------------------

```
cf create-service cf-metrics standard my-metrics-service
cf bind-service my-app my-metrics-service
cf restage my-app
```
