package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MultipassCollector struct {
	instanceTotal *prometheus.Desc
}

func NewMultipassCollector() *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instances_total",
			"Total de instancias de Multipass detectadas",
			nil, nil,
		),
	}
}

// Describe expone las métricas que vamos a implementar
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
}

// Collect devuelve los valores de las métricas
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	// De momento devolvemos un número fijo
	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		1,
	)
}
