#!/bin/bash

# Complete GoDrive Setup Script for GCP VM
# This sets up both backend and frontend on a single VM

set -e

echo "🚀 Setting up complete GoDrive stack on GCP VM..."

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

sudo usermod -aG docker $USER

# Install Docker Compose
echo "📦 Installing Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Node.js and npm
echo "📦 Installing Node.js..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install PM2 globally
echo "📦 Installing PM2..."
sudo npm install -g pm2

# Install Nginx and Certbot
echo "🌐 Installing Nginx and Certbot..."
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Install Git
echo "📥 Installing Git..."
sudo apt-get install -y git

# Install golang-migrate
echo "🗄️ Installing golang-migrate..."
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
sudo chmod +x /usr/local/bin/migrate

# Create application directories
echo "📁 Creating application directories..."
sudo mkdir -p /opt/godrive
sudo mkdir -p /opt/godrive-ui
sudo chown $USER:$USER /opt/godrive
sudo chown $USER:$USER /opt/godrive-ui

echo ""
echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Clone your repositories:"
echo "   cd /opt/godrive && git clone <backend-repo-url> ."
echo "   cd /opt/godrive-ui && git clone <frontend-repo-url> ."
echo ""
echo "2. Configure backend:"
echo "   cd /opt/godrive"
echo "   cp .env.template .env"
echo "   nano .env  # Set passwords and secrets"
echo "   ./scripts/deploy.sh"
echo ""
echo "3. Configure frontend:"
echo "   cd /opt/godrive-ui"
echo "   nano .env.local  # Set NEXT_PUBLIC_API_BASE_URL"
echo "   chmod +x scripts/deploy-frontend.sh"
echo "   ./scripts/deploy-frontend.sh"
echo ""
echo "4. Set up SSL certificates:"
echo "   sudo certbot --nginx -d api.yourdomain.com -d app.yourdomain.com -d grafana.yourdomain.com"
echo ""
echo "⚠️  You may need to log out and back in for Docker group changes to take effect"
