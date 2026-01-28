# Monitoring Stack

Chart Helm : Prometheus, Grafana, Tempo, OpenTelemetry Operator, **cert-manager** (dépendance).

Aucun prérequis : cert-manager est installé avec la chart.

## Installation

```bash
cd chart
helm dependency update
helm install monitoring-stack . --namespace monitoring --create-namespace
```

## Prometheus

```bash
kubectl port-forward svc/monitoring-stack-prometheus-server 9090:80
```

## Grafana

```bash
kubectl port-forward -n monitoring svc/monitoring-stack-grafana 3000:80
```

→ http://localhost:3000 — `admin` / `admin`

La chart vient avec 2 dashboards custom.

## Configuration

Tout est dans `chart/values.yaml` (cert-manager, Prometheus, Grafana, Tempo, OpenTelemetry Operator).
