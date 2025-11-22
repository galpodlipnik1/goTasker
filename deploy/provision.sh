#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

echo "[*] Updating packages"
apt-get update -y
apt-get upgrade -y

echo "[*] Installing Redis, SQLite, Git, CA certs, Nginx"
apt-get install -y redis-server sqlite3 git ca-certificates curl nginx

echo "[*] Installing Go 1.22"
curl -L https://go.dev/dl/go1.22.0.linux-amd64.tar.gz -o /tmp/go.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf /tmp/go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
source /etc/profile.d/go.sh

echo "[*] Creating app user and directories"
useradd -r -s /usr/sbin/nologin gotasker || true
mkdir -p /opt/gotasker /var/www/gotasker /var/lib/gotasker
chown -R gotasker:gotasker /opt/gotasker /var/lib/gotasker
chown -R www-data:www-data /var/www/gotasker

echo "[*] Building Go application"
cd /vagrant/app
/usr/local/go/bin/go mod tidy
/usr/local/go/bin/go build -o gotasker

cp ./gotasker /opt/gotasker/gotasker
chown gotasker:gotasker /opt/gotasker/gotasker
chmod 755 /opt/gotasker/gotasker

echo "[*] Installing environment file"
cp /vagrant/env/app.env /etc/gotasker.env
chmod 640 /etc/gotasker.env
chown root:root /etc/gotasker.env

echo "[*] Installing static frontend"
/bin/cp -r /vagrant/www/* /var/www/gotasker/

echo "[*] Configuring Nginx"
cp /vagrant/nginx/default /etc/nginx/sites-available/default
systemctl reload nginx

echo "[*] Installing systemd services"
cp /vagrant/systemd/gotasker.service /etc/systemd/system/gotasker.service

echo "[*] Enabling services"
systemctl daemon-reload
systemctl enable redis-server
systemctl enable gotasker
systemctl enable nginx

echo "[*] Starting services"
systemctl restart redis-server
systemctl restart gotasker
systemctl restart nginx

echo "[*] Done."
echo "Access the VM site: http://localhost:8443"
