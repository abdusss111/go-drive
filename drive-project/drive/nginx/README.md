# Nginx Configuration for GoDrive

This directory contains Nginx reverse proxy configurations for the GoDrive application.

## 📁 Files

- `api.conf` - Main API backend configuration
- `grafana.conf` - Grafana dashboard configuration

## 🚀 Quick Setup

The Nginx setup is automated via the `scripts/setup-nginx.sh` script, which:
1. Installs Nginx and Certbot
2. Copies configurations to `/etc/nginx/sites-available/`
3. Enables the sites
4. Configures SSL (optional)

## 🔧 Manual Configuration

### 1. Update Domain Names

Edit the configuration files and replace placeholder domains:

```bash
# For API
sudo nano /etc/nginx/sites-available/godrive-api
# Change: server_name _; 
# To: server_name api.yourdomain.com;

# For Grafana
sudo nano /etc/nginx/sites-available/godrive-grafana
# Change: server_name grafana.yourdomain.com;
# To: your actual domain
```

### 2. Test Configuration

```bash
sudo nginx -t
```

### 3. Reload Nginx

```bash
sudo systemctl reload nginx
```

## 🔒 SSL Setup (Let's Encrypt)

After DNS is configured and pointing to your server:

```bash
# For API
sudo certbot --nginx -d api.yourdomain.com

# For Grafana
sudo certbot --nginx -d grafana.yourdomain.com
```

Certbot will automatically:
- Obtain SSL certificates
- Update Nginx configuration
- Set up auto-renewal

## 🌐 URL Structure

After setup, your services will be accessible at:

| Service | URL | Backend Port |
|---------|-----|--------------|
| API | `https://api.yourdomain.com` | 8081 |
| Swagger | `https://api.yourdomain.com/swagger/index.html` | 8081 |
| Health Check | `https://api.yourdomain.com/health/live` | 8081 |
| Metrics | `https://api.yourdomain.com/metrics` | 8081 |
| Grafana | `https://grafana.yourdomain.com` | 3001 |

## 🔐 Security Features

The configurations include:

### Security Headers
- `X-Frame-Options: SAMEORIGIN` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - XSS protection
- `Strict-Transport-Security` - Forces HTTPS (after SSL setup)

### File Upload Support
- `client_max_body_size: 100M` - Allows uploads up to 100MB
- Extended timeouts for large file transfers

### Proxy Settings
- Proper header forwarding (`X-Real-IP`, `X-Forwarded-For`)
- WebSocket support (for future features)
- HTTP/2 support (after SSL setup)

## 🛠️ Troubleshooting

### Nginx won't start
```bash
# Check configuration
sudo nginx -t

# Check logs
sudo tail -f /var/log/nginx/error.log
```

### 502 Bad Gateway
```bash
# Ensure backend services are running
docker-compose ps

# Check if ports are listening
sudo netstat -tlnp | grep -E '8081|3001'
```

### SSL certificate issues
```bash
# Test certificate renewal
sudo certbot renew --dry-run

# Check certificate status
sudo certbot certificates
```

## 📊 Performance Tuning

For production, consider adding to `api.conf`:

```nginx
# Enable gzip compression
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css application/json application/javascript;

# Enable caching for static assets
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## 🔄 Auto-Renewal

Certbot automatically sets up a systemd timer for certificate renewal. Verify with:

```bash
sudo systemctl status certbot.timer
```

## 📝 Custom Configurations

To add custom Nginx configurations:

1. Create a new file in `/etc/nginx/sites-available/`
2. Enable it: `sudo ln -s /etc/nginx/sites-available/myconfig /etc/nginx/sites-enabled/`
3. Test: `sudo nginx -t`
4. Reload: `sudo systemctl reload nginx`
