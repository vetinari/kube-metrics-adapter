package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/oliveagle/jsonpath"
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

const (
	HTTPMetricName            = "http"
	HTTPEndpointAnnotationKey = "endpoint"
	HTTPJsonPathAnnotationKey = "json-path"
)

type HTTPCollectorPlugin struct{}

func NewHTTPCollectorPlugin() (*HTTPCollectorPlugin, error) {
	return &HTTPCollectorPlugin{}, nil
}

func (p *HTTPCollectorPlugin) NewCollector(_ *v2beta2.HorizontalPodAutoscaler, config *MetricConfig, interval time.Duration) (Collector, error) {
	collector := &HTTPCollector{}
	var (
		value string
		ok    bool
	)
	if value, ok = config.Config[HTTPJsonPathAnnotationKey]; !ok {
		return nil, fmt.Errorf("config value %s not found", HTTPJsonPathAnnotationKey)
	}
	jsonPath, err := jsonpath.Compile(value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json path: %v", err)
	}
	collector.jsonPath = jsonPath
	if value, ok = config.Config[HTTPEndpointAnnotationKey]; !ok {
		return nil, fmt.Errorf("config value %s not found", HTTPEndpointAnnotationKey)
	}
	collector.endpoint = value
	collector.interval = interval
	collector.metricType = config.Type
	return collector, nil
}

type HTTPCollector struct {
	client     *http.Client
	endpoint   string
	jsonPath   *jsonpath.Compiled
	interval   time.Duration
	metricType v2beta2.MetricSourceType
}

func (c *HTTPCollector) GetMetrics() ([]CollectedMetric, error) {
	response, err := c.client.Get(c.endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unsuccessful response: %s", response.Status)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}
	res, err := c.jsonPath.Lookup(jsonData)
	if err != nil {
		return nil, err
	}
	var metricValue float64
	switch res := res.(type) {
	case int:
		metricValue = float64(res)
	case float32:
		metricValue = float64(res)
	case float64:
		metricValue = res
	default:
		return nil, fmt.Errorf("unsupported type %T", res)
	}
	value := CollectedMetric{
		Type: c.metricType,
		External: external_metrics.ExternalMetricValue{
			TypeMeta:      metav1.TypeMeta{},
			MetricName:    c.endpoint,
			MetricLabels:  nil,
			Timestamp:     metav1.Time{},
			WindowSeconds: nil,
			Value:         *resource.NewMilliQuantity(int64(metricValue*1000), resource.DecimalSI),
		},
	}
	return []CollectedMetric{value}, nil
}

func (c *HTTPCollector) Interval() time.Duration {
	return c.interval
}
