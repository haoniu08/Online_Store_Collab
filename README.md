# CS6650 Online Store - Product API

A scalable Go-based REST API for product management, deployed on AWS ECS with comprehensive load testing capabilities.

## 🏗️ Architecture Overview

- **Backend**: Go 1.24 with Gorilla Mux router
- **Storage**: Thread-safe in-memory store with sync.RWMutex
- **Containerization**: Docker multi-stage builds
- **Infrastructure**: AWS ECS Fargate with Terraform
- **Load Testing**: Locust with HttpUser vs FastHttpUser comparison

## 📁 Project Structure

```
├── cmd/
│   └── server/              # Main server application
│       ├── main.go         # Server entry point
│       └── main_test.go    # Server tests
├── internal/
│   ├── handlers/           # HTTP request handlers
│   │   └── product.go     # Product API endpoints
│   ├── models/            # Data models
│   │   ├── product.go     # Product struct and validation
│   │   └── product_test.go
│   └── store/             # Data storage layer
│       ├── product_store.go     # Thread-safe in-memory store
│       └── product_store_test.go
├── terraform/             # Infrastructure as Code
│   ├── main.tf           # Main Terraform configuration
│   ├── variables.tf      # Variable definitions
│   ├── outputs.tf        # Output values
│   └── modules/          # Terraform modules
│       ├── network/      # VPC, subnets, security groups
│       ├── ecr/          # Container registry
│       ├── ecs/          # ECS cluster and service
│       └── logging/      # CloudWatch logs
├── test_locust/          # Load testing setup
│   ├── locustfile.py     # Comprehensive load test scenarios
│   ├── simple_locustfile.py  # HttpUser vs FastHttpUser comparison
│   └── LOAD_TESTING.md   # Load testing guide
├── Dockerfile            # Multi-stage Docker build
└── README.md            # This file
```

## 🚀 Deployment Instructions

### Prerequisites

1. **Go 1.24+**
2. **Docker** 
3. **AWS CLI** configured with appropriate permissions
4. **Terraform** (latest version)
5. **Python 3.8+** (for load testing)

### 1. Local Development Setup

```bash
# Clone the repository
git clone https://github.com/haoniu08/CS6650_Online_Store.git
cd CS6650_Online_Store

# Run locally
go run cmd/server/main.go
# Server starts on http://localhost:8080

# Run tests
go test ./...
```

### 2. Docker Deployment

```bash
# Build the Docker image
docker build -t product-api .

# Run the container
docker run -p 8080:8080 product-api
```

### 3. AWS Infrastructure Deployment

```bash
# Navigate to terraform directory
cd terraform

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Deploy infrastructure
terraform apply
# Type 'yes' when prompted

# Get the public IP of your deployed service
terraform output ecs_public_ip
```

**Note**: The deployment creates:
- ECS Fargate cluster with 512 CPU / 1024 MiB memory
- Public subnets with internet gateway
- Security groups allowing HTTP traffic on port 8080
- ECR repository for container images
- CloudWatch logs for monitoring

### 4. Load Testing Setup

```bash
# Navigate to test directory
cd test_locust

# Create Python virtual environment
python -m venv venv
source venv/bin/activate  # On macOS/Linux

# Install dependencies
pip install locust

# Run load tests against local server
./venv/bin/locust -f simple_locustfile.py --host=http://localhost:8080

# Run load tests against deployed API (replace with your IP)
./venv/bin/locust -f simple_locustfile.py --host=http://YOUR_ECS_PUBLIC_IP:8080

# Open Locust web interface
# Navigate to http://localhost:8089
```

## 📡 API Endpoints

### Base URL
- **Local**: `http://localhost:8080`
- **Deployed**: `http://YOUR_ECS_PUBLIC_IP:8080` (get from `terraform output`)

### Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/products/{id}` | Retrieve product by ID |
| POST | `/products/{id}/details` | Create or update product |

## 🧪 API Testing Examples

### 1. Health Check (200 OK)

```bash
curl -X GET http://localhost:8080/health
```

**Response:**
```
Status: 200 OK
Body: OK
```

### 2. Create Product (204 No Content)

```bash
curl -X POST http://localhost:8080/products/12345/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 12345,
    "sku": "SKU-12345-ABC",
    "manufacturer": "Acme Corporation",
    "category_id": 10,
    "weight": 250,
    "some_other_id": 999
  }'
```

**Response:**
```
Status: 204 No Content
```

### 3. Get Product (200 OK)

```bash
curl -X GET http://localhost:8080/products/12345
```

