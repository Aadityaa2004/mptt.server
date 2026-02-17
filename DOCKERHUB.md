# Pushing Images to Docker Hub

This guide explains how to build and push your Docker images to Docker Hub.

## Prerequisites

1. **Docker Hub Account**: Create an account at [hub.docker.com](https://hub.docker.com)
2. **Docker Installed**: Make sure Docker is installed and running on your machine

## Quick Start

### Option 1: Using the Automated Script

1. Make the script executable:
   ```bash
   chmod +x push-to-dockerhub.sh
   ```

2. Run the script with your Docker Hub username:
   ```bash
   ./push-to-dockerhub.sh your-dockerhub-username
   ```

   Or set a specific version:
   ```bash
   VERSION=v1.0.0 ./push-to-dockerhub.sh your-dockerhub-username
   ```

### Option 2: Manual Steps

#### 1. Login to Docker Hub

```bash
docker login
```

Enter your Docker Hub username and password when prompted.

#### 2. Build and Tag Images

Replace `your-dockerhub-username` with your actual Docker Hub username.

**MQTT Ingestor Service:**
```bash
docker build -t your-dockerhub-username/mptt-server-ingestor:latest -f Dockerfile .
docker tag your-dockerhub-username/mptt-server-ingestor:latest your-dockerhub-username/mptt-server-ingestor:v1.0.0
```

**MQTT Bridge:**
```bash
docker build -t your-dockerhub-username/mptt-server-bridge:latest -f ./src/production/MQT.Bridge/Dockerfile ./src/production/MQT.Bridge
```

**MQTT Mosquitto:**
```bash
docker build -t your-dockerhub-username/mptt-server-mosquitto:latest -f ./src/production/MQT.Mosquitto/Dockerfile ./src/production/MQT.Mosquitto
```

**API Service:**
```bash
docker build -t your-dockerhub-username/mptt-server-api:latest -f ./src/production/MQT.ApiService/Dockerfile .
```

**Frontend:**
```bash
docker build -t your-dockerhub-username/mptt-server-frontend:latest -f ./src/production/mqt.frontend/Dockerfile ./src/production/mqt.frontend
```

#### 3. Push Images to Docker Hub

```bash
# Push all images
docker push your-dockerhub-username/mptt-server-ingestor:latest
docker push your-dockerhub-username/mptt-server-bridge:latest
docker push your-dockerhub-username/mptt-server-mosquitto:latest
docker push your-dockerhub-username/mptt-server-api:latest
docker push your-dockerhub-username/mptt-server-frontend:latest

# Push versioned tags
docker push your-dockerhub-username/mptt-server-ingestor:v1.0.0
# ... repeat for other services
```

## Using Docker Hub Images in docker-compose

After pushing images to Docker Hub, you can update your `docker-compose.yml` to use them:

```yaml
services:
  mqtt-ingestor:
    image: your-dockerhub-username/mptt-server-ingestor:latest
    # Remove the build section if using pre-built images
    # build:
    #   context: .
    #   dockerfile: ./src/production/MQT.IngestorService/Dockerfile
```

## Best Practices

1. **Use Version Tags**: Always tag your images with version numbers (e.g., `v1.0.0`) in addition to `latest`
2. **Private Repositories**: For sensitive projects, consider using private repositories on Docker Hub
3. **Multi-Architecture**: For production, consider building multi-arch images (amd64, arm64)
4. **CI/CD**: Automate the build and push process using GitHub Actions, GitLab CI, or similar

## Troubleshooting

### Authentication Issues
- Make sure you're logged in: `docker login`
- Check your credentials are correct
- For organizations, you may need to use: `docker login -u your-username`

### Push Errors
- Ensure you have write access to the repository
- Check if the repository exists on Docker Hub (it will be created automatically on first push)
- Verify your internet connection

### Build Errors
- Make sure all dependencies are available
- Check Dockerfile paths are correct
- Ensure build context includes all necessary files

## Example: Complete Workflow

```bash
# 1. Login
docker login

# 2. Build and push (using script)
./push-to-dockerhub.sh your-username

# 3. Verify images are on Docker Hub
# Visit: https://hub.docker.com/r/your-username/mptt-server-ingestor

# 4. Pull and test
docker pull your-username/mptt-server-ingestor:latest
docker run -p 9002:9002 your-username/mptt-server-ingestor:latest
```

## Next Steps

- Set up automated builds using Docker Hub's build automation
- Configure webhooks for CI/CD pipelines
- Set up image scanning for security vulnerabilities
- Consider using Docker Hub's organization features for team collaboration
