# GoDrive - Google Cloud Platform Deployment Guide

Complete guide to deploying GoDrive on a Google Cloud VM with full monitoring stack.

---

## 🎯 Overview

This deployment gives you:
- ✅ Full control over infrastructure
- ✅ All services running (API, PostgreSQL, MinIO, Prometheus, Grafana)
- ✅ Cost-effective (~$10-15/month for e2-medium)
- ✅ Easy scaling and customization

---

## 📋 Prerequisites

1. **Google Cloud Account** with billing enabled
2. **gcloud CLI** installed ([Install Guide](https://cloud.google.com/sdk/docs/install))
3. **SSH key** for secure access

---

## 🚀 Step 1: Create GCP VM Instance

### Option A: Using gcloud CLI (Recommended)

```bash
# Set your project
gcloud config set project YOUR_PROJECT_ID

# Create VM instance
gcloud compute instances create godrive-vm \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --boot-disk-size=30GB \
  --boot-disk-type=pd-standard \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --tags=http-server,https-server \
  --metadata=startup-script='#!/bin/bash
    apt-get update
    apt-get install -y git curl'

# Create firewall rules
gcloud compute firewall-rules create allow-godrive-api \
  --allow=tcp:8081 \
  --target-tags=http-server \
  --description="Allow GoDrive API traffic"

gcloud compute firewall-rules create allow-grafana \
  --allow=tcp:3001 \
  --target-tags=http-server \
  --description="Allow Grafana traffic"

gcloud compute firewall-rules create allow-prometheus \
  --allow=tcp:9091 \
  --target-tags=http-server \
  --description="Allow Prometheus traffic"

gcloud compute firewall-rules create allow-minio-console \
  --allow=tcp:9003 \
  --target-tags=http-server \
  --description="Allow MinIO Console traffic"
```

### Option B: Using GCP Console

1. Go to **Compute Engine** → **VM Instances**
2. Click **Create Instance**
3. Configure:
   - **Name**: `godrive-vm`
   - **Region**: `us-central1` (or nearest to you)
   - **Machine type**: `e2-medium` (2 vCPU, 4GB RAM)
   - **Boot disk**: Ubuntu 22.04 LTS, 30GB
   - **Firewall**: Allow HTTP and HTTPS traffic
4. Click **Create**

---

## 🔧 Step 2: Connect to VM and Setup

```bash
# SSH into your VM
gcloud compute ssh godrive-vm --zone=us-central1-a

# Download and run setup script
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/YOUR_REPO/main/scripts/setup-gcp-vm.sh -o setup.sh
chmod +x setup.sh
./setup.sh

# Log out and back in for Docker group changes
exit
gcloud compute ssh godrive-vm --zone=us-central1-a
```

---

## 📦 Step 3: Deploy Application

```bash
# Clone your repository
cd /opt/godrive
git clone https://github.com/YOUR_USERNAME/YOUR_REPO.git .

# Configure environment
cp .env.template .env
nano .env  # Edit and set strong passwords

# Make deploy script executable
chmod +x scripts/deploy.sh

# Deploy!
./scripts/deploy.sh
```

---

## 🔐 Step 4: Configure Environment Variables

Edit `/opt/godrive/.env`:

```bash
# Generate strong secrets
GODRIVE_JWT_SECRET=$(openssl rand -hex 32)
GODRIVE_JWT_REFRESH_SECRET=$(openssl rand -hex 64)
POSTGRES_PASSWORD=$(openssl rand -base64 32)
MINIO_ROOT_PASSWORD=$(openssl rand -base64 32)
GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)
```

---

## 🌐 Step 5: Access Your Services

Get your VM's external IP:
```bash
gcloud compute instances describe godrive-vm \
  --zone=us-central1-a \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)'
```

Access services at:
- **API**: `http://YOUR_VM_IP:8081`
- **Swagger Docs**: `http://YOUR_VM_IP:8081/swagger/index.html`
- **Grafana**: `http://YOUR_VM_IP:3001`
- **Prometheus**: `http://YOUR_VM_IP:9091`
- **MinIO Console**: `http://YOUR_VM_IP:9003`

---

## 🔒 Step 6: Secure Your Deployment (Production)

### A. Set up SSL with Let's Encrypt

```bash
# Install Nginx
sudo apt-get install -y nginx certbot python3-certbot-nginx

# Configure Nginx as reverse proxy
sudo nano /etc/nginx/sites-available/godrive

# Add this configuration:
```

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/godrive /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com
```

### B. Restrict Firewall Rules

```bash
# Remove public access to internal services
gcloud compute firewall-rules delete allow-prometheus
gcloud compute firewall-rules delete allow-minio-console

# Only allow HTTPS
gcloud compute firewall-rules create allow-https \
  --allow=tcp:443 \
  --target-tags=http-server
```

---

## 📊 Step 7: Set Up Monitoring

### Configure Grafana Dashboards

1. Go to `http://YOUR_VM_IP:3001`
2. Login with credentials from `.env`
3. Dashboards are auto-provisioned:
   - **API Health Dashboard**
   - **Storage & Activity Dashboard**

### Set Up Alerts (Optional)

Configure Grafana alerts to notify you via:
- Email
- Slack
- Discord
- PagerDuty

---

## 🔄 Step 8: Set Up Auto-Deployment (CI/CD)

### Using GitHub Actions

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to GCP

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to GCP VM
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.GCP_VM_IP }}
          username: ${{ secrets.GCP_VM_USER }}
          key: ${{ secrets.GCP_SSH_KEY }}
          script: |
            cd /opt/godrive
            git pull origin main
            ./scripts/deploy.sh
