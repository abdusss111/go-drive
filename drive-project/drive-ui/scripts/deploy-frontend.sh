#!/bin/bash

# Frontend Build and Deploy Script for GCP VM
# This script builds the Next.js app and deploys it with Nginx

set -e

echo "🎨 Building and deploying GoDrive Frontend..."

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: package.json not found!"
    echo "Please run this script from the drive-ui directory"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Build the Next.js application
echo "🏗️  Building Next.js application..."
npm run build

# Create deployment directory
echo "📁 Creating deployment directory..."
sudo mkdir -p /var/www/godrive-ui
sudo chown -R $USER:$USER /var/www/godrive-ui

# Copy built files
echo "📋 Copying built files..."
rm -rf /var/www/godrive-ui/*
cp -r .next /var/www/godrive-ui/
cp -r public /var/www/godrive-ui/
cp -r out /var/www/godrive-ui/ 2>/dev/null || echo "No 'out' directory (using server mode)"

# Copy package files for server mode
cp package.json /var/www/godrive-ui/
cp package-lock.json /var/www/godrive-ui/ 2>/dev/null || true

# Install production dependencies
echo "📦 Installing production dependencies..."
cd /var/www/godrive-ui
npm install --production

# Setup Nginx configuration
echo "🌐 Setting up Nginx configuration..."
sudo cp /opt/godrive-ui/nginx/app.conf /etc/nginx/sites-available/godrive-ui
sudo ln -sf /etc/nginx/sites-available/godrive-ui /etc/nginx/sites-enabled/

# Test Nginx configuration
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Reload Nginx
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx

# Setup PM2 for Next.js server (if using server mode)
if command -v pm2 &> /dev/null; then
    echo "🚀 Starting Next.js server with PM2..."
    cd /var/www/godrive-ui
    pm2 delete godrive-ui 2>/dev/null || true
    pm2 start npm --name "godrive-ui" -- start
    pm2 save
else
    echo "⚠️  PM2 not installed. Install with: npm install -g pm2"
    echo "   For production, run: pm2 start npm --name godrive-ui -- start"
fi

echo ""
echo "✅ Frontend deployment complete!"
echo ""
echo "🌐 Access your app:"
echo "  - Frontend: http://$(curl -s ifconfig.me)"
echo ""
echo "📝 Next steps:"
echo "1. Update domain in Nginx config:"
echo "   sudo nano /etc/nginx/sites-available/godrive-ui"
echo ""
echo "2. Set up SSL:"
echo "   sudo certbot --nginx -d app.yourdomain.com"
echo ""
echo "3. Reload Nginx:"
echo "   sudo systemctl reload nginx"
