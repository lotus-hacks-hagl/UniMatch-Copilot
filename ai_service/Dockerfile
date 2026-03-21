FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
# Node.js and npm are required for npx (agentql-mcp)
RUN apt-get update && apt-get install -y \
    curl \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

# Install uv globally (provides uv and uvx for mcp-neo4j-cypher)
RUN pip install --no-cache-dir uv

# Pre-install agentql-mcp globally to speed up startup
RUN npm install -g agentql-mcp

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the application code
COPY . .

# Expose port (can be overridden by PORT env)
EXPOSE 9000

# Start Uvicorn
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
