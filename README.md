# UniMatch Copilot: Architecture & Technical Instructions

## Overview
UniMatch Copilot is an AI-powered study-abroad counseling system that automates university data aggregation and student profile analysis. By combining microservices architecture with autonomous AI agents and a Knowledge Graph, the platform delivers accurate, explainable, and personalized university recommendations in real time.

---

## System Architecture
![System Architecture](https://i.ibb.co/fgpP88M/image.png)
The system follows a decoupled, event-driven microservices architecture with clear separation between orchestration, AI processing, and data layers.

---

## 1. Technology Stack

### 1.1. Frontend (Web Client)
- Vue 3 (Composition API) with Vite  
- Pinia for state management, Vue Router  
- Tailwind CSS for rapid UI development  

### 1.2. Backend (Core API & Orchestrator)
- Golang (Gin framework)  
- PostgreSQL (core business data: students, cases, recommendations)  
- Redis (caching, async coordination)  

### 1.3. AI Service (Autonomous Worker)
- Python (FastAPI)  
- Claude Agent SDK (LLM reasoning layer)  
- TinyFish MCP tools (`web_search`, `fetch_page`)  
- Neo4j (Knowledge Graph for semantic mapping)  
- PostgreSQL (job state tracking & crawl cache)  

---

## 2. Core Architectural Principles

### 2.1. Orchestrator-Worker Paradigm
The Golang backend acts as the Orchestrator, handling business logic and user interaction, while the AI service operates as an autonomous Worker. The two are strictly decoupled via API contracts, ensuring modularity and scalability.

### 2.2. Asynchronous Processing (Fire-and-Forget)
Heavy AI workflows are executed asynchronously:
- Backend sends job → receives immediate acknowledgment  
- AI service processes tasks in background queues  
- Results are returned via webhook callbacks  
- Frontend updates via polling or live refresh  

### 2.3. Polyglot Persistence
Data is distributed across specialized storage:
- Backend PostgreSQL: source of truth  
- AI PostgreSQL: temporary job state & raw data  
- Neo4j: semantic Knowledge Graph  

### 2.4. Autonomous Agentic Enrichment
The AI agent dynamically:
- identifies missing data  
- performs web research  
- enriches the Knowledge Graph  
- returns structured, incremental results  

---

# Inspiration
The study-abroad consulting process is highly manual, fragmented, and often biased by limited human knowledge. Students struggle to find accurate, up-to-date information, while counselors spend excessive time gathering and validating data. We wanted to build a system that democratizes access to high-quality guidance using AI.

---

# What it does
UniMatch Copilot:
- analyzes student profiles (GPA, IELTS, preferences)  
- aggregates real-time university data  
- maps requirements using a Knowledge Graph  
- generates personalized university recommendations  
- continuously enriches its data autonomously via AI agents  

---

# How we built it
We designed a microservices architecture with:
- Golang backend for orchestration and APIs  
- Python AI service for autonomous reasoning and data enrichment  
- Neo4j Knowledge Graph for semantic matching  
- PostgreSQL for structured data  
- Vue frontend for interactive dashboards  

AI agents use web tools to crawl, extract, and map university data into a structured graph, enabling intelligent matching between student profiles and program requirements.

---

# Challenges we ran into
- Handling inconsistent and unstructured university data from multiple sources  
- Designing a reliable async workflow between backend and AI services  
- Ensuring data consistency across PostgreSQL and Neo4j  
- Preventing hallucination and maintaining accuracy in AI outputs  
- Balancing performance with real-time data enrichment  

---

# Accomplishments that we're proud of
- Built a fully autonomous AI pipeline for data aggregation  
- Successfully integrated Knowledge Graph for explainable recommendations  
- Designed a scalable orchestrator-worker architecture  
- Delivered real-time, personalized recommendations with minimal manual input  
- Achieved clear separation between business logic and AI reasoning  

---

# What we learned
- AI systems are most effective when combined with structured data (Knowledge Graphs)  
- Asynchronous architecture is critical for AI-heavy workloads  
- Clear boundaries between services improve scalability and maintainability  
- Explainability is just as important as accuracy in AI-driven products  

---

# What's next for UniMatch Copilot
- Improve recommendation accuracy with feedback loops  
- Expand Knowledge Graph coverage (scholarships, visa policies, career outcomes)  
- Add real-time conversational AI advisor (Copilot chat)  
- Integrate with universities and official data sources  
- Support multi-country and multi-language expansion  
