package prometheus

import (
	"sync"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	registryOnce    sync.Once        //nolint:gochecknoglobals
	registryWrapped *registryWrapper //nolint:gochecknoglobals
)

type registryWrapper struct {
	mu                sync.Mutex
	registry          *prometheus.Registry
	grpcClientMetrics *grpcprom.ClientMetrics
}

func SetRegistry(r *prometheus.Registry) *prometheus.Registry {
	registryOnce.Do(func() {
		registryWrapped = &registryWrapper{
			mu: sync.Mutex{},
			grpcClientMetrics: grpcprom.NewClientMetrics(
				grpcprom.WithClientHandlingTimeHistogram(
					grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
				),
			),
		}
		r.MustRegister(collectors.NewGoCollector())
		r.MustRegister(registryWrapped.grpcClientMetrics)
	})
	registryWrapped.mu.Lock()
	defer registryWrapped.mu.Unlock()
	registryWrapped.registry = r
	return r
}

func GetRegistry() *prometheus.Registry {
	registryWrapped.mu.Lock()
	defer registryWrapped.mu.Unlock()
	return registryWrapped.registry
}

func GetGrpcClientMetrics() *grpcprom.ClientMetrics {
	registryWrapped.mu.Lock()
	defer registryWrapped.mu.Unlock()
	return registryWrapped.grpcClientMetrics
}
