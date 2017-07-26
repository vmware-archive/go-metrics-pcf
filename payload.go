package pcfmetrics

type metricForwarderPayload struct {
	Applications []*metricForwarderApplication `json:"applications"`
}

func newMetricForwarderPayload(points []*dataPoint, options *Options) *metricForwarderPayload {
	return &metricForwarderPayload{
		Applications: []*metricForwarderApplication{
			newMetricForwarderApplication(points, options),
		},
	}
}

type metricForwarderApplication struct {
	Id        string `json:"id"`
	Instances []*metricForwarderInstance `json:"instances"`
}

func newMetricForwarderApplication(points []*dataPoint, options *Options) *metricForwarderApplication {
	return &metricForwarderApplication{
		Id: options.AppGuid,
		Instances: []*metricForwarderInstance{
			newMetricForwarderInstance(points, options),
		},
	}
}

type metricForwarderInstance struct {
	Id      string `json:"id"`
	Index   string `json:"index"`
	Metrics []*dataPoint `json:"metrics"`
}

func newMetricForwarderInstance(points []*dataPoint, options *Options) *metricForwarderInstance {
	return &metricForwarderInstance{
		Id:      options.InstanceId,
		Index:   options.InstanceIndex,
		Metrics: points,
	}
}