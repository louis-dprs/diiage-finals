# Traffic Generator App

Application Go simple qui génère du trafic continu vers les différents endpoints du backend.

## Fonctionnalités

L'application fait des requêtes HTTP GET périodiques vers les endpoints suivants du backend :
- `/` - Endpoint racine
- `/health` - Health check
- `/config` - Lecture du ConfigMap
- `/pods` - Liste des pods
- `/metrics` - Métriques Prometheus

## Configuration

L'application utilise les variables d'environnement suivantes :

- `BACKEND_URL` : URL de base du backend (défaut: `http://backend-prod:80`)
- `INTERVAL` : Intervalle entre les cycles de requêtes (défaut: `5s`)

## Utilisation locale

Pour exécuter localement :

```bash
export BACKEND_URL=http://localhost:8080
export INTERVAL=10s
go run main.go
```

## Build

Pour construire l'image Docker :

```bash
docker build -t traffic-generator -f Dockerfile .
```

## Structure

```
traffic-gen-app/
├── main.go          # Code source de l'application
├── go.mod           # Module Go
├── Dockerfile       # Dockerfile pour builder l'image
└── README.md        # Documentation
```
