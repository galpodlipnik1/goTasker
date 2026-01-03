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

This section describes how to deploy the application stack to a Kubernetes cluster.

### Prerequisites

- Kubernetes cluster (e.g., Minikube, Kind, GKE, EKS, AKS)
- `kubectl` configured
- Nginx Ingress Controller installed
- cert-manager installed (for TLS)

### Architecture on Kubernetes

- **Frontend**: Nginx serving static files, scaled to 3 replicas for HA.
- **Backend**: GoTasker API (1 replica, using SQLite).
- **Redis**: Redis instance (1 replica).
- **Ingress**: Exposes the application via `localhost` (HTTP for local dev).

### Deployment Steps

1.  **Create Namespace**:
    ```bash
    kubectl apply -f k8s/00-namespace.yaml
    ```

2.  **Apply Manifests**:
    ```bash
    kubectl apply -f k8s/
    ```
    *Note: This will apply all manifests in the `k8s` directory.*

3.  **Verify Deployment**:
    ```bash
    kubectl get pods -n gotasker
    kubectl get ingress -n gotasker
    ```

4.  **Access Application**:
    Open `http://localhost` in your browser.

### CI/CD

A GitHub Actions workflow is provided in `.github/workflows/ci.yaml`. It automatically builds and pushes the Docker images to GitHub Container Registry (GHCR) on push to `main`.

### High Availability & Rolling Updates

- **Frontend**: Configured with 3 replicas.
- **Rolling Update**:
    - Strategy: `maxUnavailable: 0`, `maxSurge: 1`.
    - This ensures zero downtime during updates.
    - To demo: Update the image tag in `k8s/04-frontend.yaml` and apply. Watch pods with `kubectl get pods -n gotasker -w`.

### Blue/Green Deployment

### Blue/Green Deployment Demo

Manifests for Blue/Green deployment are in `k8s/blue-green/`.

1.  **Setup**:
    Apply the ConfigMap, Deployments, Service, and Ingress:
    ```bash
    kubectl apply -f k8s/blue-green/
    ```
    This deploys both Blue (standard) and Green (modified UI) versions, but the Service points to **Blue**.

2.  **Access**:
    Open `http://bluegreen.localhost` in your browser. You should see the standard UI.

3.  **Start Load Test**:
    Open a PowerShell terminal and run the provided script to monitor availability:
    ```powershell
    ./test-downtime.ps1
    ```
    You should see a stream of `Status: 200`.

4.  **Perform Switch (Blue -> Green)**:
    Edit `k8s/blue-green/02-service.yaml` and change the selector to `version: green`.
    Apply the change:
    ```bash
    kubectl apply -f k8s/blue-green/02-service.yaml
    ```

5.  **Verify**:
    - Check the PowerShell script: You should see **NO** failed requests (Status 200 continues uninterrupted).
    - Refresh your browser: You should now see the **Green Version** (Light green background).

### Probes

- **Liveness**: Checks if the container is running.
    - Backend: `/healthz`
    - Frontend: `/`
- **Readiness**: Checks if the application is ready to serve traffic (DB/Redis connected).
    - Backend: `/readyz`
    - Frontend: `/`

### Running with Kind (Kubernetes in Docker)

If you want to run this locally using `kind`, follow these steps:

1.  **Create Cluster**:
    Use the provided config to map ports 80 and 443 to your host.
    ```bash
    kind create cluster --config deploy/kind-config.yaml --name gotasker
    ```

2.  **Install Nginx Ingress Controller**:
    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
    ```
    Wait for the ingress controller to be ready:
    ```bash
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=90s
    ```

3.  **Install Cert Manager** (Optional, for self-signed certs):
    ```bash
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
    ```

4.  **Build and Load Images** (If not using GHCR):
    ```bash
    # Build images
    docker build -t ghcr.io/galpodlipnik1/gotasker:latest -f deploy/Dockerfile .
    docker build -t ghcr.io/galpodlipnik1/gotasker-frontend:latest -f deploy/nginx/Dockerfile .

    # Load into Kind
    kind load docker-image ghcr.io/galpodlipnik1/gotasker:latest --name gotasker
    kind load docker-image ghcr.io/galpodlipnik1/gotasker-frontend:latest --name gotasker
    ```

5.  **Deploy Application**:
    Follow the [Deployment Steps](#deployment-steps) above.

6.  **Access**:
    Add `127.0.0.1 devops.radovan.si` to your hosts file.
    Access at `http://localhost
## Extra Credit: Helm Chart Deployment

I have also created a Helm chart to simplify the deployment and management of the application, including the Blue/Green demo.

### Prerequisites
- Helm installed (`choco install kubernetes-helm` on Windows)

### Deploying with Helm

1.  **Install the Chart**:
    ```bash
    helm install gotasker ./helm/gotasker
    ```
    This deploys the standard application stack (Backend, Frontend, Redis, Ingress).

2.  **Access**:
    Open `http://localhost`.

### Blue/Green Demo with Helm

1.  **Enable Blue/Green Mode**:
    Upgrade the release with the `blueGreen.enabled` flag:
    ```bash
    helm upgrade gotasker ./helm/gotasker --set blueGreen.enabled=true
    ```

2.  **Access**:
    Open `http://bluegreen.localhost`. You should see the **Blue** version.

3.  **Switch to Green**:
    Run the upgrade command changing the `activeVersion`:
    ```bash
    helm upgrade gotasker ./helm/gotasker --set blueGreen.enabled=true --set blueGreen.activeVersion=green
    ```
    Refresh your browser to see the **Green** version.

4.  **Uninstall**:
    ```bash
    helm uninstall gotasker
    ```

## VPS Deployment (CI/CD)

The project is configured to automatically deploy to a VPS on every push to the `master` branch.

### Prerequisites on VPS
1.  **K3s Installed**: `curl -sfL https://get.k3s.io | sh -`
2.  **Cert-Manager Installed**: For automatic TLS certificates.
3.  **DNS Configured**: `devops-kube.galpodlipnik.com` pointing to VPS IP.

### GitHub Secrets Setup
To enable the automatic deployment, you must add the following Secret to your GitHub Repository:

*   **Name**: `KUBECONFIG`
*   **Value**: The content of `/etc/rancher/k3s/k3s.yaml` from your VPS.
    *   *Important*: Replace `127.0.0.1` or `localhost` in the file with your VPS Public IP address.

### How it works
1.  **Build**: GitHub Actions builds the Docker images and tags them with the commit SHA (e.g., `sha-a1b2c3d`).
2.  **Deploy**: It connects to your VPS using the `KUBECONFIG` secret.
3.  **Helm Upgrade**: It runs `helm upgrade --install` setting:
    *   Domain: `devops-kube.galpodlipnik.com`
    *   TLS: `true`
    *   Ingress Class: `traefik` (Default for K3s)
    *   Image Tag: `sha-a1b2c3d` (Ensuring the exact version you just built is deployed).
