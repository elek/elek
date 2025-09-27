package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CLI struct {
	Source          string            `short:"s" long:"source" help:"Source IP address for MTR packets (optional)"`
	Targets         []string          `arg:"" name:"targets" help:"IP addresses or hostnames to monitor"`
	Port            int               `short:"p" long:"port" default:"8080" help:"HTTP server port for metrics endpoint"`
	Interval        int               `short:"i" long:"interval" default:"60" help:"Interval in seconds between MTR runs"`
	Labels          map[string]string `short:"l" long:"labels" help:"Additional labels to add"`
	latencyGauge    *prometheus.GaugeVec
	packetLossGauge *prometheus.GaugeVec
}

type MTRReport struct {
	Report struct {
		Mtr struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"mtr"`
		Hubs []struct {
			Count int     `json:"count"`
			Host  string  `json:"host"`
			Loss  float64 `json:"Loss%"`
			Snt   int     `json:"Snt"`
			Last  float64 `json:"Last"`
			Avg   float64 `json:"Avg"`
			Best  float64 `json:"Best"`
			Wrst  float64 `json:"Wrst"`
			StDev float64 `json:"StDev"`
		} `json:"hubs"`
	} `json:"report"`
}

func runMTR(target string) (*MTRReport, error) {
	cmd := exec.Command("mtr", "-z", target, "--json", "-c", "5", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run mtr for %s: %v", target, err)
	}

	var report MTRReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("failed to parse mtr output for %s [%s]: %v", target, string(output), err)
	}

	return &report, nil
}

func (c CLI) updateMetrics() {
	for _, target := range c.Targets {
		label, ip, _ := strings.Cut(target, ":")
		report, err := runMTR(ip)
		if err != nil {
			log.Printf("Error running MTR for %s: %v", target, err)
			continue
		}

		if len(report.Report.Hubs) == 0 {
			log.Printf("No hops found in MTR report for %s", target)
			continue
		}
		last := report.Report.Hubs[len(report.Report.Hubs)-1]
		if last.Host != ip {
			log.Printf("Warning: Last hop host (%s) does not match target (%s)", last.Host, target)
			continue
		}
		labels := []string{c.Source, label}
		for _, v := range c.Labels {
			labels = append(labels, v)
		}
		c.latencyGauge.WithLabelValues(labels...).Set(last.Avg)
		c.packetLossGauge.WithLabelValues(labels...).Set(last.Loss)

	}
}

func (c CLI) startMetricsCollection() {
	ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)
	defer ticker.Stop()

	// Run once immediately
	c.updateMetrics()

	for range ticker.C {
		c.updateMetrics()
	}
}

func (c CLI) Run() error {
	if len(c.Targets) == 0 {
		fmt.Fprintf(os.Stderr, "Error: At least one target must be specified\n")
		os.Exit(1)
	}

	// Validate targets
	for _, target := range c.Targets {
		if strings.TrimSpace(target) == "" {
			fmt.Fprintf(os.Stderr, "Error: Empty target specified\n")
			os.Exit(1)
		}
	}

	log.Printf("Starting smokeping monitoring for targets: %v", c.Targets)
	log.Printf("Metrics will be available at http://localhost:%d/metrics", c.Port)
	log.Printf("Update interval: %d seconds", c.Interval)

	labels := []string{"source", "target"}
	for k := range c.Labels {
		labels = append(labels, k)
	}

	c.latencyGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "smokeping_latency_ms",
			Help: "Average latency to target in milliseconds",
		},
		labels,
	)

	c.packetLossGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "smokeping_packet_loss_percent",
			Help: "Packet loss percentage to target",
		},
		labels,
	)

	prometheus.MustRegister(c.latencyGauge)
	prometheus.MustRegister(c.packetLossGauge)

	// Start metrics collection in background
	go c.startMetricsCollection()

	// Setup HTTP server
	http.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting HTTP server on :%d", c.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
	return nil
}

func main() {
	var cli CLI
	ktx := kong.Parse(&cli)
	err := ktx.Run()
	if err != nil {
		ktx.Fatalf("Error: %++v", err)
	}

}
