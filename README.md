# GoTasker - VM Provisioning Project

## Project Description
GoTasker is a robust, automated VM provisioning project designed to deploy a high-performance Task Management application stack. The goal is to demonstrate Infrastructure as Code (IaC) principles using **Vagrant** for local development and **cloud-init** for cloud deployments.

The application stack consists of 4 main components running on Linux (Ubuntu 22.04):
1.  **HTTP Server**: Nginx (Reverse Proxy & Static File Serving)
2.  **Application**: Go (Golang) API using Gin Framework
3.  **Database**: SQLite (Embedded SQL Database)
4.  **Cache**: Redis (In-memory Data Structure Store)

## Architecture
-   **Nginx**: Listens on port 80, serves the static HTML/JS frontend, and proxies `/api` requests to the Go backend. Adds `X-Response-Time` headers for observability.
-   **Go API**: Runs on port 8081. Handles business logic, CRUD operations, and bulk task generation.
-   **Redis**: Caches task lists to reduce database load and improve response times.
-   **SQLite**: Persists task data.

## Features
-   **CRUD Operations**: Create, Read, Delete tasks.
-   **Performance Metrics**: Real-time display of "Cache Hit" status and Request Latency.
-   **Load Generation**: Generate 1000+ tasks instantly to test performance.
-   **Bulk Operations**: Clear the entire database with one click.

## Deployment Instructions

### Option 1: Local Deployment (Vagrant)
Ideal for local development and testing.

**Prerequisites:**
-   Vagrant installed
-   VirtualBox (or another provider) installed

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

> **[Screenshot Placeholder: Vagrant `vagrant up` success output]**
> *Add a screenshot here showing the terminal output after a successful `vagrant up`.*

### Option 2: Cloud Deployment (Cloud-init)
Ideal for deploying to AWS, Azure, Google Cloud, or Proxmox.

**Steps:**
1.  Use the `deploy/cloud-init.yaml` file as the User Data / Custom Data when creating a new VM instance.
2.  **Important**: Update the `git clone` URL in `cloud-init.yaml` to point to your repository.
3.  Once the instance is running, access the application via the VM's Public IP.

> **[Screenshot Placeholder: Cloud Provider Instance Details]**
> *Add a screenshot here showing the running instance in your cloud provider's dashboard.*

## Usage & Verification

### Application UI
The frontend provides a simple interface to manage tasks.

> **[Screenshot Placeholder: Main Application UI]**
> *Add a screenshot of the GoTasker UI showing the task list and buttons.*

### Performance Testing
1.  **Generate Load**: Click the green "Generate 1000 Tasks" button.
2.  **Check Metrics**: Observe the "Time" metric. It should remain low due to Redis caching.
3.  **Cache Hit**: Refresh the page. The "Cache Hit" indicator should show **true**.

> **[Screenshot Placeholder: Metrics Display]**
> *Add a screenshot zooming in on the "Cache Hit: true | Time: 2ms" metrics.*

## Security
-   **User Isolation**: The application runs under a dedicated `gotasker` user, not root.
-   **Reverse Proxy**: Nginx sits in front of the application, providing an additional layer of security and control.
-   *(Note: For production, HTTPS certificates should be configured via Let's Encrypt/Certbot on Nginx)*.

## Project Structure
```
.
├── app/                # Go Application Source Code
│   ├── main.go         # API Logic
│   └── go.mod          # Dependencies
├── deploy/             # Deployment Configuration
│   ├── nginx/          # Nginx Config
│   ├── systemd/        # Systemd Service File
│   ├── www/            # Static Frontend (HTML/JS)
│   ├── env/            # Environment Variables
│   ├── provision.sh    # Vagrant Provisioning Script
│   ├── cloud-init.yaml # Cloud-init Configuration
│   └── Vagrantfile     # Vagrant Configuration
└── docs/               # Documentation
    └── README.md       # This file
```
