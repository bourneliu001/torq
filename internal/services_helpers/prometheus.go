package services_helpers

import (
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lncapital/torq/build"
)

type Metrics struct {
	State            *prometheus.GaugeVec
	CoreServiceState *prometheus.GaugeVec
	NodeServiceState *prometheus.GaugeVec
}

func (m *Metrics) SetState(status ServiceStatus) {
	m.State.With(prometheus.Labels{"version": build.ExtendedVersion()}).Set(float64(status))
}

func (m *Metrics) SetCoreServiceState(serviceType ServiceType, status ServiceStatus) {
	m.CoreServiceState.With(prometheus.Labels{"serviceType": serviceType.String()}).Set(float64(status))
}

func (m *Metrics) SetNodeServiceState(serviceType ServiceType, nodeId int, status ServiceStatus) {
	m.NodeServiceState.With(
		prometheus.Labels{
			"serviceType": serviceType.String(),
			"nodeId":      strconv.Itoa(nodeId),
		}).Set(float64(status))
}

var (
	metricsOnce    sync.Once       //nolint:gochecknoglobals
	metricsWrapped *metricsWrapper //nolint:gochecknoglobals
)

type metricsWrapper struct {
	mu      sync.Mutex
	metrics *Metrics
}

func SetMetrics(m *Metrics) *Metrics {
	metricsOnce.Do(func() {
		metricsWrapped = &metricsWrapper{
			mu: sync.Mutex{},
		}
	})
	metricsWrapped.mu.Lock()
	defer metricsWrapped.mu.Unlock()
	metricsWrapped.metrics = m
	return m
}

func GetMetrics() *Metrics {
	metricsWrapped.mu.Lock()
	defer metricsWrapped.mu.Unlock()
	return metricsWrapped.metrics
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		State: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "torq",
				Name:      "state",
				Help:      "State of Torq itself",
			},
			[]string{"version"}),
		CoreServiceState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "torq",
				Name:      "coreServiceState",
				Help:      "State of a Torq core service",
			},
			[]string{"serviceType"}),
		NodeServiceState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "torq",
				Name:      "nodeServiceState",
				Help:      "State of a Torq node service",
			},
			[]string{"serviceType", "nodeId"}),
	}
	registry.MustRegister(m.State, m.CoreServiceState, m.NodeServiceState)
	return m
}
