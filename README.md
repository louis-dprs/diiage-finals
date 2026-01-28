# DIIAGE – Exercice final

Ce dépôt contient une application (backend + traffic-generator), une chart de monitoring et un script pour installer ArgoCD. Plusieurs éléments ont été volontairement retirés.

**Objectif** : compléter le système et corriger les erreurs pour obtenir un déploiement fonctionnel en mode GitOps.

---

## Prérequis

- Cluster Kubernetes (kind, minikube, ou même [killercoda](https://killercoda.com/playgrounds/course/kubernetes-playgrounds/one-node-4GB) …)
- `kubectl`, `helm`

## Énoncé – 5 étapes (1h réalisation + 30min debrief)

### 1. Installer ArgoCD

Installez ArgoCD dans le cluster à l’aide du script fourni :

```bash
./bootstrap.sh
```

Accédez à l’UI ArgoCD (adaptez selon votre contexte : port-forward, NodePort, Ingress…) et connectez-vous avec :

- **Username** : `admin`  
- **Password** : `admin`

---

### 2. Installer toutes les apps « the GitOps-way »

Déployez **toutes** les applications du dépôt via ArgoCD (GitOps) :

- Le **backend** (`app/chart`)
- Le **traffic-generator** (`traffic-gen-app/chart`)
- La **stack de monitoring** (`monitoring/chart`)

Vous devez définir vous-même la manière dont ArgoCD pointera vers ce repo et vers ces charts Helm, puis laisser ArgoCD synchroniser.

**Critère de succès** : à la fin de l’énoncé, vous devriez pouvoir supprimer le cluster et le recréer sans intervention manuelle (à part le bootstrap) ; le setup GitOps suffit à tout redéployer.

---

### 3. Ajouter les pièces manquantes du monitoring

Le pipeline d'ingestion est incomplet : certains composants ou ressources nécessaires à la télémétrie ont été retirés.

Identifiez ce qui manque, ajoutez les manifests ou templates Helm adéquats, et assurez-vous que la stack de monitoring fonctionne correctement (Prometheus, Grafana, Tempo, collecte OTLP, etc.). Le Grafana est fourni avec des dashboards custom, à la fin vous devriez avoir des dashboard fonctionnel

![grafana dashboard](./images/grafana-dashboard.png)

---

### 4. Corriger les erreurs dans l’application

L’application (backend et/ou traffic-generator) peut présenter des erreurs au runtime (permissions, configuration, dépendances…).

Repérez les problèmes (logs, événements Kubernetes, endpoints qui échouent), corrigez la configuration ou le code si besoin, et validez que les apps tournent correctement.

---

### 5. S’assurer que les 2 apps peuvent uniquement s’appeler entre elles

Les deux applications (backend et traffic-generator) doivent pouvoir communiquer **entre elles**, mais **pas** avec le reste du cluster de manière non maîtrisée.

Mettez en place toute mécanique qui vous semble appropriée pour que :

- le traffic-generator puisse appeler le backend ;
- le backend et le traffic-generator ne soient pas accessibles ou n’appellent pas d’autres services de façon non désirée.

---

## Structure du dépôt

| Dossier / fichier | Description |
|-------------------|-------------|
| `app/` | Backend Go + chart Helm |
| `traffic-gen-app/` | Traffic-generator Go + chart Helm |
| `monitoring/chart/` | Chart Helm monitoring (cert-manager, Prometheus, Grafana, Tempo, OpenTelemetry Operator) |
| `bootstrap.sh` | Installation d’ArgoCD via Helm |

---

## Les 2 applications

### Backend (`app/`)

Service HTTP Go (port 8080) qui expose :

- **`/`** : renvoie le namespace du pod
- **`/health`** : healthcheck
- **`/config`** : lit un ConfigMap Kubernetes (nom configurable) via l'API Kubernetes et renvoie son contenu en JSON
- **`/pods`** : liste tous les pods du cluster via l'API Kubernetes
- **`/metrics`** : métriques Prometheus (`http_requests_total`, `configmap_read_total`)

Il utilise le *in-cluster* Kubernetes config (client go kubernetes) pour accéder aux ConfigMaps (namespace) et aux Pods (cluster).

**Configuration** (Helm `app/chart/values.yaml`, env dans le Déploiement) :

- `POD_NAMESPACE` : injecté depuis le pod
- `CONFIGMAP_NAME` : nom du ConfigMap lu par `/config` (défaut `app-config`)
- Le chart peut créer un ConfigMap (`backend.config.createConfigMap`) et remplir `configmapData` ; l'app le lit via `/config`.

Le Service est en ClusterIP (port 80 → 8080). Le nom du Service dépend du release Helm.

---

### Traffic-generator (`traffic-gen-app/`)

Binaire Go qui tourne en boucle : à intervalle régulier, il envoie des requêtes HTTP GET vers le backend sur `/`, `/health`, `/config`, `/pods` et `/metrics`. Il sert à générer du trafic pour le monitoring et à vérifier que le backend répond.

**Configuration** (Helm `traffic-gen-app/chart/values.yaml`, variables d'environnement) :

- **`BACKEND_URL`** : URL de base du backend (ex. `http://<service-backend>:80`). Doit pointer vers le Service Kubernetes du backend ; le nom du service dépend du release et du namespace.
- **`INTERVAL`** : période entre deux cycles de requêtes (défaut `5s`).

Pas de Service : le traffic-gen sort vers le backend, il n'expose rien.

