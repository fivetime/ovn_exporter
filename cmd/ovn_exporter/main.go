package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	ovn "github.com/Liquescent-Development/ovn_exporter/pkg/ovn_exporter"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var listenAddress string
	var metricsPath string
	var pollTimeout int
	var pollInterval int
	var isShowVersion bool
	var logLevel string
	var collectorComponents string
	var collectorTopology bool
	var databaseNorthboundName string
	var databaseNorthboundSocketRemote string
	var databaseNorthboundSocketControl string
	var databaseNorthboundFileDataPath string
	var databaseNorthboundFileLogPath string
	var databaseNorthboundFilePidPath string
	var databaseNorthboundPortDefault int
	var databaseNorthboundPortSsl int
	var databaseNorthboundPortRaft int
	var databaseSouthboundName string
	var databaseSouthboundSocketRemote string
	var databaseSouthboundSocketControl string
	var databaseSouthboundFileDataPath string
	var databaseSouthboundFileLogPath string
	var databaseSouthboundFilePidPath string
	var databaseSouthboundPortDefault int
	var databaseSouthboundPortSsl int
	var databaseSouthboundPortRaft int
	var serviceNorthdFileLogPath string
	var serviceNorthdFilePidPath string
	var serviceNorthdSocketControl string

	flag.StringVar(&listenAddress, "web.listen-address", ":9476", "Address to listen on for web interface and telemetry.")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.IntVar(&pollTimeout, "ovn.timeout", 2, "Timeout on gRPC requests to OVN.")
	flag.IntVar(&pollInterval, "ovn.poll-interval", 15, "The minimum interval (in seconds) between collections from OVN server.")
	flag.BoolVar(&isShowVersion, "version", false, "version information")
	flag.StringVar(&logLevel, "log.level", "info", "logging severity level")
	flag.StringVar(&collectorComponents, "collector.components", strings.Join(ovn.DefaultComponents, ","), "Comma separated local components to probe through control sockets, pid and log files. Empty disables local probes (use it when the databases are reached over TCP).")
	flag.BoolVar(&collectorTopology, "collector.topology", true, "Query the NB/SB tables for chassis, logical switch/router and port binding metrics.")

	flag.StringVar(&databaseNorthboundName, "database.northbound.name", "OVN_Northbound", "The name of OVN NB (northbound) db.")
	flag.StringVar(&databaseNorthboundSocketRemote, "database.northbound.socket.remote", "unix:/run/openvswitch/ovnnb_db.sock", "JSON-RPC unix socket to OVN NB db.")
	flag.StringVar(&databaseNorthboundSocketControl, "database.northbound.socket.control", "unix:/run/openvswitch/ovnnb_db.ctl", "JSON-RPC unix socket to OVN NB app.")
	flag.StringVar(&databaseNorthboundFileDataPath, "database.northbound.file.data.path", "/var/lib/openvswitch/ovnnb_db.db", "OVN NB db file.")
	flag.StringVar(&databaseNorthboundFileLogPath, "database.northbound.file.log.path", "/var/log/openvswitch/ovsdb-server-nb.log", "OVN NB db log file.")
	flag.StringVar(&databaseNorthboundFilePidPath, "database.northbound.file.pid.path", "/run/openvswitch/ovnnb_db.pid", "OVN NB db process id file.")
	flag.IntVar(&databaseNorthboundPortDefault, "database.northbound.port.default", 6641, "OVN NB db network socket port.")
	flag.IntVar(&databaseNorthboundPortSsl, "database.northbound.port.ssl", 6631, "OVN NB db network socket secure port.")
	flag.IntVar(&databaseNorthboundPortRaft, "database.northbound.port.raft", 6643, "OVN NB db network port for clustering (raft)")

	flag.StringVar(&databaseSouthboundName, "database.southbound.name", "OVN_Southbound", "The name of OVN SB (southbound) db.")
	flag.StringVar(&databaseSouthboundSocketRemote, "database.southbound.socket.remote", "unix:/run/openvswitch/ovnsb_db.sock", "JSON-RPC unix socket to OVN SB db.")
	flag.StringVar(&databaseSouthboundSocketControl, "database.southbound.socket.control", "unix:/run/openvswitch/ovnsb_db.ctl", "JSON-RPC unix socket to OVN SB app.")
	flag.StringVar(&databaseSouthboundFileDataPath, "database.southbound.file.data.path", "/var/lib/openvswitch/ovnsb_db.db", "OVN SB db file.")
	flag.StringVar(&databaseSouthboundFileLogPath, "database.southbound.file.log.path", "/var/log/openvswitch/ovsdb-server-sb.log", "OVN SB db log file.")
	flag.StringVar(&databaseSouthboundFilePidPath, "database.southbound.file.pid.path", "/run/openvswitch/ovnsb_db.pid", "OVN SB db process id file.")
	flag.IntVar(&databaseSouthboundPortDefault, "database.southbound.port.default", 6642, "OVN SB db network socket port.")
	flag.IntVar(&databaseSouthboundPortSsl, "database.southbound.port.ssl", 6632, "OVN SB db network socket secure port.")
	flag.IntVar(&databaseSouthboundPortRaft, "database.southbound.port.raft", 6644, "OVN SB db network port for clustering (raft)")

	flag.StringVar(&serviceNorthdFileLogPath, "service.ovn.northd.file.log.path", "/var/log/openvswitch/ovn-northd.log", "OVN northd daemon log file.")
	flag.StringVar(&serviceNorthdFilePidPath, "service.ovn.northd.file.pid.path", "/run/openvswitch/ovn-northd.pid", "OVN northd daemon process id file.")
	flag.StringVar(&serviceNorthdSocketControl, "service.ovn.northd.socket.control", "unix:/run/openvswitch/ovn-northd.ctl", "JSON-RPC unix socket to OVN northd app.")

	var usageHelp = func() {
		fmt.Fprintf(os.Stderr, "\n%s - Prometheus Exporter for Open Virtual Network (OVN)\n\n", ovn.GetExporterName())
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", ovn.GetExporterName())
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: https://github.com/Liquescent-Development/ovn_exporter/\n\n")
	}
	flag.Usage = usageHelp
	flag.Parse()

	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s %s", ovn.GetExporterName(), ovn.GetVersion())
		if ovn.GetRevision() != "" {
			fmt.Fprintf(os.Stdout, ", commit: %s\n", ovn.GetRevision())
		} else {
			fmt.Fprint(os.Stdout, "\n")
		}
		os.Exit(0)
	}

	logger, err := ovn.NewLogger(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed initializing logger: %v", err)
		os.Exit(1)
	}

	level.Info(logger).Log(
		"msg", "Starting exporter",
		"exporter", ovn.GetExporterName(),
		"version", ovn.GetVersionInfo(),
		"build_context", ovn.GetVersionBuildContext(),
	)

	opts := ovn.Options{
		Timeout: pollTimeout,
		Logger:  logger,
	}

	exporter, err := ovn.NewExporter(opts)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to init properly",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	components, err := ovn.ParseComponents(collectorComponents)
	if err != nil {
		level.Error(logger).Log(
			"msg", "invalid -collector.components",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	exporter.SetComponents(components)
	exporter.SetCollectTopology(collectorTopology)

	exporter.Client.Database.Northbound.Name = databaseNorthboundName
	exporter.Client.Database.Northbound.Socket.Remote = databaseNorthboundSocketRemote
	exporter.Client.Database.Northbound.Socket.Control = databaseNorthboundSocketControl
	exporter.Client.Database.Northbound.File.Data.Path = databaseNorthboundFileDataPath
	exporter.Client.Database.Northbound.File.Log.Path = databaseNorthboundFileLogPath
	exporter.Client.Database.Northbound.File.Pid.Path = databaseNorthboundFilePidPath
	exporter.Client.Database.Northbound.Port.Default = databaseNorthboundPortDefault
	exporter.Client.Database.Northbound.Port.Ssl = databaseNorthboundPortSsl
	exporter.Client.Database.Northbound.Port.Raft = databaseNorthboundPortRaft

	exporter.Client.Database.Southbound.Name = databaseSouthboundName
	exporter.Client.Database.Southbound.Socket.Remote = databaseSouthboundSocketRemote
	exporter.Client.Database.Southbound.Socket.Control = databaseSouthboundSocketControl
	exporter.Client.Database.Southbound.File.Data.Path = databaseSouthboundFileDataPath
	exporter.Client.Database.Southbound.File.Log.Path = databaseSouthboundFileLogPath
	exporter.Client.Database.Southbound.File.Pid.Path = databaseSouthboundFilePidPath
	exporter.Client.Database.Southbound.Port.Default = databaseSouthboundPortDefault
	exporter.Client.Database.Southbound.Port.Ssl = databaseSouthboundPortSsl
	exporter.Client.Database.Southbound.Port.Raft = databaseSouthboundPortRaft

	exporter.Client.Service.Northd.File.Log.Path = serviceNorthdFileLogPath
	exporter.Client.Service.Northd.File.Pid.Path = serviceNorthdFilePidPath
	exporter.Client.Service.Northd.Socket.Control = serviceNorthdSocketControl

	exporter, err = ovn.ExporterPerformClientCalls(exporter)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to finalize exporter calls properly",
			"exporter_name", ovn.GetExporterName(),
			"error", err.Error(),
		)
	}

	level.Info(logger).Log("ovs_system_id", exporter.Client.System.ID)

	exporter.SetPollInterval(int64(pollInterval))
	prometheus.MustRegister(exporter)

	// Create a new ServeMux instead of using DefaultServeMux for better security
	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte(`<html>
             <head><title>OVN Exporter</title></head>
             <body>
             <h1>OVN Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	// Create server with security configurations
	server := &http.Server{
		Addr:         listenAddress,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	level.Info(logger).Log("listen_on", listenAddress)
	if err := server.ListenAndServe(); err != nil {
		level.Error(logger).Log(
			"msg", "listener failed",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
