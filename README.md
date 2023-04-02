Omada Controller metrics exporter for connected wifi stations

Tested against Omada Controller 5.9 with EAP245v3 access points.

Exporter for Prometheus https://prometheus.io/

#### Install

    go install github.com/jamessanford/omada-controller-exporter@latest

#### Run

May be configured with environment variables:

    OMADA_PATH=https://192.168.255.123:8043/ OMADA_USER=admin OMADA_PASS=foo omada-controller-exporter

Or a config file, see [config.yaml](config.yaml)

#### Metrics exported

Sample metrics for a single station:

```
omada_collector_requests_total 123
omada_collector_errors_total 0
omada_collector_duration_seconds 0.004
omada_station_channel{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 40
omada_station_last_seen_time_seconds{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 1.609035439941e+09
omada_station_powersaving{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 0
omada_station_receive_bytes_total{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 9.14206521e+08
omada_station_receive_packets_total{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 2.428055e+06
omada_station_receive_bitrate_mbps{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 780
omada_station_signal_rssi{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} -36
omada_station_signal_level_pct{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 100
omada_station_transmit_bytes_total{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 1.0260304329e+10
omada_station_transmit_packets_total{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 7.841603e+06
omada_station_transmit_bitrate_mbps{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 866
omada_station_uptime_seconds{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 260161
omada_station_wireless{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 1
omada_station_wireless_mode{ap_address="54:54:54:00:00:03",ap_name="AP-Couch",name="iPad",ssid="HappyHome5",station="86:4c:8d:00:00:00"} 5
```

