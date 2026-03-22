FROM python:3.11-slim
# Install system dependencies
RUN apt-get update && apt-get install -y \
    bash \
    curl \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*


RUN useradd -m -u 1000 appuser
USER appuser

# Set working directory
WORKDIR /app

# Install uv globally (provides uv and uvx for mcp-neo4j-cypher)
RUN pip install --no-cache-dir uv

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
RUN curl -fsSL https://claude.ai/install.sh | bash

# Copy the application code
COPY . .

# Expose port (can be overridden by PORT env)
EXPOSE 8895

# Start Uvicorn
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8895"]
