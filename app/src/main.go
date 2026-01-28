package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "status"},
	)

	configmapReadTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "configmap_read_total",
			Help: "Total number of configmap reads",
		},
	)
)

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting backend application")

	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("Failed to get in-cluster config", "error", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Failed to create Kubernetes client", "error", err)
		os.Exit(1)
	}

	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
		slog.Warn("POD_NAMESPACE not set, using default namespace")
	}

	slog.Info("Kubernetes client initialized", "namespace", namespace)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Handling request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		httpRequestsTotal.WithLabelValues("/", "200").Inc()
		w.Write([]byte(fmt.Sprintf("Backend running in namespace: %s\n", namespace)))
		slog.Debug("Request completed", "path", r.URL.Path, "status", 200)
	})

	configmapName := os.Getenv("CONFIGMAP_NAME")
	if configmapName == "" {
		configmapName = "app-config"
		slog.Warn("CONFIGMAP_NAME not set, using default", "configmap", configmapName)
	}

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Handling config request", "method", r.Method, "path", r.URL.Path)
		ctx := context.Background()

		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configmapName, metav1.GetOptions{})
		if err != nil {
			slog.Error("Failed to read ConfigMap", "namespace", namespace, "configmap", configmapName, "error", err)
			httpRequestsTotal.WithLabelValues("/config", "500").Inc()
			http.Error(w, fmt.Sprintf("Failed to read ConfigMap: %v", err), http.StatusInternalServerError)
			return
		}

		configmapReadTotal.Inc()
		httpRequestsTotal.WithLabelValues("/config", "200").Inc()
		slog.Info("ConfigMap read successfully", "namespace", namespace, "configmap", configmapName, "keys", len(cm.Data))

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cm.Data); err != nil {
			slog.Error("Failed to encode ConfigMap data", "error", err)
		}
	})

	http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Handling pods request", "method", r.Method, "path", r.URL.Path)
		ctx := context.Background()

		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			slog.Error("Failed to list pods", "error", err)
			httpRequestsTotal.WithLabelValues("/pods", "500").Inc()
			http.Error(w, fmt.Sprintf("Failed to list pods: %v", err), http.StatusInternalServerError)
			return
		}

		httpRequestsTotal.WithLabelValues("/pods", "200").Inc()
		slog.Info("Pods listed successfully", "count", len(pods.Items))

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pods.Items); err != nil {
			slog.Error("Failed to encode pods list", "error", err)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Health check request", "path", r.URL.Path)
		httpRequestsTotal.WithLabelValues("/health", "200").Inc()
		w.Write([]byte("OK"))
	})

	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		slog.Info("PORT not set, using default port", "port", port)
	}

	slog.Info("Starting HTTP server", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
}
