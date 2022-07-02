package collector

import (
	"github.com/jamessanford/omada-controller-exporter/omada"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	counter := prometheus.CounterValue
	gauge := prometheus.GaugeValue
	frombool := func(value bool) float64 {
		if value {
			return 1.0
		}
		return 0.0
	}

	// register every per-station metric
	register("wireless", "is wireless station", gauge,
		func(s sta) float64 { return frombool(s.Wireless) })
	register("channel", "wireless channel", gauge,
		func(s sta) float64 { return float64(s.Channel) })
	register("wireless_mode", "wireless mode in ?", gauge,
		func(s sta) float64 { return float64(s.WifiMode) })
	register("signal_rssi", "received signal strength in dBm", gauge,
		func(s sta) float64 { return float64(s.RSSI) })
	register("signal_level_pct", "signal level in percent", gauge,
		func(s sta) float64 { return float64(s.SignalLevel) })
	register("powersaving", "station is powersaving", gauge,
		func(s sta) float64 { return frombool(s.PowerSave) })

	register("transmit_bitrate_mbps", "transmit bitrate in Mbps", gauge,
		func(s sta) float64 { return float64(s.TxRate) / 1000.0 })
	register("receive_bitrate_mbps", "receive bitrate in Mbps", gauge,
		func(s sta) float64 { return float64(s.RxRate) / 1000.0 })
	register("transmit_bytes_total", "bytes transmitted to station", counter,
		func(s sta) float64 { return float64(s.TrafficDown) })
	register("receive_bytes_total", "bytes received from station", counter,
		func(s sta) float64 { return float64(s.TrafficUp) })
	register("transmit_packets_total", "packet count transmitted to station", counter,
		func(s sta) float64 { return float64(s.DownPacket) })
	register("receive_packets_total", "packet count received from station", counter,
		func(s sta) float64 { return float64(s.UpPacket) })

	register("last_seen_time_seconds", "station seen at this time in seconds since unix epoch", counter,
		func(s sta) float64 { return float64(s.LastSeen) / 1000.0 })
	register("uptime_seconds", "duration this station has been connected in seconds", counter,
		func(s sta) float64 { return float64(s.Uptime) })
}

var stationMetrics []*stationMetric

type stationMetric struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
	Value     stationValueReader
}

type sta *omada.ConnectedClient // convenience
type stationValueReader func(s sta) float64

func register(name, description string, valueType prometheus.ValueType, read stationValueReader) {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName("omada_station", "", name),
		description,
		[]string{"station", "name", "ap_address", "ap_name", "ssid", "site"},
		nil)

	sm := &stationMetric{desc, valueType, read}
	stationMetrics = append(stationMetrics, sm)
}
