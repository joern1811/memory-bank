# Memory Bank Deployment Guide

This guide covers deploying Memory Bank in various environments, from development to production, including containerization, scaling, and operational considerations.

## Table of Contents

- [Overview](#overview)
- [Development Deployment](#development-deployment)
- [Production Deployment](#production-deployment)
- [Container Deployment](#container-deployment)
- [Cloud Deployment](#cloud-deployment)
- [High Availability Setup](#high-availability-setup)
- [Monitoring and Observability](#monitoring-and-observability)
- [Security Considerations](#security-considerations)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Optimization](#performance-optimization)
- [Troubleshooting](#troubleshooting)

## Overview

Memory Bank can be deployed in several configurations:

1. **Single Instance**: Basic deployment for individual developers
2. **Team Instance**: Shared deployment for development teams
3. **Production Instance**: High-availability deployment with monitoring
4. **Cloud Native**: Kubernetes-based deployment with auto-scaling
5. **Hybrid**: On-premises with cloud backup and disaster recovery

### Deployment Components

```
┌─────────────────────────────────────────────────────────┐
│                Memory Bank Instance                     │
│  ┌─────────────────────────────────────────────────────┐ │
│  │              Core Application                       │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │ │
│  │  │ MCP Server  │ │ CLI Client  │ │ HTTP API    │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘   │ │
│  └─────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────┐ │
│  │              Data Layer                             │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │ │
│  │  │ SQLite DB   │ │ Ollama      │ │ ChromaDB    │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘   │ │
│  └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Development Deployment

### Local Development Setup

#### Quick Setup (Mock Providers)

```bash
# 1. Build Memory Bank
cd /path/to/memory-bank
go build ./cmd/memory-bank

# 2. Initialize configuration
mkdir -p ~/.memory-bank
cat > ~/.memory-bank/config.yaml << EOF
database:
  path: "~/.memory-bank/memory_bank.db"

embedding:
  provider: "mock"  # Fast for development

vector:
  provider: "mock"  # No external dependencies

logging:
  level: "debug"
  format: "text"    # Readable for development
EOF

# 3. Test installation
./memory-bank --help
./memory-bank init . --name "Test Project"
```

#### Full Development Setup (External Services)

```bash
# 1. Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. Pull embedding model
ollama pull nomic-embed-text

# 3. Install ChromaDB
# Option A: Docker
docker run -d --name chroma -p 8000:8000 -v chroma-data:/chroma/chroma chromadb/chroma

# Option B: Native with uvx
uvx --from "chromadb[server]" chroma run --host 0.0.0.0 --port 8000 --path ./chroma_data &

# 4. Configure Memory Bank
cat > ~/.memory-bank/config.yaml << EOF
database:
  path: "~/.memory-bank/memory_bank.db"

embedding:
  provider: "ollama"
  ollama:
    base_url: "http://localhost:11434"
    model: "nomic-embed-text"
    timeout: "30s"

vector:
  provider: "chromadb"
  chromadb:
    base_url: "http://localhost:8000"
    collection: "memory_bank_dev"
    timeout: "10s"

logging:
  level: "debug"
  format: "json"
EOF

# 5. Verify setup
./memory-bank health check --verbose
```

### Development Environment Variables

```bash
# Development configuration
export MEMORY_BANK_ENV=development
export MEMORY_BANK_LOG_LEVEL=debug
export MEMORY_BANK_DB_PATH="$HOME/.memory-bank/dev_memory.db"

# External services
export OLLAMA_BASE_URL="http://localhost:11434"
export CHROMADB_BASE_URL="http://localhost:8000"

# Development features
export MEMORY_BANK_ENABLE_PROFILING=true
export MEMORY_BANK_ENABLE_DEBUG_ENDPOINTS=true
```

## Production Deployment

### Production Requirements

- **Hardware**: Minimum 2 CPU cores, 4GB RAM, 50GB storage
- **Operating System**: Linux (Ubuntu 20.04+, CentOS 8+, or RHEL 8+)
- **Network**: Outbound HTTPS for initial setup, inbound only for web UI
- **Security**: SSL/TLS certificates, firewall configuration
- **Monitoring**: Log aggregation, metrics collection, alerting

### Production Installation

#### Step 1: System Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Create dedicated user
sudo useradd -r -s /bin/false -d /opt/memory-bank memory-bank
sudo mkdir -p /opt/memory-bank/{bin,data,config,logs}
sudo chown -R memory-bank:memory-bank /opt/memory-bank

# Install required packages
sudo apt install -y sqlite3 curl wget jq
```

#### Step 2: Install Memory Bank

```bash
# Download or build Memory Bank
wget https://github.com/memory-bank/memory-bank/releases/download/v1.10.0/memory-bank-linux-amd64.tar.gz
tar -xzf memory-bank-linux-amd64.tar.gz

# Install binary
sudo cp memory-bank /opt/memory-bank/bin/
sudo chmod +x /opt/memory-bank/bin/memory-bank
sudo chown memory-bank:memory-bank /opt/memory-bank/bin/memory-bank

# Create symlink
sudo ln -sf /opt/memory-bank/bin/memory-bank /usr/local/bin/memory-bank
```

#### Step 3: Configure External Services

```bash
# Install and configure Ollama
curl -fsSL https://ollama.com/install.sh | sh
sudo systemctl enable ollama
sudo systemctl start ollama

# Pull embedding model
sudo -u ollama ollama pull nomic-embed-text

# Install ChromaDB with Docker
sudo docker run -d \
  --name memory-bank-chromadb \
  --restart unless-stopped \
  -p 127.0.0.1:8000:8000 \
  -v /opt/memory-bank/data/chroma:/chroma/chroma \
  chromadb/chroma:latest
```

#### Step 4: Production Configuration

```bash
# Create production configuration
sudo tee /opt/memory-bank/config/config.yaml << EOF
environment: "production"

database:
  path: "/opt/memory-bank/data/memory_bank.db"
  max_connections: 25
  connection_timeout: "10s"
  backup_enabled: true
  backup_interval: "24h"

embedding:
  provider: "ollama"
  ollama:
    base_url: "http://localhost:11434"
    model: "nomic-embed-text"
    timeout: "60s"
    max_retries: 3

vector:
  provider: "chromadb"
  chromadb:
    base_url: "http://localhost:8000"
    collection: "memory_bank_prod"
    timeout: "30s"
    max_retries: 3

performance:
  embedding:
    cache_enabled: true
    cache_size: 1000
    cache_ttl: "1h"
    batch_size: 10
  
  search:
    result_cache_enabled: true
    result_cache_size: 100
    result_cache_ttl: "10m"

security:
  encryption:
    enabled: true
    algorithm: "AES-256-GCM"
    key_source: "environment"
  
  audit:
    enabled: true
    log_file: "/opt/memory-bank/logs/audit.log"
    events: ["create", "update", "delete", "search"]

logging:
  level: "info"
  format: "json"
  output: "/opt/memory-bank/logs/memory-bank.log"
  max_size: "100MB"
  max_backups: 10
  max_age: "30d"
  compress: true

monitoring:
  metrics_enabled: true
  metrics_port: 9090
  health_check_port: 8080
EOF

sudo chown memory-bank:memory-bank /opt/memory-bank/config/config.yaml
sudo chmod 600 /opt/memory-bank/config/config.yaml
```

#### Step 5: Systemd Service

```bash
# Create systemd service
sudo tee /etc/systemd/system/memory-bank.service << EOF
[Unit]
Description=Memory Bank Semantic Memory Management System
After=network.target
Wants=ollama.service chromadb.service

[Service]
Type=simple
User=memory-bank
Group=memory-bank
WorkingDirectory=/opt/memory-bank
ExecStart=/opt/memory-bank/bin/memory-bank server
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=process
Restart=on-failure
RestartSec=10s
LimitNOFILE=65536

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/memory-bank/data /opt/memory-bank/logs

# Environment
Environment=MEMORY_BANK_CONFIG=/opt/memory-bank/config/config.yaml
Environment=MEMORY_BANK_ENCRYPTION_KEY_FILE=/opt/memory-bank/config/encryption.key

[Install]
WantedBy=multi-user.target
EOF

# Generate encryption key
sudo openssl rand -base64 32 > /opt/memory-bank/config/encryption.key
sudo chown memory-bank:memory-bank /opt/memory-bank/config/encryption.key
sudo chmod 600 /opt/memory-bank/config/encryption.key

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable memory-bank
sudo systemctl start memory-bank
sudo systemctl status memory-bank
```

### Production Health Checks

```bash
# Check service status
sudo systemctl status memory-bank

# Check logs
sudo journalctl -u memory-bank -f

# Health check endpoint
curl http://localhost:8080/health

# Check external services
curl http://localhost:11434/api/tags
curl http://localhost:8000/api/v2/heartbeat

# Test basic functionality
sudo -u memory-bank memory-bank health check --verbose
```

## Container Deployment

### Docker Deployment

#### Single Container (Development)

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o memory-bank ./cmd/memory-bank

FROM alpine:3.18

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

COPY --from=builder /app/memory-bank .

# Create directories
RUN mkdir -p /data /config /logs

# Default configuration
COPY deploy/docker/config.yaml /config/

EXPOSE 8080 9090
VOLUME ["/data", "/logs"]

CMD ["./memory-bank", "server", "--config", "/config/config.yaml"]
```

```bash
# Build and run
docker build -t memory-bank:latest .

docker run -d \
  --name memory-bank \
  -p 8080:8080 \
  -p 9090:9090 \
  -v memory-bank-data:/data \
  -v memory-bank-logs:/logs \
  -e MEMORY_BANK_ENCRYPTION_KEY="$(openssl rand -base64 32)" \
  memory-bank:latest
```

#### Multi-Container with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  memory-bank:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - memory-bank-data:/data
      - memory-bank-logs:/logs
      - ./config:/config
    environment:
      - MEMORY_BANK_CONFIG=/config/config.yaml
      - MEMORY_BANK_ENCRYPTION_KEY_FILE=/run/secrets/encryption_key
    secrets:
      - encryption_key
    depends_on:
      - ollama
      - chromadb
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  ollama:
    image: ollama/ollama:latest
    ports:
      - "127.0.0.1:11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    restart: unless-stopped
    command: ["serve"]

  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "127.0.0.1:8000:8000"
    volumes:
      - chromadb-data:/chroma/chroma
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/api/v2/heartbeat"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Optional: Monitoring stack
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "127.0.0.1:9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "127.0.0.1:3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

volumes:
  memory-bank-data:
  memory-bank-logs:
  ollama-data:
  chromadb-data:
  prometheus-data:
  grafana-data:

secrets:
  encryption_key:
    file: ./secrets/encryption.key
```

```bash
# Deploy with Docker Compose
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f memory-bank

# Scale if needed
docker-compose up -d --scale memory-bank=3
```

## Cloud Deployment

### AWS Deployment

#### EC2 Deployment with Auto Scaling

```bash
#!/bin/bash
# user-data.sh for EC2 instances

# Update system
yum update -y
yum install -y docker

# Start Docker
systemctl start docker
systemctl enable docker

# Download Memory Bank
wget -O /tmp/memory-bank-linux-amd64.tar.gz \
  https://github.com/memory-bank/memory-bank/releases/download/v1.10.0/memory-bank-linux-amd64.tar.gz
tar -xzf /tmp/memory-bank-linux-amd64.tar.gz -C /usr/local/bin/

# Get instance metadata
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
AZ=$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)

# Configure Memory Bank
mkdir -p /opt/memory-bank/{data,config,logs}
cat > /opt/memory-bank/config/config.yaml << EOF
environment: "production"
instance:
  id: "$INSTANCE_ID"
  availability_zone: "$AZ"

database:
  path: "/opt/memory-bank/data/memory_bank.db"

embedding:
  provider: "ollama"
  ollama:
    base_url: "http://localhost:11434"

vector:
  provider: "chromadb"
  chromadb:
    base_url: "http://chromadb.internal:8000"

logging:
  level: "info"
  format: "json"
  output: "/opt/memory-bank/logs/memory-bank.log"

monitoring:
  cloudwatch:
    enabled: true
    region: "us-west-2"
    log_group: "/aws/memory-bank"
EOF

# Start services
docker run -d --name ollama --restart unless-stopped \
  -p 11434:11434 \
  ollama/ollama:latest

# Wait for Ollama to start
sleep 30
docker exec ollama ollama pull nomic-embed-text

# Start Memory Bank
/usr/local/bin/memory-bank server --config /opt/memory-bank/config/config.yaml
```

#### CloudFormation Template

```yaml
# memory-bank-stack.yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Memory Bank deployment with Auto Scaling'

Parameters:
  InstanceType:
    Type: String
    Default: t3.medium
    Description: EC2 instance type
  
  KeyName:
    Type: AWS::EC2::KeyPair::KeyName
    Description: EC2 Key Pair for SSH access

Resources:
  MemoryBankVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true
      EnableDnsSupport: true

  MemoryBankSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref MemoryBankVPC
      CidrBlock: 10.0.1.0/24
      AvailabilityZone: !Select [0, !GetAZs '']

  MemoryBankSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for Memory Bank
      VpcId: !Ref MemoryBankVPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 22
          ToPort: 22
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 8080
          ToPort: 8080
          CidrIp: 10.0.0.0/16

  MemoryBankLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: memory-bank-template
      LaunchTemplateData:
        ImageId: ami-0c02fb55956c7d316  # Amazon Linux 2
        InstanceType: !Ref InstanceType
        KeyName: !Ref KeyName
        SecurityGroupIds:
          - !Ref MemoryBankSecurityGroup
        UserData:
          Fn::Base64: !Sub |
            #!/bin/bash
            # Insert user-data.sh content here

  MemoryBankAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      LaunchTemplate:
        LaunchTemplateId: !Ref MemoryBankLaunchTemplate
        Version: !GetAtt MemoryBankLaunchTemplate.LatestVersionNumber
      MinSize: 1
      MaxSize: 3
      DesiredCapacity: 2
      VPCZoneIdentifier:
        - !Ref MemoryBankSubnet
      HealthCheckType: ELB
      HealthCheckGracePeriod: 300

  MemoryBankLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Type: application
      Scheme: internal
      SecurityGroups:
        - !Ref MemoryBankSecurityGroup
      Subnets:
        - !Ref MemoryBankSubnet

  MemoryBankTargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Port: 8080
      Protocol: HTTP
      VpcId: !Ref MemoryBankVPC
      HealthCheckPath: /health
      HealthCheckIntervalSeconds: 30
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3
```

### Kubernetes Deployment

#### Basic Deployment

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: memory-bank
  labels:
    name: memory-bank

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: memory-bank-config
  namespace: memory-bank
data:
  config.yaml: |
    environment: "production"
    
    database:
      path: "/data/memory_bank.db"
    
    embedding:
      provider: "ollama"
      ollama:
        base_url: "http://ollama:11434"
    
    vector:
      provider: "chromadb"
      chromadb:
        base_url: "http://chromadb:8000"
    
    logging:
      level: "info"
      format: "json"

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: memory-bank-secrets
  namespace: memory-bank
type: Opaque
stringData:
  encryption-key: "your-base64-encoded-encryption-key"

---
# k8s/persistent-volumes.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: memory-bank-data
  namespace: memory-bank
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-bank
  namespace: memory-bank
  labels:
    app: memory-bank
spec:
  replicas: 2
  selector:
    matchLabels:
      app: memory-bank
  template:
    metadata:
      labels:
        app: memory-bank
    spec:
      containers:
      - name: memory-bank
        image: memory-bank:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: MEMORY_BANK_CONFIG
          value: "/config/config.yaml"
        - name: MEMORY_BANK_ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: memory-bank-secrets
              key: encryption-key
        volumeMounts:
        - name: config
          mountPath: /config
        - name: data
          mountPath: /data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
      volumes:
      - name: config
        configMap:
          name: memory-bank-config
      - name: data
        persistentVolumeClaim:
          claimName: memory-bank-data

---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: memory-bank
  namespace: memory-bank
  labels:
    app: memory-bank
spec:
  selector:
    app: memory-bank
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: memory-bank
  namespace: memory-bank
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - memory-bank.yourdomain.com
    secretName: memory-bank-tls
  rules:
  - host: memory-bank.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: memory-bank
            port:
              number: 8080

---
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: memory-bank-hpa
  namespace: memory-bank
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: memory-bank
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n memory-bank
kubectl get svc -n memory-bank
kubectl logs -f deployment/memory-bank -n memory-bank

# Scale deployment
kubectl scale deployment memory-bank --replicas=5 -n memory-bank
```

## High Availability Setup

### Load Balancer Configuration

#### NGINX Configuration

```nginx
# /etc/nginx/sites-available/memory-bank
upstream memory_bank_backend {
    least_conn;
    server 10.0.1.10:8080 max_fails=3 fail_timeout=30s;
    server 10.0.1.11:8080 max_fails=3 fail_timeout=30s;
    server 10.0.1.12:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    listen 443 ssl http2;
    server_name memory-bank.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/memory-bank.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/memory-bank.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://memory_bank_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Main application
    location / {
        proxy_pass http://memory_bank_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Error handling
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
        proxy_next_upstream_tries 3;
        proxy_next_upstream_timeout 30s;
    }

    # Metrics endpoint (restrict access)
    location /metrics {
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
        
        proxy_pass http://memory_bank_backend;
        proxy_set_header Host $host;
    }
}
```

### Database High Availability

#### SQLite with Replication (Litestream)

```bash
# Install Litestream
wget https://github.com/benbjohnson/litestream/releases/download/v0.3.9/litestream-v0.3.9-linux-amd64.tar.gz
tar -xzf litestream-v0.3.9-linux-amd64.tar.gz
sudo mv litestream /usr/local/bin/

# Configure Litestream
sudo tee /etc/litestream.yml << EOF
dbs:
  - path: /opt/memory-bank/data/memory_bank.db
    replicas:
      - type: s3
        bucket: memory-bank-backups
        path: db
        region: us-west-2
        access-key-id: \$AWS_ACCESS_KEY_ID
        secret-access-key: \$AWS_SECRET_ACCESS_KEY
        sync-interval: 1s
        retention: 72h
      - type: file
        path: /backup/memory_bank.db
        sync-interval: 10s
        retention: 24h
EOF

# Create systemd service for Litestream
sudo tee /etc/systemd/system/litestream.service << EOF
[Unit]
Description=Litestream
After=network.target

[Service]
Type=exec
User=memory-bank
Group=memory-bank
ExecStart=/usr/local/bin/litestream replicate -config /etc/litestream.yml
Environment=AWS_ACCESS_KEY_ID=your-access-key
Environment=AWS_SECRET_ACCESS_KEY=your-secret-key
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable litestream
sudo systemctl start litestream
```

## Monitoring and Observability

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "memory_bank_rules.yml"

scrape_configs:
  - job_name: 'memory-bank'
    static_configs:
      - targets: ['memory-bank:9090']
    scrape_interval: 10s
    metrics_path: /metrics
    
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

```yaml
# monitoring/memory_bank_rules.yml
groups:
  - name: memory_bank
    rules:
      - alert: MemoryBankDown
        expr: up{job="memory-bank"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Memory Bank instance is down"
          description: "Memory Bank instance {{ $labels.instance }} has been down for more than 1 minute."

      - alert: MemoryBankHighMemoryUsage
        expr: memory_bank_memory_usage_bytes / memory_bank_memory_total_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Memory Bank high memory usage"
          description: "Memory Bank instance {{ $labels.instance }} memory usage is above 90%"

      - alert: MemoryBankHighCPUUsage
        expr: rate(memory_bank_cpu_usage_seconds_total[5m]) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Memory Bank high CPU usage"
          description: "Memory Bank instance {{ $labels.instance }} CPU usage is above 80%"

      - alert: MemoryBankDatabaseConnections
        expr: memory_bank_database_connections_active > 20
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Memory Bank high database connections"
          description: "Memory Bank instance {{ $labels.instance }} has {{ $value }} active database connections"

      - alert: MemoryBankEmbeddingLatency
        expr: histogram_quantile(0.95, rate(memory_bank_embedding_duration_seconds_bucket[5m])) > 5
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "Memory Bank high embedding latency"
          description: "Memory Bank embedding generation 95th percentile latency is {{ $value }}s"
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Memory Bank Dashboard",
    "panels": [
      {
        "title": "Service Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"memory-bank\"}",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "memory_bank_memory_usage_bytes",
            "legendFormat": "Memory Used"
          },
          {
            "expr": "memory_bank_memory_total_bytes",
            "legendFormat": "Memory Total"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(memory_bank_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Times",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(memory_bank_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(memory_bank_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Database Operations",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(memory_bank_database_operations_total[5m])",
            "legendFormat": "{{operation}}"
          }
        ]
      },
      {
        "title": "Embedding Generation",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(memory_bank_embeddings_generated_total[5m])",
            "legendFormat": "Embeddings/sec"
          },
          {
            "expr": "histogram_quantile(0.95, rate(memory_bank_embedding_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile latency"
          }
        ]
      }
    ]
  }
}
```

### Log Aggregation

#### ELK Stack Configuration

```yaml
# logging/filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /opt/memory-bank/logs/*.log
  fields:
    service: memory-bank
    environment: production
  fields_under_root: true
  json:
    keys_under_root: true
    add_error_key: true

output.logstash:
  hosts: ["logstash:5044"]

processors:
- add_host_metadata:
    when.not.contains.tags: forwarded
```

```ruby
# logging/logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [service] == "memory-bank" {
    if [level] {
      mutate {
        uppercase => [ "level" ]
      }
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
    
    if [error] {
      mutate {
        add_tag => [ "error" ]
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "memory-bank-%{+YYYY.MM.dd}"
  }
}
```

## Security Considerations

### Network Security

```bash
# Firewall configuration (UFW)
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH
sudo ufw allow 22/tcp

# Allow Memory Bank (internal only)
sudo ufw allow from 10.0.0.0/8 to any port 8080
sudo ufw allow from 172.16.0.0/12 to any port 8080
sudo ufw allow from 192.168.0.0/16 to any port 8080

# Allow monitoring (restricted)
sudo ufw allow from 10.0.1.0/24 to any port 9090

# Enable firewall
sudo ufw enable
```

### SSL/TLS Configuration

```bash
# Generate SSL certificate with Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d memory-bank.yourdomain.com

# Or create self-signed certificate for internal use
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/memory-bank-selfsigned.key \
  -out /etc/ssl/certs/memory-bank-selfsigned.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=memory-bank.local"
```

### Access Control

```yaml
# config/security.yaml
security:
  authentication:
    enabled: true
    method: "jwt"
    jwt:
      secret: "${JWT_SECRET}"
      expiry: "24h"
  
  authorization:
    enabled: true
    roles:
      admin:
        permissions: ["read", "write", "delete", "admin"]
      developer:
        permissions: ["read", "write"]
      readonly:
        permissions: ["read"]
  
  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst: 10
  
  cors:
    enabled: true
    allowed_origins: ["https://yourdomain.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["Authorization", "Content-Type"]
```

## Backup and Recovery

### Automated Backup Script

```bash
#!/bin/bash
# backup-memory-bank.sh

set -euo pipefail

BACKUP_DIR="/backup/memory-bank"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Stop Memory Bank service
sudo systemctl stop memory-bank

# Create backup
tar -czf "$BACKUP_DIR/memory-bank-backup-$DATE.tar.gz" \
  -C /opt/memory-bank \
  data/ config/ logs/

# Backup database separately with SQLite backup command
sqlite3 /opt/memory-bank/data/memory_bank.db ".backup $BACKUP_DIR/memory_bank-$DATE.db"

# Start Memory Bank service
sudo systemctl start memory-bank

# Upload to S3 (optional)
if command -v aws &> /dev/null; then
  aws s3 cp "$BACKUP_DIR/memory-bank-backup-$DATE.tar.gz" \
    s3://memory-bank-backups/daily/
  aws s3 cp "$BACKUP_DIR/memory_bank-$DATE.db" \
    s3://memory-bank-backups/daily/
fi

# Clean old backups
find "$BACKUP_DIR" -name "memory-bank-backup-*.tar.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "memory_bank-*.db" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: memory-bank-backup-$DATE.tar.gz"
```

### Recovery Procedures

```bash
#!/bin/bash
# restore-memory-bank.sh

BACKUP_FILE="$1"
RESTORE_DATE=$(date +%Y%m%d_%H%M%S)

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup-file>"
  exit 1
fi

# Stop Memory Bank service
sudo systemctl stop memory-bank

# Backup current state
tar -czf "/backup/memory-bank-pre-restore-$RESTORE_DATE.tar.gz" \
  -C /opt/memory-bank \
  data/ config/ logs/

# Restore from backup
tar -xzf "$BACKUP_FILE" -C /opt/memory-bank/

# Fix permissions
sudo chown -R memory-bank:memory-bank /opt/memory-bank/

# Start Memory Bank service
sudo systemctl start memory-bank

# Verify restoration
sleep 10
sudo systemctl status memory-bank
memory-bank health check

echo "Restoration completed from $BACKUP_FILE"
```

## Performance Optimization

### Database Optimization

```bash
# SQLite optimization script
sqlite3 /opt/memory-bank/data/memory_bank.db << EOF
-- Analyze database statistics
ANALYZE;

-- Vacuum to reclaim space
VACUUM;

-- Set performance pragmas
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;  -- 64MB cache
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456;  -- 256MB mmap

-- Create additional indexes for common queries
CREATE INDEX IF NOT EXISTS idx_memories_content_length ON memories(length(content));
CREATE INDEX IF NOT EXISTS idx_memories_created_date ON memories(date(created_at));
CREATE INDEX IF NOT EXISTS idx_sessions_project_status ON sessions(project_id, status);

-- Update statistics
ANALYZE;
EOF
```

### Application Tuning

```yaml
# config/performance.yaml
performance:
  # Connection pooling
  database:
    max_connections: 25
    max_idle_connections: 5
    connection_lifetime: "1h"
  
  # Embedding optimization
  embedding:
    batch_size: 20
    max_concurrent: 5
    cache_enabled: true
    cache_size: 2000
    cache_ttl: "2h"
  
  # Vector search optimization
  vector:
    batch_size: 100
    timeout: "30s"
    max_retries: 3
    cache_enabled: true
    cache_size: 500
    cache_ttl: "15m"
  
  # Memory management
  memory:
    gc_percent: 100
    max_heap_size: "1GB"
  
  # Request handling
  http:
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    max_header_bytes: 1048576  # 1MB
```

### System-Level Optimization

```bash
# System tuning for Memory Bank
cat >> /etc/sysctl.conf << EOF
# Network optimization
net.core.somaxconn = 1024
net.ipv4.tcp_max_syn_backlog = 1024
net.ipv4.tcp_fin_timeout = 30

# Memory optimization
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# File system optimization
fs.file-max = 65536
EOF

sysctl -p

# Set resource limits for memory-bank user
cat >> /etc/security/limits.conf << EOF
memory-bank soft nofile 65536
memory-bank hard nofile 65536
memory-bank soft nproc 4096
memory-bank hard nproc 4096
EOF
```

## Troubleshooting

### Common Deployment Issues

#### Service Won't Start

```bash
# Check service status
sudo systemctl status memory-bank

# Check logs
sudo journalctl -u memory-bank -f

# Check configuration
memory-bank config validate

# Check dependencies
curl http://localhost:11434/api/tags
curl http://localhost:8000/api/v2/heartbeat

# Check file permissions
ls -la /opt/memory-bank/data/
sudo -u memory-bank ls -la /opt/memory-bank/
```

#### Database Issues

```bash
# Check database integrity
sqlite3 /opt/memory-bank/data/memory_bank.db "PRAGMA integrity_check;"

# Check database size and locks
ls -lh /opt/memory-bank/data/
lsof /opt/memory-bank/data/memory_bank.db

# Test database access
sudo -u memory-bank sqlite3 /opt/memory-bank/data/memory_bank.db ".tables"
```

#### Performance Issues

```bash
# Check resource usage
top -p $(pgrep memory-bank)
htop -p $(pgrep memory-bank)

# Check memory usage
sudo cat /proc/$(pgrep memory-bank)/status | grep -E "(VmPeak|VmSize|VmRSS)"

# Check disk I/O
sudo iotop -p $(pgrep memory-bank)

# Check network connections
netstat -an | grep :8080
ss -an | grep :8080
```

#### External Service Issues

```bash
# Ollama troubleshooting
curl -v http://localhost:11434/api/tags
ollama list
ollama ps
docker logs ollama  # if using Docker

# ChromaDB troubleshooting
curl -v http://localhost:8000/api/v2/heartbeat
curl http://localhost:8000/api/v2/collections
docker logs chromadb  # if using Docker

# Check service connectivity
telnet localhost 11434
telnet localhost 8000
```

### Health Monitoring Script

```bash
#!/bin/bash
# health-check.sh

set -euo pipefail

HEALTH_ENDPOINT="http://localhost:8080/health"
METRICS_ENDPOINT="http://localhost:9090/metrics"
LOG_FILE="/var/log/memory-bank-health.log"

log() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$LOG_FILE"
}

check_service() {
  if ! systemctl is-active --quiet memory-bank; then
    log "ERROR: Memory Bank service is not running"
    return 1
  fi
  log "INFO: Memory Bank service is running"
}

check_health_endpoint() {
  if ! curl -f -s "$HEALTH_ENDPOINT" > /dev/null; then
    log "ERROR: Health endpoint is not responding"
    return 1
  fi
  log "INFO: Health endpoint is responding"
}

check_external_services() {
  # Check Ollama
  if ! curl -f -s http://localhost:11434/api/tags > /dev/null; then
    log "WARNING: Ollama service is not responding"
  else
    log "INFO: Ollama service is responding"
  fi
  
  # Check ChromaDB
  if ! curl -f -s http://localhost:8000/api/v2/heartbeat > /dev/null; then
    log "WARNING: ChromaDB service is not responding"
  else
    log "INFO: ChromaDB service is responding"
  fi
}

check_disk_space() {
  local usage=$(df /opt/memory-bank/data | awk 'NR==2 {print $5}' | sed 's/%//')
  if [ "$usage" -gt 90 ]; then
    log "ERROR: Disk usage is at ${usage}%"
    return 1
  elif [ "$usage" -gt 80 ]; then
    log "WARNING: Disk usage is at ${usage}%"
  else
    log "INFO: Disk usage is at ${usage}%"
  fi
}

check_memory_usage() {
  local memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
  if (( $(echo "$memory_usage > 90" | bc -l) )); then
    log "ERROR: Memory usage is at ${memory_usage}%"
    return 1
  elif (( $(echo "$memory_usage > 80" | bc -l) )); then
    log "WARNING: Memory usage is at ${memory_usage}%"
  else
    log "INFO: Memory usage is at ${memory_usage}%"
  fi
}

main() {
  log "Starting health check..."
  
  local exit_code=0
  
  check_service || exit_code=1
  check_health_endpoint || exit_code=1
  check_external_services
  check_disk_space || exit_code=1
  check_memory_usage || exit_code=1
  
  if [ $exit_code -eq 0 ]; then
    log "Health check completed successfully"
  else
    log "Health check completed with errors"
  fi
  
  return $exit_code
}

main "$@"
```

---

This deployment guide provides comprehensive coverage for deploying Memory Bank in various environments, from development to production. Choose the deployment strategy that best fits your infrastructure and requirements.

For additional support or specific deployment questions, refer to the [troubleshooting guide](troubleshooting.md) or create an issue in the project repository.