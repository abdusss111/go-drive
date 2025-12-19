#!/bin/bash

# GoDrive GCP VM Setup Script
# This script sets up a fresh Ubuntu VM on Google Cloud Platform

set -e

echo "🚀 Starting GoDrive deployment on GCP VM..."

# Update system
echo "📦 Updating system packages..."
sudo apt-get update
sudo apt-get upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add current user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose standalone (for compatibility)
echo "📦 Installing Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Git
echo "📥 Installing Git..."
sudo apt-get install -y git

# Install golang-migrate for database migrations
echo "🗄️ Installing golang-migrate..."
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
sudo chmod +x /usr/local/bin/migrate

# Install Nginx and Certbot
echo "🌐 Installing Nginx and Certbot..."
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Create application directory
echo "📁 Creating application directory..."
sudo mkdir -p /opt/godrive
sudo chown $USER:$USER /opt/godrive
cd /opt/godrive

# Clone repository (you'll need to update this with your repo URL)
echo "📥 Cloning repository..."
echo "⚠️  Please run: git clone <your-repo-url> ."
echo "⚠️  Then run: ./deploy.sh"

# Create .env file template
cat > .env.template << 'EOF'
# Server Configuration
GODRIVE_API_HOST=0.0.0.0
GODRIVE_API_PORT=8080

# Database Configuration
POSTGRES_USER=godrive_app
POSTGRES_PASSWORD=CHANGE_ME_STRONG_PASSWORD
POSTGRES_DB=godrive
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_SSL_MODE=disable

# MinIO Configuration
MINIO_ROOT_USER=godrive
MINIO_ROOT_PASSWORD=CHANGE_ME_STRONG_PASSWORD
MINIO_ENDPOINT=minio:9000
MINIO_BUCKET=godrive
MINIO_USE_SSL=false
MINIO_REGION=us-east-1

# Authentication
GODRIVE_JWT_SECRET=CHANGE_ME_32_CHAR_SECRET_KEY_HERE
GODRIVE_JWT_REFRESH_SECRET=CHANGE_ME_64_CHAR_REFRESH_SECRET_KEY_HERE

# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=CHANGE_ME_ADMIN_PASSWORD

# Application Settings
GODRIVE_LOG_LEVEL=info
ENVIRONMENT=production
GODRIVE_METRICS_PATH=/metrics
EOF

echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "1. Clone your repository: git clone <your-repo-url> /opt/godrive"
echo "2. Copy .env.template to .env and update passwords"
echo "3. Run: ./deploy.sh"
echo ""
echo "⚠️  You may need to log out and back in for Docker group changes to take effect"
