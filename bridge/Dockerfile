FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements first for better caching
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy bridge script
COPY mqtt_bridge.py .

# Make script executable
RUN chmod +x mqtt_bridge.py

# Run the bridge
CMD ["python", "mqtt_bridge.py"]
