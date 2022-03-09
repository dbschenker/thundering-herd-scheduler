package nodestate

import (
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	"sync"
)

var (
	startingPodsMetric = metrics.NewGaugeVec(&metrics.GaugeOpts{
		Name:           "thundering_herd_scheduler_starting_pods",
		Help:           "Shows pods in starting phase per node",
		StabilityLevel: metrics.STABLE,
	}, []string{"node"})

	runningPodsMetric = metrics.NewGaugeVec(&metrics.GaugeOpts{
		Name:           "thundering_herd_scheduler_running_pods",
		Help:           "Shows pods in running phase per node",
		StabilityLevel: metrics.STABLE,
	}, []string{"node"})

	unhealthyPodsMetric = metrics.NewGaugeVec(&metrics.GaugeOpts{
		Name:           "thundering_herd_scheduler_unhealthy_pods",
		Help:           "Shows pods in unhealthy phase per node",
		StabilityLevel: metrics.STABLE,
	}, []string{"node"})
)

var registerMetricsOnce sync.Once

func init() {
	registerMetricsOnce.Do(func() {
		legacyregistry.MustRegister(startingPodsMetric)
		legacyregistry.MustRegister(runningPodsMetric)
		legacyregistry.MustRegister(unhealthyPodsMetric)
	})
}
