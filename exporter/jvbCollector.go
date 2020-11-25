/*
 *  Copyright 2019 karriere tutor GmbH
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  	http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//statsSet describes the stats belonging to a jvb instance
//lastUpdated: time of last update
//stats: as the are unmarshalled by the PresExtension
//jvbIdentifier: will be attached as tag to metric to identify individual JVBs
type statsSet struct {
	lastUpdated   time.Time
	stats         Stats
	jvbIdentifier string
}

type metric struct {
	name       string
	desc       *prometheus.Desc
	metricType prometheus.ValueType
}

func newMetric(name string, metricType prometheus.ValueType, help string,
	varLabels []string, constLabels prometheus.Labels) metric {

	var metric = metric{
		name:       name,
		metricType: metricType,
	}
	metric.desc = prometheus.NewDesc(name, help, varLabels, constLabels)
	return metric
}

//JvbCollector collects metrics for jitsi JVBs
//NamePrefix for naming the metrics, see https://godoc.org/github.com/prometheus/client_golang/prometheus#Opts
//Retention defines how long the jvb collector will consider a set of stats valid, once retention has passed since the last update,
//	the stats set will not be included in the collect output anymore
type JvbCollector struct {
	NamePrefix string
	Retention  time.Duration
	statsSets  []statsSet
	metrics    []metric
}

//NewJvbCollector initializes a Jvb collector
//namespace and subsystem may be empty if you dont need them, see https://godoc.org/github.com/prometheus/client_golang/prometheus#Opts
func NewJvbCollector(namespace, subsystem string, retention time.Duration) *JvbCollector {
	var collector = &JvbCollector{
		Retention: retention,
	}

	var namePrefix = ""
	if subsystem != "" {
		namePrefix += subsystem
		namePrefix += "_"
	}

	if namespace != "" {
		namePrefix += namespace
		namePrefix += "_"
	}

	collector.NamePrefix = namePrefix

	var constLabels = prometheus.Labels{
		"app": "jitsi",
	}

	//add metrics
	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"packet_rate_download", prometheus.GaugeValue,
		"download packet rate", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"conference_sizes", prometheus.UntypedValue,
		"histogram of conference sizes (ie. how many conferences have 5 participants and so on)", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_packets_sent_octo", prometheus.CounterValue,
		"total number of octo packets sent", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_loss_degraded_participant_seconds", prometheus.CounterValue,
		"The total number of participant-seconds that are loss-degraded.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"bit_rate_download", prometheus.GaugeValue,
		"download rate kbit/s", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"jitter_aggregate", prometheus.GaugeValue,
		"Experimental. An average value (in milliseconds) of the jitter calculated for incoming and outgoing streams. This hasn't been tested and it is currently not known whether the values are correct or not.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_packets_received", prometheus.CounterValue,
		"Total number of packets received", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"rtt_aggregate", prometheus.GaugeValue,
		"An average value (in milliseconds) of the RTT across all streams.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"packet_rate_upload", prometheus.GaugeValue,
		"Upload packets/s", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"conferences", prometheus.GaugeValue,
		"The current number of conferences hosted by the bridge", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"participants", prometheus.GaugeValue,
		"The current number of participants.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_loss_limited_participant_seconds", prometheus.CounterValue,
		"The total number of participant-seconds that are loss-limited.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"largest_conference", prometheus.GaugeValue,
		"The current number of participants in the largest conference", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_packets_sent", prometheus.CounterValue,
		"The total number of packets sent.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_data_channel_messages_sent", prometheus.CounterValue,
		"The total number of data channel messages sent.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_bytes_received_octo", prometheus.CounterValue,
		"The total number octo bytes sent.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"threads", prometheus.GaugeValue,
		"The current number of threads.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_colibri_web_socket_messages_received", prometheus.CounterValue,
		"The total number messages received through COLIBRI web sockets.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"videochannels", prometheus.GaugeValue,
		"The current number of videochannels.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_packets_received_octo", prometheus.CounterValue,
		"Total octo packets received.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_colibri_web_socket_messages_sent", prometheus.CounterValue,
		"The total number messages sent through COLIBRI web sockets.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_bytes_sent_octo", prometheus.CounterValue,
		"Total octo bytes sent.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_data_channel_messages_received", prometheus.CounterValue,
		"Total data channel messages received.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_conference_seconds", prometheus.CounterValue,
		"The sum of the lengths of all completed conferences, in seconds.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_bytes_received", prometheus.CounterValue,
		"Total bytes received.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_loss_controlled_participant_seconds", prometheus.CounterValue,
		"The total number of participant-seconds that are loss-controlled.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_partially_failed_conferences", prometheus.CounterValue,
		"The total number of partially failed conferences on the bridge. A conference is marked as partially failed when some of its channels has failed. A channel is marked as failed if it had no payload activity.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"bit_rate_upload", prometheus.GaugeValue,
		"Current upload rate in kbit/s.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_conferences_completed", prometheus.CounterValue,
		"Total conferences completed.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_bytes_sent", prometheus.CounterValue,
		"The number of total bytes sent.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_failed_conferences", prometheus.CounterValue,
		"The total number of failed conferences on the bridge. A conference is marked as failed when all of its channels have failed. A channel is marked as failed if it had no payload activity.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"conferences_by_audio_senders", prometheus.UntypedValue,
		"Histogram of conferences by number of audio senders (ie. how many conferences have 5 audio senders and so on)", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"conferences_by_video_senders", prometheus.UntypedValue,
		"Histogram of conferences by number of video senders (ie. how many conferences have 5 video senders and so on)", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"dtls_failed_endpoints", prometheus.GaugeValue,
		"The number of failed dtls endpoints on the bridge. An endpoint has failed DTLS if it has completed ICE but not.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"endpoints_sending_audio", prometheus.GaugeValue,
		"The number of endpoints which are sending audio.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"endpoints_sending_video", prometheus.GaugeValue,
		"The number of endpoints which are sending video.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"inactive_conferences", prometheus.GaugeValue,
		"The number of inactive conferences.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"inactive_endpoints", prometheus.GaugeValue,
		"The number of inactive endpoints.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"incoming_loss", prometheus.GaugeValue,
		"The percentage of incoming packets which are lost.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"muc_clients_configured", prometheus.GaugeValue,
		"The number of configured muc clients.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"muc_clients_connected", prometheus.GaugeValue,
		"The number of connected muc clients.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"mucs_configured", prometheus.GaugeValue,
		"The number of configured mucs.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"mucs_joined", prometheus.GaugeValue,
		"The number of joined mucs.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"num_eps_no_msg_transport_after_delay", prometheus.GaugeValue,
		"The number of endpoints with no message transport after delay.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_conferences", prometheus.GaugeValue,
		"The number of conferences using Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_endpoints", prometheus.GaugeValue,
		"The number of endpoints using Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_receive_bitrate", prometheus.GaugeValue,
		"The bitrate of data being received from Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_receive_packet_rate", prometheus.GaugeValue,
		"The rate of packets being received from Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_send_bitrate", prometheus.GaugeValue,
		"The bitrate of data being send to Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"octo_send_packet_rate", prometheus.GaugeValue,
		"The rate of packets being send to Octo", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"outgoing_loss", prometheus.GaugeValue,
		"The percentage of outgoing packets which are lost.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"overall_loss", prometheus.GaugeValue,
		"The overall percentage of packets which are lost.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"p2p_conferences", prometheus.GaugeValue,
		"The number of P2P conferences.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"receive_only_endpoints", prometheus.GaugeValue,
		"The number of endpoints which are sending neither audio nor video and aren't inactive.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"stress_level", prometheus.GaugeValue,
		"The stress level.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_conferences_created", prometheus.CounterValue,
		"The total number of conferences created.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_dominant_speaker_changes", prometheus.CounterValue,
		"The total number of dominant speaker changes.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_ice_failed", prometheus.CounterValue,
		"The total number of failed ICE connections.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_ice_succeeded", prometheus.CounterValue,
		"The total number of succeeded ICE connections.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_ice_succeeded_relayed", prometheus.CounterValue,
		"The total number of succeeded ICE connections which are relayed.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_packets_dropped_octo", prometheus.CounterValue,
		"The total number of packets dropped to or from Octo.", []string{"jvb_instance"}, constLabels))

	collector.metrics = append(collector.metrics, newMetric(collector.NamePrefix+"total_participants", prometheus.CounterValue,
		"The total number of participants.", []string{"jvb_instance"}, constLabels))

	return collector
}

//Describe implements prometheus.Collector interface
func (c *JvbCollector) Describe(desc chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		desc <- m.desc
	}
}

//Collect implements prometheus.Collector interface
func (c *JvbCollector) Collect(metrics chan<- prometheus.Metric) {
	for _, set := range c.statsSets {
		if time.Since(set.lastUpdated) <= c.Retention {

			//match metric names with stats
			for _, stat := range set.stats.Stats {
				for _, metric := range c.metrics {
					if metric.name == c.NamePrefix+stat.Name {

						//special case for histograms
						if stat.Name == "conference_sizes" || stat.Name == "conferences_by_audio_senders" || stat.Name == "conferences_by_video_senders" {
							buckets, sum := bucketsHelper(stat.Value)
							m, err := prometheus.NewConstHistogram(metric.desc, sum, float64(sum), buckets, set.jvbIdentifier)

							if err != nil {
								fmt.Printf("Unable to publish metric %s: %s\n", metric.name, err.Error())
								continue
							}

							metrics <- m
							continue
						}

						//simple metrics
						value, err := strconv.ParseFloat(stat.Value, 64)
						if err != nil {
							fmt.Printf("unable to convert value %s to numeric: %s\n", stat.Value, err.Error())
							continue
						}
						m, err := prometheus.NewConstMetric(metric.desc, metric.metricType, float64(value), set.jvbIdentifier)
						if err != nil {
							fmt.Printf("Unable to create metric %s: %s\n", metric.name, err.Error())
							continue
						}
						metrics <- m
					}
				}
			}
		}
	}
}

//Update updates the cached stats for the JVB identified by identifier, inserts a new stats set if none present yet.
//identifier: any string that identifies the specific JVB, you might want to consider using the node part of the JVB jid (<node>@<domain>/<resource>)
//	instead of the whole jid. This helps to keep track of JVBs being autoscaled
//stats: as they are unmarshalled by the PresExtension
func (c *JvbCollector) Update(identifier string, stats *Stats) {
	for i, s := range c.statsSets {
		if s.jvbIdentifier == identifier {
			c.statsSets[i].lastUpdated = time.Now()
			c.statsSets[i].stats = *stats
			return
		}
	}

	c.statsSets = append(c.statsSets, statsSet{
		lastUpdated:   time.Now(),
		stats:         *stats,
		jvbIdentifier: identifier,
	})
}

func bucketsHelper(value string) (histogram map[float64]uint64, sum uint64) {
	histogram = make(map[float64]uint64)
	value = strings.Trim(value, "[]")
	var values []uint64
	for _, v := range strings.Split(value, ",") {
		vuint, _ := strconv.ParseUint(v, 10, 64)
		values = append(values, vuint)
	}

	//calculate sum (makes this metric independent from conferences metric)
	sum = 0
	for _, v := range values {
		sum += v
	}

	//for the histgram buckets we need to omit the last field b/c the +inf bucket is added automatically
	values = values[:len(values)-1]

	//the bucket values have to be cumulative
	var i int
	for i = len(values) - 1; i >= 0; i-- {
		var cumulative uint64
		var j int
		for j = i; j >= 0; j-- {
			cumulative += values[j]
		}
		values[i] = cumulative
	}

	for i, v := range values {
		histogram[float64(i)] = v
	}

	return
}
