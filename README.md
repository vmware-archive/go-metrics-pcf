This is the `go-metrics` exporter for Pivotal Cloud Foundry.


Adding the exporter to your app
-------------------------------

Add to your main.go:

```
func main() {
    go pcf.Pcf(metrics.DefaultRegistry)
    ...
}
```

Creating and Binding the service
--------------------------------

```
cf create-service cf-metrics standard my-metrics-service
cf bind-service my-app my-metrics-service
cf restage my-app
```