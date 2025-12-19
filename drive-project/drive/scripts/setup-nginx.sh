#!/bin/bash

# Nginx Setup Script for GoDrive on GCP VM
# Run this after the main setup-gcp-vm.sh

set -e

echo "🌐 Setting up Nginx reverse proxy..."

# Install Nginx and Certbot
echo "📦 Installing Nginx and Certbot..."
sudo apt-get update
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Stop Nginx temporarily
sudo systemctl stop nginx

# Backup default config
sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup

# Copy our configurations
echo "📝 Copying Nginx configurations..."
sudo cp /opt/godrive/nginx/api.conf /etc/nginx/sites-available/godrive-api
sudo cp /opt/godrive/nginx/grafana.conf /etc/nginx/sites-available/godrive-grafana

# Remove default site
sudo rm -f /etc/nginx/sites-enabled/default

# Enable our sites
sudo ln -sf /etc/nginx/sites-available/godrive-api /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/godrive-grafana /etc/nginx/sites-enabled/

# Test Nginx configuration
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Start Nginx
echo "🚀 Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# Configure firewall to allow Nginx
echo "🔥 Configuring firewall..."
sudo ufw allow 'Nginx Full' || echo "UFW not enabled, skipping..."

echo ""
echo "✅ Nginx setup complete!"
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Update DNS records to point to this server:"
echo "   - api.yourdomain.com → $(curl -s ifconfig.me)"
echo "   - grafana.yourdomain.com → $(curl -s ifconfig.me)"
echo ""
echo "2. Edit Nginx configs with your domain:"
echo "   sudo nano /etc/nginx/sites-available/godrive-api"
echo "   sudo nano /etc/nginx/sites-available/godrive-grafana"
echo ""
echo "3. Set up SSL certificates (after DNS propagation):"
echo "   sudo certbot --nginx -d api.yourdomain.com"
echo "   sudo certbot --nginx -d grafana.yourdomain.com"
echo ""
echo "4. Reload Nginx:"
echo "   sudo systemctl reload nginx"
echo ""
echo "🌐 Current access (HTTP only):"
echo "  - API: http://$(curl -s ifconfig.me)"
echo "  - Grafana: http://$(curl -s ifconfig.me):3001 (direct)"
