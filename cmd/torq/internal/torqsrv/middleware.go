package torqsrv

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

	prometheus2 "github.com/lncapital/torq/pkg/prometheus"
)

const defaultMetricPath = "/metrics"

//https://raw.githubusercontent.com/wei840222/gin-prometheus/master/middleware.go

/*
RequestCounterURLLabelMappingFn is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:

	func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, p := range c.Params {
			if p.Key == "name" {
				url = strings.Replace(url, p.Value, ":name", 1)
				break
			}
		}
		return url
	}

which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Metric is a definition for the name, description, type, ID, and
// prometheus.Collector type (i.e. CounterVec, Summary, etc) of each metric
type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt        *prometheus.CounterVec
	reqDur        *prometheus.HistogramVec
	reqSz, resSz  prometheus.Summary
	router        *gin.Engine
	listenAddress string
	Ppg           PrometheusPushGateway

	MetricsList []*Metric
	MetricsPath string

	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn

	// gin.Context string to use as a prometheus URL label
	URLLabelFromContext string

	EnableExemplar bool
}

// PrometheusPushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PrometheusPushGateway struct {

	// Push interval
	PushInterval time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// Local metrics URL where metrics are fetched from, this could be ommited in the future
	// if implemented using prometheus common/expfmt instead
	MetricsURL string

	// pushgateway job name, defaults to "gin"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string, customMetricsList ...[]*Metric) *Prometheus {

	var metricsList []*Metric

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus( string, <optional []*Metric> ).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	// Standard default metrics
	//
	//	counter, counter_vec, gauge, gauge_vec,
	//	histogram, histogram_vec, summary, summary_vec
	var reqCnt = &Metric{
		ID:          "reqCnt",
		Name:        "requests_total",
		Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		Type:        "counter_vec",
		Args:        []string{"code", "method", "handler", "host", "url"}}

	var reqDur = &Metric{
		ID:          "reqDur",
		Name:        "request_duration_seconds",
		Description: "The HTTP request latencies in seconds.",
		Type:        "histogram_vec",
		Args:        []string{"code", "method", "url"},
	}

	var resSz = &Metric{
		ID:          "resSz",
		Name:        "response_size_bytes",
		Description: "The HTTP response sizes in bytes.",
		Type:        "summary"}

	var reqSz = &Metric{
		ID:          "reqSz",
		Name:        "request_size_bytes",
		Description: "The HTTP request sizes in bytes.",
		Type:        "summary"}

	var standardMetrics = []*Metric{
		reqCnt,
		reqDur,
		resSz,
		reqSz,
	}

	metricsList = append(metricsList, standardMetrics...)

	p := &Prometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path // i.e. by default do nothing, i.e. return URL as is
		},
	}

	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		prometheus2.GetRegistry().MustRegister(metric)
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(prometheus.Summary)
		case reqSz:
			p.reqSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}

	return p
}

// SetPushGateway sends metrics to a remote pushgateway exposed on pushGatewayURL
// every pushInterval. Metrics are fetched from metricsURL
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushInterval time.Duration) *Prometheus {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.MetricsURL = metricsURL
	p.Ppg.PushInterval = pushInterval
	p.startPushTicker()
	return p
}

// SetPushGatewayJob job name, defaults to "gin"
func (p *Prometheus) SetPushGatewayJob(j string) *Prometheus {
	p.Ppg.Job = j
	return p
}

// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
// same address of the gin engine that is being used
func (p *Prometheus) SetListenAddress(address string) *Prometheus {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.New()
		p.router.Use(gin.Recovery())
	}
	return p
}

// SetListenAddressWithRouter for using a separate router to expose metrics. (this keeps things like GET /metrics out of
// your content's access log).
func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) *Prometheus {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
	return p
}

// SetMetricsPath set metrics paths
func (p *Prometheus) SetMetricsPath(e *gin.Engine) *Prometheus {
	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, p.prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, p.prometheusHandler())
	}
	return p
}

// SetMetricsPathWithAuth set metrics paths with authentication
func (p *Prometheus) SetMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) *Prometheus {
	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, gin.BasicAuth(accounts), p.prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, gin.BasicAuth(accounts), p.prometheusHandler())
	}
	return p
}

// SetEnableExemplar set enable exemplar feature
func (p *Prometheus) SetEnableExemplar(enableExemplar bool) *Prometheus {
	p.EnableExemplar = enableExemplar
	return p
}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go func() {
			err := p.router.Run(p.listenAddress)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to boot prometheus at %v", p.listenAddress)
			}
		}()
	}
}

func (p *Prometheus) getMetrics() []byte {
	res, err := http.Get(p.Ppg.MetricsURL)
	if err != nil {
		log.Printf("Error get local metrics: %s", err)
		return nil
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "gin"
	}
	return p.Ppg.PushGatewayURL + "/metrics/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	if len(metrics) == 0 {
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	if err != nil {
		log.Printf("Error build push gateway request: %s", err)
		return
	}
	if _, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("Error sending to push gateway: %s", err)
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Ppg.PushInterval)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

// NewMetric associates prometheus.Collector based on Metric.Type
func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
}

// UseWithAuth adds the middleware to a gin engine with BasicAuth.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPathWithAuth(e, accounts)
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	if p.EnableExemplar {
		return func(c *gin.Context) {
			if c.Request.URL.Path == p.MetricsPath {
				c.Next()
				return
			}

			start := time.Now()
			reqSz := computeApproximateRequestSize(c.Request)

			c.Next()

			status := strconv.Itoa(c.Writer.Status())
			elapsed := float64(time.Since(start)) / float64(time.Second)
			resSz := float64(c.Writer.Size())

			url := p.ReqCntURLLabelMappingFn(c)
			// jlambert Oct 2018 - sidecar specific mod
			if len(p.URLLabelFromContext) > 0 {
				u, found := c.Get(p.URLLabelFromContext)
				if !found {
					u = "unknown"
				}
				url = u.(string)
			}

			traceID := trace.SpanContextFromContext(c.Request.Context()).TraceID()
			if exemplarObserver, ok := p.reqDur.WithLabelValues(status, c.Request.Method, url).(prometheus.ExemplarObserver); ok && traceID.IsValid() {
				exemplarObserver.ObserveWithExemplar(elapsed, prometheus.Labels{"traceID": traceID.String()})
			} else {
				p.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
			}

			if exemplarAdder, ok := p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).(prometheus.ExemplarAdder); ok && traceID.IsValid() {
				exemplarAdder.AddWithExemplar(1, prometheus.Labels{"traceID": traceID.String()})
			} else {
				p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
			}

			p.reqSz.Observe(float64(reqSz))
			p.resSz.Observe(resSz)
		}
	}
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.ReqCntURLLabelMappingFn(c)
		// jlambert Oct 2018 - sidecar specific mod
		if len(p.URLLabelFromContext) > 0 {
			u, found := c.Get(p.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}
		p.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}
}

func (p *Prometheus) prometheusHandler() gin.HandlerFunc {
	if p.EnableExemplar {
		h := promhttp.HandlerFor(prometheus2.GetRegistry(), promhttp.HandlerOpts{
			Registry:          prometheus2.GetRegistry(),
			EnableOpenMetrics: true,
		})
		return func(c *gin.Context) {
			h.ServeHTTP(c.Writer, c.Request)
		}
	}

	h := promhttp.HandlerFor(prometheus2.GetRegistry(), promhttp.HandlerOpts{Registry: prometheus2.GetRegistry()})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
