# GoTasker

## Project Description

GoTasker is a Golang based Task Management application stack. The application stack consists of 4 main components:

1.  **HTTP Server**: Nginx (Reverse Proxy & Static File Serving)
2.  **Application**: Go (Golang) API using Gin Framework
3.  **Database**: SQLite (Embedded SQL Database)
4.  **Cache**: Redis (In-memory Data Structure Store)

## Architecture

- **Nginx**: Listens on port 80, serves the static HTML/JS frontend, and proxies `/api` requests to the Go backend.
- **Go API**: Runs on port 8081. Handles business logic, CRUD operations, and bulk task generation.
- **Redis**: Caches task lists to reduce database load and improve response times.
- **SQLite**: Persists task data.

## Features

- **CRUD Operations**: Create, Read, Update, Delete tasks.
- **Performance Metrics**: Real-time display of "Cache Hit" status and Request Latency.
- **Load Generation**: Generate 1000+ tasks instantly to test performance.
- **Bulk Operations**: Clear the entire database with one click.

## Deployment Instructions

### Part 1: Virtual Machine Deployment

#### Option 1: Vagrant

**Prerequisites:**

- Vagrant installed
- VirtualBox (or another provider) installed

**Steps:**

1.  Clone the repository.
2.  Navigate to the `deploy` directory:
    ```bash
    cd deploy
    ```
3.  Start the VM:
    ```bash
    vagrant up
    ```
