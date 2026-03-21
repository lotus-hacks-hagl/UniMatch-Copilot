FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    bash \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install uv globally (provides uv and uvx for mcp-neo4j-cypher)
RUN pip install --no-cache-dir uv

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
RUN curl -fsSL https://claude.ai/install.sh | bash


# Copy the application code
COPY . .

# Ensure TinyFish helper script is executable inside the container
RUN chmod +x /app/.claude/skills/tinyfish-web-agent/scripts/extract.sh

# Expose port (can be overridden by PORT env)
EXPOSE 9000

# Start Uvicorn
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
