package stats

import (
	"fmt"
	"os"
	"time"

	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	FilerGather        = prometheus.NewRegistry()
	VolumeServerGather = prometheus.NewRegistry()

	FilerRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "SeaweedFS",
			Subsystem: "filer",
			Name:      "request_total",
			Help:      "Counter of filer requests.",
		}, []string{"type"})

	FilerRequestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "SeaweedFS",
			Subsystem: "filer",
			Name:      "request_seconds",
			Help:      "Bucketed histogram of filer request processing time.",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 24),
		}, []string{"type"})

	VolumeServerRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "SeaweedFS",
			Subsystem: "volumeServer",
			Name:      "request_total",
			Help:      "Counter of filer requests.",
		}, []string{"type"})

	VolumeServerRequestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "SeaweedFS",
			Subsystem: "volumeServer",
			Name:      "request_seconds",
			Help:      "Bucketed histogram of filer request processing time.",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 24),
		}, []string{"type"})

	VolumeServerVolumeCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "SeaweedFS",
			Subsystem: "volumeServer",
			Name:      "volumes",
			Help:      "Number of volumes or shards.",
		}, []string{"collection", "type"})

	VolumeServerMaxVolumeCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "SeaweedFS",
			Subsystem: "volumeServer",
			Name:      "volumes",
			Help:      "Maximum number of volumes.",
		})

	VolumeServerDiskSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "SeaweedFS",
			Subsystem: "volumeServer",
			Name:      "total_disk_size",
			Help:      "Actual disk size used by volumes.",
		}, []string{"collection", "type"})
)

func init() {

	FilerGather.MustRegister(FilerRequestCounter)
	FilerGather.MustRegister(FilerRequestHistogram)

	VolumeServerGather.MustRegister(VolumeServerRequestCounter)
	VolumeServerGather.MustRegister(VolumeServerRequestHistogram)
	VolumeServerGather.MustRegister(VolumeServerVolumeCounter)
	VolumeServerGather.MustRegister(VolumeServerMaxVolumeCounter)
	VolumeServerGather.MustRegister(VolumeServerDiskSizeGauge)

}

func LoopPushingMetric(name, instance string, gatherer *prometheus.Registry, fnGetMetricsDest func() (addr string, intervalSeconds int)) {

	addr, intervalSeconds := fnGetMetricsDest()
	pusher := push.New(addr, name).Gatherer(gatherer).Grouping("instance", instance)
	currentAddr := addr

	for {
		if currentAddr != "" {
			err := pusher.Push()
			if err != nil {
				glog.V(0).Infof("could not push metrics to prometheus push gateway %s: %v", addr, err)
			}
		}
		if intervalSeconds <= 0 {
			intervalSeconds = 15
		}
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
		addr, intervalSeconds = fnGetMetricsDest()
		if currentAddr != addr {
			pusher = push.New(addr, name).Gatherer(gatherer).Grouping("instance", instance)
			currentAddr = addr
		}

	}
}

func SourceName(port int) string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", hostname, port)
}