4.  Access the application at: [http://localhost:8443](http://localhost:8443)

![Vagrant Success](/images/vagrant_success.png)

#### Option 2: Cloud-init

**Steps:**

1.  Use the `deploy/cloud-init.yaml` file as the User Data / Custom Data when creating a new VM instance.
2.  **Important**: Update the `git clone` URL in `cloud-init.yaml` to point to your repository.
3.  Once the instance is running, access the application via the VM's Public IP.

![Multipass Dashboard](/images/multipass_dashboard.png)
![App Accessed via VM's IP](/images/access_via_ip.png)

### Part 2: Containerized Deployment

#### 1. Manual Docker Build

The application uses a **Multi-Stage Dockerfile** (`deploy/Dockerfile`) to ensure a small and secure final image.

-   **Builder Stage**: Uses `golang:1.22-alpine`. Installs build dependencies (`gcc`, `musl-dev`), downloads Go modules, and compiles the application with CGO enabled.
-   **Runtime Stage**: Uses `alpine:latest`. Installs runtime dependencies (`curl`, `ca-certificates`), copies the compiled binary from the builder stage, and sets the entrypoint.

**To build the image manually:**

```bash
# Run from the root of the repository
docker build -t gotasker:local -f deploy/Dockerfile .
```

#### 2. Docker Compose

The `deploy/docker-compose.yaml` file orchestrates the entire stack. It pulls the pre-built image from the GitHub Container Registry (GHCR) by default.

**Services:**
-   **Traefik**: Acts as the edge router/reverse proxy. It handles incoming HTTP/HTTPS requests, manages Let's Encrypt SSL certificates automatically, and routes traffic to the backend services.
-   **App**: The Go backend service. It connects to Redis and SQLite.
    -   *Environment Variables*: Configured to connect to Redis and define the SQLite path.
    -   *Volumes*: Persists the SQLite database in a Docker volume.
-   **Redis**: Provides caching. Persists data to a volume.
-   **Nginx**: Serves the static frontend files (`deploy/www`) and proxies API requests to the Go app.
-   **Prometheus**: Collects metrics from the Go application and itself.

**Configuration Files:**

The `deploy` directory contains specific configurations for the services:

-   **Nginx (`deploy/nginx/`)**:
    -   `default`: Configuration for VM/Vagrant deployment. Proxies to `127.0.0.1:8081`.
    -   `docker-default`: Configuration for Docker deployment. Proxies to the `app` service container.
-   **Prometheus (`deploy/prometheus/prometheus.yml`)**:
    -   Configured to scrape metrics from the `gotasker` app service on port 8081 and its own metrics on port 9090.
-   **Docker Compose (`deploy/docker-compose.yaml`)**:
    -   This is the production-ready configuration including Traefik for SSL termination and load balancing.
    -   *Note*: For development, you might want to override the Traefik labels or use a simpler compose file (e.g., `docker-compose.override.yaml`) to expose ports directly.

**Commands:**

```bash
cd deploy

# Start all services in detached mode
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

#### 3. CI/CD Pipeline (GitHub Actions)

The project utilizes GitHub Actions (`.github/workflows/docker-build.yml`) for Continuous Integration and Deployment.

**Workflow Triggers:**
-   **Push to `master`**: Triggers a build, push to GHCR, and deployment to the VPS.
-   **Tags (`v*`)**: Triggers a build, push, and creates a GitHub Release.
-   **Pull Requests**: Triggers a build check (no push).

**Workflow Steps:**
1.  **Checkout**: Clones the repository.
2.  **Setup Buildx**: Initializes Docker Buildx for multi-platform builds.
3.  **Metadata**: Extracts tags and labels based on the git ref (branch or tag).
4.  **Login**: Authenticates with GitHub Container Registry (GHCR).
5.  **Build & Push**:
    -   Builds the image for multiple platforms (`linux/amd64`, `linux/arm64`).
    -   Pushes the image to `ghcr.io/<owner>/gotasker`.
    -   Uses GitHub Actions cache to speed up builds.
6.  **Deploy to VPS**:
    -   Connects to the production server via SSH.
    -   Pulls the latest code.
    -   Logs into GHCR.
    -   Updates the running services with `docker compose up -d --pull=always`.
7.  **Release**: Creates a GitHub Release with auto-generated notes (only for tags).

### Application UI

![App UI](/images/app_ui.png)

### Performance Testing

1.  **Generate Load**: Click the green "Generate 1000 Tasks" button.
2.  **Check Metrics**: Observe the "Time" metric. It should remain low due to Redis caching.
3.  **Cache Hit**: Refresh the page. The "Cache Hit" indicator should show **true**.

![Cache Hit](/images/cache_hit.png)

## Security

- **User Isolation**: The application runs under a dedicated `gotasker` user, not root.
- **Reverse Proxy**: Nginx sits in front of the application, providing an additional layer of security and control.
- _(Note: For production, HTTPS certificates should be configured via Let's Encrypt/Certbot on Nginx)_.

## Project Structure

```
.
├── app/                # Go Application Source Code
│   ├── main.go         # API Logic
│   └── go.mod          # Dependencies
├── .github/            # GitHub Actions Workflows
├── deploy/             # Deployment Configuration
│   ├── nginx/          # Nginx Config
│   ├── systemd/        # Systemd Service File
│   ├── www/            # Static Frontend (HTML/JS)
│   ├── env/            # Environment Variables
│   ├── docker-compose.yaml # Docker Compose Config
│   ├── Dockerfile      # Docker Build File
│   ├── provision.sh    # Vagrant Provisioning Script
│   ├── cloud-init.yaml # Cloud-init Configuration
│   └── Vagrantfile     # Vagrant Configuration
├── images/             #Project images
|
└── README.md
```

## Part 3: Kubernetes Deployment

This section describes how to deploy the application stack to a Kubernetes cluster. The project supports both manual deployment via raw manifests and automated deployment via **Helm Charts**.

### Prerequisites

- Kubernetes cluster (e.g., Kind, K3s, Minikube)
- `kubectl` configured
- `helm` installed
- Ingress Controller (Nginx for Kind, Traefik for K3s)
- cert-manager (for TLS)

### Architecture

- **Frontend**: Nginx serving static files (3 replicas for HA).
- **Backend**: GoTasker API (1 replica, using SQLite).
- **Redis**: Redis instance (1 replica).
- **Ingress**: Exposes the application via HTTP/HTTPS.

---

### Option 1: Deployment with Helm (Recommended)

I have created a custom Helm chart in `./helm/gotasker` to simplify deployment and management.

#### 1. Local Development (Kind)

**Setup Cluster:**
```bash
kind create cluster --config deploy/kind-config.yaml --name gotasker
# Install Nginx Ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

**Install App:**
```bash
helm install gotasker ./helm/gotasker
```
Access the app at **[http://localhost](http://localhost)**.

![Helm Deployment](/images/deploy-kube.png)

#### 2. Blue/Green Deployment Demo

<video src="images/greenblue.mp4" controls="controls" style="max-width: 730px;">
</video>

The Helm chart includes built-in support for Blue/Green deployments.

**Enable Blue/Green Mode:**
This deploys a second "Green" version alongside the "Blue" version.
```bash
helm upgrade gotasker ./helm/gotasker --set blueGreen.enabled=true
```
Access the Blue version at **[http://bluegreen.localhost](http://bluegreen.localhost)**.

**Switch Traffic to Green:**
Perform a zero-downtime switch to the Green version.
```bash
helm upgrade gotasker ./helm/gotasker --set blueGreen.enabled=true --set blueGreen.activeVersion=green
```
Refresh your browser to see the Green version.

**Verify Zero Downtime:**
Use the provided PowerShell script to monitor availability during the switch:
```powershell
./test-downtime.ps1
```

---

### Option 2: Manual Deployment (Raw Manifests)

If you prefer using raw Kubernetes manifests, they are located in the `k8s/` directory.

1.  **Apply Manifests**:
    ```bash
    kubectl apply -f k8s/00-namespace.yaml
    kubectl apply -f k8s/
    ```
2.  **Blue/Green Manual Switch**:
    *   Deploy: `kubectl apply -f k8s/blue-green/`
    *   Switch: Edit `k8s/blue-green/02-service.yaml` and change selector to `version: green`, then apply.

---

### Production Deployment (VPS & CI/CD)

The project is configured with a **GitHub Actions** pipeline that automatically deploys changes to a production VPS.

#### Infrastructure
- **Server**: VPS running **K3s** (Lightweight Kubernetes).
- **Domain**: `devops-kube.galpodlipnik.com`
- **Security**: Automatic TLS certificates via **cert-manager** and Let's Encrypt.

#### CI/CD Workflow (`.github/workflows/ci.yaml`)
1.  **Build**: On every push to `master`, Docker images are built and tagged with the commit SHA.
2.  **Push**: Images are pushed to GitHub Container Registry (GHCR).
3.  **Deploy**: The workflow connects to the VPS and performs a **Helm Upgrade**:
    *   Updates the image tag to the specific commit SHA.
    *   Ensures zero-downtime rolling updates.

![Build and Push](/images/build-and-push-kube.png)

#### Setup Instructions for VPS
To replicate this setup:
1.  Install K3s: `curl -sfL https://get.k3s.io | sh -`
2.  Add `KUBECONFIG` secret to GitHub (content of `/etc/rancher/k3s/k3s.yaml` with public IP).
3.  Configure DNS A record for your domain.

![VPS Deployment](/images/deploy-to-vps-kube.png)
