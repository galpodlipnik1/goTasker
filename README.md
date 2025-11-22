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

### Option 1: Vagrant

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

![Vagrant Success](/images/{A1D2B840-A555-4A69-B2FE-41C0C2A6EF4F}.png)

### Option 2: Cloud-init

**Steps:**

1.  Use the `deploy/cloud-init.yaml` file as the User Data / Custom Data when creating a new VM instance.
2.  **Important**: Update the `git clone` URL in `cloud-init.yaml` to point to your repository.
3.  Once the instance is running, access the application via the VM's Public IP.

![Multipass Dashboard](/images/{29138FFB-0B59-45A3-ABDA-45E860528BCB}.png)
![App Accessed via VM's IP](/images/{1FDA0633-6C4F-4177-938F-3860DEFEEAB6}.png)

### Application UI

![App UI](/images/{762791B3-BF02-4ADA-9E27-588AF8F68DC0}.png)

### Performance Testing

1.  **Generate Load**: Click the green "Generate 1000 Tasks" button.
2.  **Check Metrics**: Observe the "Time" metric. It should remain low due to Redis caching.
3.  **Cache Hit**: Refresh the page. The "Cache Hit" indicator should show **true**.

![Cache Hit](/images/{DAC93F10-0626-437A-B1C5-48E79BC2C317}.png)

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
├── deploy/             # Deployment Configuration
│   ├── nginx/          # Nginx Config
│   ├── systemd/        # Systemd Service File
│   ├── www/            # Static Frontend (HTML/JS)
│   ├── env/            # Environment Variables
│   ├── provision.sh    # Vagrant Provisioning Script
│   ├── cloud-init.yaml # Cloud-init Configuration
│   └── Vagrantfile     # Vagrant Configuration
├── images/             #Project images
|
└── README.md
```