```

Add secrets in GitHub:
- `GCP_VM_IP`: Your VM's external IP
- `GCP_VM_USER`: Your SSH username
- `GCP_SSH_KEY`: Your private SSH key

---

## 💰 Cost Estimation

### Monthly Costs (us-central1)

| Resource | Specs | Cost |
|----------|-------|------|
| e2-medium VM | 2 vCPU, 4GB RAM | ~$24/month |
| 30GB Standard Disk | Persistent storage | ~$1.20/month |
| Network Egress | First 1GB free, then $0.12/GB | ~$5-10/month |
| **Total** | | **~$30-35/month** |

### Cost Optimization

1. **Use Preemptible VM**: Save 60-80% (~$7/month)
   ```bash
   --preemptible
   ```

2. **Use Spot VM**: Similar savings
   ```bash
   --provisioning-model=SPOT
   ```

3. **Smaller instance**: e2-small (1 vCPU, 2GB) for ~$12/month

---

## 🛠️ Maintenance Commands

```bash
# View logs
docker-compose logs -f

# Restart services
docker-compose restart

# Update application
cd /opt/godrive
git pull
./scripts/deploy.sh

# Backup database
docker-compose exec postgres pg_dump -U godrive_app godrive > backup.sql

# Restore database
docker-compose exec -T postgres psql -U godrive_app godrive < backup.sql

# Check disk usage
df -h
docker system df

# Clean up old images
docker system prune -a
```

---

## 🆘 Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs

# Check disk space
df -h

# Restart Docker
sudo systemctl restart docker
```

### Can't connect to services
```bash
# Check firewall rules
gcloud compute firewall-rules list

# Check if ports are listening
sudo netstat -tlnp | grep -E '8081|3001|9091'
```

### Out of memory
```bash
# Check memory usage
free -h

# Increase swap
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

---

## 🎓 Next Steps

1. **Set up domain name** and SSL certificates
2. **Configure automated backups** to Google Cloud Storage
3. **Set up monitoring alerts** in Grafana
4. **Deploy frontend** to Vercel/Netlify pointing to your API
5. **Set up Cloud CDN** for static assets

---

## 📚 Additional Resources

- [GCP VM Documentation](https://cloud.google.com/compute/docs)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
