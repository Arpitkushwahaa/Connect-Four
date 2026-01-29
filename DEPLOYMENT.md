# Deployment Guide

This guide covers deploying the 4 in a Row application to production.

## Table of Contents
1. [Docker Deployment](#docker-deployment)
2. [Cloud Deployment](#cloud-deployment)
3. [Environment Variables](#environment-variables)
4. [SSL/HTTPS Setup](#sslhttps-setup)
5. [Monitoring](#monitoring)

## Docker Deployment

### Prerequisites
- Docker Engine 20.10+
- Docker Compose 2.0+
- 2GB RAM minimum
- 10GB disk space

### Production Docker Compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: connectfour-db
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: connectfour
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network
    restart: always

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    networks:
      - app-network
    restart: always

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/connectfour?sslmode=disable
      KAFKA_BROKER: kafka:29092
      PORT: 8080
    networks:
      - app-network
    restart: always

  frontend:
    build: ./frontend
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./ssl:/etc/nginx/ssl:ro
    networks:
      - app-network
    restart: always

volumes:
  postgres_data:

networks:
  app-network:
    driver: bridge
```

### Deploy with Docker

```bash
# Set environment variables
export DB_USER=your_db_user
export DB_PASSWORD=your_secure_password

# Build and start
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

## Cloud Deployment

### AWS Deployment

#### Using ECS (Elastic Container Service)

1. **Push images to ECR:**
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build and tag
docker build -t connectfour-backend ./backend
docker tag connectfour-backend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/connectfour-backend:latest

# Push
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/connectfour-backend:latest
```

2. **Create ECS Task Definitions** for backend, frontend, and analytics

3. **Set up RDS for PostgreSQL:**
- Create PostgreSQL instance
- Configure security groups
- Note connection endpoint

4. **Set up MSK for Kafka:**
- Create Kafka cluster
- Configure bootstrap servers
- Update backend environment variables

5. **Deploy services:**
- Create ECS cluster
- Deploy services with task definitions
- Configure load balancer for frontend

#### Using EC2

```bash
# SSH into EC2 instance
ssh -i your-key.pem ec2-user@your-instance-ip

# Install Docker
sudo yum update -y
sudo yum install docker -y
sudo service docker start
sudo usermod -a -G docker ec2-user

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Clone repository
git clone <your-repo-url>
cd Connect-four

# Deploy
docker-compose up -d
```

### Google Cloud Platform

#### Using Cloud Run

```bash
# Build and push to Container Registry
gcloud builds submit --tag gcr.io/PROJECT-ID/connectfour-backend ./backend

# Deploy to Cloud Run
gcloud run deploy connectfour-backend \
  --image gcr.io/PROJECT-ID/connectfour-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### Heroku

```bash
# Login
heroku login

# Create app
heroku create connectfour-app

# Add Postgres
heroku addons:create heroku-postgresql:hobby-dev

# Add Kafka
heroku addons:create heroku-kafka:basic-0

# Deploy
git push heroku main
```

## Environment Variables

### Production Environment Variables

Create `.env.production`:

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/dbname
DB_USER=production_user
DB_PASSWORD=secure_password_here

# Kafka
KAFKA_BROKER=your-kafka-broker:9092

# Backend
PORT=8080
ENV=production

# Frontend
REACT_APP_WS_URL=wss://your-domain.com/ws
REACT_APP_API_URL=https://your-domain.com

# Security
JWT_SECRET=your-jwt-secret-here
CORS_ORIGIN=https://your-domain.com
```

### Secrets Management

**AWS Secrets Manager:**
```bash
aws secretsmanager create-secret \
  --name connectfour/database \
  --secret-string '{"username":"dbuser","password":"dbpass"}'
```

**Environment Variables in ECS:**
- Use AWS Systems Manager Parameter Store
- Reference in task definitions

## SSL/HTTPS Setup

### Using Let's Encrypt with Nginx

1. **Install Certbot:**
```bash
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx
```

2. **Obtain Certificate:**
```bash
sudo certbot --nginx -d your-domain.com -d www.your-domain.com
```

3. **Update Nginx Config:**
```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    location /ws {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
    }
}
```

4. **Auto-renewal:**
```bash
sudo certbot renew --dry-run
```

### CloudFront with S3 (Frontend Only)

1. **Build frontend:**
```bash
cd frontend
npm run build
```

2. **Upload to S3:**
```bash
aws s3 sync build/ s3://your-bucket-name/
```

3. **Create CloudFront distribution**
4. **Configure SSL certificate in ACM**

## Monitoring

### Application Monitoring

**Prometheus + Grafana:**

Add to `docker-compose.prod.yml`:
```yaml
prometheus:
  image: prom/prometheus
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana
  ports:
    - "3001:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Log Aggregation

**ELK Stack (Elasticsearch, Logstash, Kibana):**
```yaml
elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:8.5.0
  environment:
    - discovery.type=single-node

logstash:
  image: docker.elastic.co/logstash/logstash:8.5.0
  volumes:
    - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf

kibana:
  image: docker.elastic.co/kibana/kibana:8.5.0
  ports:
    - "5601:5601"
```

### Health Checks

Add health check endpoints:

```go
// Backend health check
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "timestamp": time.Now().String(),
    })
})
```

### Monitoring Services

1. **Datadog** - Application performance monitoring
2. **New Relic** - Full-stack observability
3. **Sentry** - Error tracking
4. **CloudWatch** - AWS native monitoring

## Database Backup

### Automated Backups

```bash
#!/bin/bash
# backup.sh

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="connectfour"

docker exec connectfour-db pg_dump -U postgres $DB_NAME > $BACKUP_DIR/backup_$TIMESTAMP.sql

# Keep only last 7 days
find $BACKUP_DIR -type f -mtime +7 -delete
```

Add to cron:
```bash
0 2 * * * /path/to/backup.sh
```

## Scaling

### Horizontal Scaling

1. **Load Balancer:**
   - Use Nginx or AWS ALB
   - Distribute traffic across multiple backend instances

2. **Database:**
   - Use read replicas for read-heavy operations
   - Connection pooling

3. **Kafka:**
   - Multiple brokers for high availability
   - Partitioning for throughput

### Configuration

```yaml
# docker-compose.scale.yml
backend:
  deploy:
    replicas: 3
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
```

Deploy:
```bash
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d --scale backend=3
```

## Security Best Practices

1. **Use environment variables for secrets**
2. **Enable HTTPS/WSS only**
3. **Set up firewall rules**
4. **Regular security updates**
5. **Use strong passwords**
6. **Enable rate limiting**
7. **Implement authentication (if needed)**
8. **Regular backups**

## Performance Optimization

1. **Enable caching:**
   - Redis for session management
   - CDN for static assets

2. **Database indexing:**
   ```sql
   CREATE INDEX idx_games_completed ON games(completed_at);
   CREATE INDEX idx_leaderboard_wins ON leaderboard(wins DESC);
   ```

3. **Compression:**
   - Enable gzip in Nginx
   - Minify frontend assets

4. **Connection pooling:**
   - Configure max connections in PostgreSQL
   - Use connection pools in Go

## Rollback Strategy

1. **Tag Docker images with versions:**
```bash
docker build -t connectfour-backend:v1.0.0 ./backend
docker tag connectfour-backend:v1.0.0 connectfour-backend:latest
```

2. **Rollback:**
```bash
docker-compose down
docker-compose up -d --image connectfour-backend:v0.9.0
```

3. **Database migrations:**
   - Always have down migrations
   - Test rollback procedures

## Cost Optimization

1. **Use spot instances** (AWS)
2. **Auto-scaling** based on load
3. **Reserved instances** for predictable workloads
4. **Optimize database queries**
5. **Clean up old data**
6. **Use CDN for static content**

---

For questions or issues, refer to the main README.md or create an issue in the repository.