**Response:**
```json
Status: 200 OK
{
  "product_id": 12345,
  "sku": "SKU-12345-ABC", 
  "manufacturer": "Acme Corporation",
  "category_id": 10,
  "weight": 250,
  "some_other_id": 999
}
```

### 4. Get Non-Existent Product (404 Not Found)

```bash
curl -X GET http://localhost:8080/products/99999
```

**Response:**
```json
Status: 404 Not Found
{
  "error": "Product not found"
}
```

### 5. Invalid Product ID (400 Bad Request)

```bash
curl -X GET http://localhost:8080/products/abc
```

**Response:**
```json
Status: 400 Bad Request  
{
  "error": "Invalid product ID: must be a positive integer"
}
```

### 6. Invalid JSON in POST Request (400 Bad Request)

```bash
curl -X POST http://localhost:8080/products/12345/details \
  -H "Content-Type: application/json" \
  -d '{"invalid_json":'
```

**Response:**
```json
Status: 400 Bad Request
{
  "error": "Invalid JSON"
}
```

### 7. Validation Error (400 Bad Request)

```bash
curl -X POST http://localhost:8080/products/12345/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": -1,
    "sku": "",
    "manufacturer": "",
    "category_id": -5,
    "weight": -100
  }'
```

**Response:**
```json
Status: 400 Bad Request
{
  "error": "Validation failed: product_id must be positive, sku is required, manufacturer is required, category_id must be positive, weight must be positive"
}
```

## 🔧 Key Files Location

| Component | File Path | Description |
|-----------|-----------|-------------|
| **Server Code** | `cmd/server/main.go` | Main application entry point |
| **API Handlers** | `internal/handlers/product.go` | HTTP request handlers |
| **Data Models** | `internal/models/product.go` | Product struct and validation |
| **Data Store** | `internal/store/product_store.go` | Thread-safe storage implementation |
| **Dockerfile** | `./Dockerfile` | Multi-stage container build |
| **Infrastructure** | `terraform/` | Complete AWS infrastructure |
| **Load Testing** | `test_locust/` | Performance testing setup |

## 🎯 Load Testing

### HttpUser vs FastHttpUser Comparison

The project includes two user classes for performance testing:

- **ProductAPIUser (HttpUser)**: Better validation and error handling
- **FastProductAPIUser (FastHttpUser)**: ~20-30% higher performance with connection pooling

### Test Scenarios

1. **Normal Load**: 100 users, 10/sec ramp-up, 5 minutes
2. **Write-Heavy Load**: 50 users, 5/sec ramp-up, 3 minutes  
3. **Spike Test**: 500 users, 25/sec ramp-up, 3 minutes

See `test_locust/LOAD_TESTING.md` for detailed instructions.

## 🔒 Security & Best Practices

### .gitignore Configuration

The project includes comprehensive `.gitignore` to exclude:

```gitignore
# Terraform sensitive files
*.tfstate
*.tfstate.*
*.tfvars
.terraform/

# Environment files
.env
.env.local
.env.production

# AWS credentials
.aws/
*.pem
*.key

# Build artifacts  
/bin/
/pkg/
vendor/

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Python
__pycache__/
*.pyc
venv/
.pytest_cache/

# Large files
*.log
*.tar.gz
*.zip
```

### Infrastructure Security

- Security groups restrict access to port 8080 only
- ECS tasks run with minimal required permissions
- Container uses non-root user in production
- CloudWatch logging for audit trails

## 🚨 Troubleshooting

### Common Issues

1. **ECS Task Exit Code 255**: Reduce CPU/memory allocation in `terraform/modules/ecs/main.tf`
2. **Large Terraform Files**: Ensure `.gitignore` excludes `*.tfstate` files
3. **Locust Import Errors**: Use full path to virtual environment locust executable
4. **Port Conflicts**: Ensure no other services running on ports 8080 or 8089

### Performance Tuning

- Monitor ECS CPU/Memory utilization in CloudWatch
- Adjust task resource allocation based on load test results
- Use FastHttpUser for high-throughput scenarios
- Scale ECS service based on demand

## 📊 Monitoring

### AWS CloudWatch Metrics

- ECS CPU utilization (target: < 80%)
- ECS Memory utilization (target: < 80%) 
- Request latency (target: < 200ms GET, < 500ms POST)
- Error rates (target: < 1%)

### Local Development

```bash
# View server logs
go run cmd/server/main.go

# Run with verbose logging
GO_LOG_LEVEL=debug go run cmd/server/main.go
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Run tests: `go test ./...`
4. Commit changes: `git commit -am 'Add feature'`
5. Push to branch: `git push origin feature-name`
6. Submit a Pull Request

## 📄 License

This project is part of CS6650 coursework at Northeastern University.