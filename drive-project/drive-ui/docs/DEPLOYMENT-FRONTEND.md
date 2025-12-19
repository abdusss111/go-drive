# GoDrive Frontend - Nginx Deployment Guide

Complete guide for deploying the Next.js frontend with Nginx on GCP VM.

---

## 🎯 Overview

This setup deploys your Next.js frontend alongside the backend on the same GCP VM, using Nginx as a reverse proxy.

### Architecture
```
Internet → Nginx (Port 80/443)
  ├─→ app.yourdomain.com → Next.js App (Port 3000)
  ├─→ app.yourdomain.com/api/* → Backend API (Port 8081)
  └─→ api.yourdomain.com → Direct API access
```

---

## 🚀 Quick Deployment

### Option 1: Complete Stack (Recommended)

Deploy both backend and frontend on one VM:

```bash
# 1. SSH into your VM
gcloud compute ssh godrive-vm

# 2. Run complete setup
curl -fsSL <your-repo>/drive-ui/scripts/setup-complete-vm.sh | bash
exit  # Log out and back in

# 3. Clone repositories
gcloud compute ssh godrive-vm
cd /opt/godrive && git clone <backend-repo> .
cd /opt/godrive-ui && git clone <frontend-repo> .

# 4. Deploy backend first
cd /opt/godrive
cp .env.template .env
nano .env  # Configure
./scripts/deploy.sh

# 5. Deploy frontend
cd /opt/godrive-ui
nano .env.local  # Set NEXT_PUBLIC_API_BASE_URL=http://localhost:8081
chmod +x scripts/deploy-frontend.sh
./scripts/deploy-frontend.sh

# 6. Set up SSL
sudo certbot --nginx -d app.yourdomain.com -d api.yourdomain.com
```

### Option 2: Frontend Only

If backend is already deployed elsewhere:

```bash
cd /opt/godrive-ui
nano .env.local  # Set NEXT_PUBLIC_API_BASE_URL=https://your-api-url.com
./scripts/deploy-frontend.sh
```

---

## 📁 File Structure

```
/var/www/godrive-ui/          # Deployed app
├── .next/                     # Built Next.js files
├── public/                    # Static assets
├── package.json              # Dependencies
└── node_modules/             # Production deps

/etc/nginx/sites-available/
└── godrive-ui                # Nginx config

/opt/godrive-ui/              # Source code
├── nginx/
│   └── app.conf             # Nginx template
└── scripts/
    └── deploy-frontend.sh   # Deployment script
```

---

## 🔧 Configuration

### Environment Variables

Create `/opt/godrive-ui/.env.local`:

```bash
# API Backend URL
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081

# For production with SSL:
# NEXT_PUBLIC_API_BASE_URL=https://api.yourdomain.com
```

### Nginx Configuration

The Nginx config (`nginx/app.conf`) provides:

1. **Static File Serving**
   - Serves built Next.js files from `/var/www/godrive-ui`
   - Optimized caching for static assets

2. **API Proxy**
   - Routes `/api/*` to backend at `localhost:8081/v1/*`
   - Handles CORS automatically
   - Supports file uploads up to 100MB

3. **Client-Side Routing**
   - Fallback to `index.html` for SPA routing
   - Proper 404 handling

---

## 🌐 Domain Setup

### DNS Configuration

Point your domain to the VM's IP:

```
Type    Name    Value
A       app     <VM_IP_ADDRESS>
A       api     <VM_IP_ADDRESS>
```

### Update Nginx Configs

```bash
# Frontend
sudo nano /etc/nginx/sites-available/godrive-ui
# Change: server_name app.yourdomain.com;

# Backend
sudo nano /etc/nginx/sites-available/godrive-api
# Change: server_name api.yourdomain.com;

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🔒 SSL Setup

```bash
# Get certificates for both domains
sudo certbot --nginx \
  -d app.yourdomain.com \
  -d api.yourdomain.com \
  -d grafana.yourdomain.com

# Certbot will:
# 1. Obtain SSL certificates
# 2. Update Nginx configs
# 3. Set up auto-renewal
```

---

## 🔄 Updates and Redeployment

### Update Frontend

```bash
cd /opt/godrive-ui
git pull origin main
./scripts/deploy-frontend.sh
```

### Update Backend

```bash
cd /opt/godrive
git pull origin main
./scripts/deploy.sh
```

### Restart Services

```bash
# Restart Next.js
pm2 restart godrive-ui

# Restart backend
cd /opt/godrive && docker-compose restart

# Reload Nginx
sudo systemctl reload nginx
```

---

## 📊 Monitoring

### Check Service Status

```bash
# Next.js app
pm2 status
pm2 logs godrive-ui

# Nginx
sudo systemctl status nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Backend
cd /opt/godrive && docker-compose ps
docker-compose logs -f api
```

### Performance Monitoring

Access Grafana at `https://grafana.yourdomain.com` to monitor:
- API response times
- Error rates
- File upload/download metrics

---

## 🛠️ Troubleshooting

### Frontend won't build

```bash
cd /opt/godrive-ui
rm -rf .next node_modules
npm install
npm run build
```

### 502 Bad Gateway

```bash
# Check if Next.js is running
pm2 status

# Restart if needed
pm2 restart godrive-ui

# Check logs
pm2 logs godrive-ui
```

### API calls failing

```bash
# Check backend is running
cd /opt/godrive && docker-compose ps

# Check Nginx proxy
sudo tail -f /var/log/nginx/error.log

# Verify API URL in frontend
cat /var/www/godrive-ui/.env.local
```

### Static files not loading

```bash
# Check file permissions
ls -la /var/www/godrive-ui

# Fix if needed
sudo chown -R www-data:www-data /var/www/godrive-ui
```

---

## 🎨 Customization

### Custom Domain

Update all configs with your domain:

```bash
# Frontend Nginx
sudo nano /etc/nginx/sites-available/godrive-ui

# Frontend environment
nano /opt/godrive-ui/.env.local

# Rebuild and redeploy
cd /opt/godrive-ui
./scripts/deploy-frontend.sh
```

### Custom Port

To run Next.js on a different port:

```bash
# Edit PM2 startup
cd /var/www/godrive-ui
pm2 delete godrive-ui
PORT=3001 pm2 start npm --name "godrive-ui" -- start
pm2 save

# Update Nginx proxy_pass
sudo nano /etc/nginx/sites-available/godrive-ui
# Change: proxy_pass http://localhost:3001;
```

---

## 💰 Resource Usage

### Single VM Deployment

| Component | RAM | CPU | Disk |
|-----------|-----|-----|------|
| Next.js | ~200MB | 5-10% | - |
| Backend (Docker) | ~1GB | 10-20% | - |
| PostgreSQL | ~100MB | 5% | 2GB |
| MinIO | ~100MB | 5% | Variable |
| Nginx | ~10MB | 1% | - |
| **Total** | **~1.5GB** | **25-40%** | **5-10GB** |

**Recommended**: e2-medium (2 vCPU, 4GB RAM)

---

## 🔐 Security Best Practices

1. **Enable HTTPS** - Always use SSL in production
2. **Restrict Metrics** - Block `/metrics` endpoint from public access
3. **Rate Limiting** - Add Nginx rate limiting for API endpoints
4. **Firewall** - Only open ports 80, 443, and 22
5. **Updates** - Keep all packages updated

---

## 📚 Additional Resources

- [Next.js Deployment Docs](https://nextjs.org/docs/deployment)
- [Nginx Configuration Guide](https://nginx.org/en/docs/)
- [PM2 Documentation](https://pm2.keymetrics.io/)
- [Let's Encrypt](https://letsencrypt.org/)
