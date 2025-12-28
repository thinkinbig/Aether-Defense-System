# Aether Defense System - Architecture Documentation

This document provides a comprehensive overview of the Aether Defense System architecture, covering both the high-level system layers and the cloud infrastructure deployment.

## System Architecture Overview

The Aether Defense System follows a layered microservices architecture, designed for scalability, maintainability, and separation of concerns.

![Layered System Architecture](./layer.png)

### Architecture Layers

#### 1. API Consumers
The entry point for all external interactions. The system supports multiple client types:
- **Web Applications** - Browser-based clients
- **Mobile Applications** - Native and hybrid mobile apps
- **Other Services** - Third-party integrations and service-to-service communication

All consumers interact with the system via **HTTP REST** protocols.

#### 2. API Gateway Layer
The API Gateway serves as the single entry point for all client requests, providing:
- **user-api** - User management and authentication services
- **Trade-api / promo-api** - Trading and promotion management services

The gateway handles request routing, authentication, rate limiting, and protocol translation between HTTP REST (external) and gRPC/RPC (internal).

#### 3. Domain Services Layer
Core business logic is implemented in domain-specific RPC services that communicate via **gRPC / RPC**:

- **User RPC Service** - User domain operations and business logic
- **Trade RPC Service** - Trading domain operations and business logic
- **Promotion RPC Service** - Promotion and marketing domain operations

These services communicate with each other through inter-service calls and interact with shared components and the data layer.

#### 4. Shared Components Layer
Common utilities and infrastructure components shared across all domain services:

- **Middleware & Utils** - Cross-cutting concerns and utility functions
- **Snowflake ID Generator** - Distributed unique ID generation
- **Message Queue (MQ)** - Asynchronous messaging and event processing
- **Config & Common Libs** - Shared configuration management and common libraries

#### 5. Platform / Operations Layer
DevOps and operational tooling supporting the entire system:

- **Deployment & CI/CD** - Automated deployment pipelines and continuous integration
- **Monitoring & Logging** - System observability and log aggregation
- **Testing & Integration** - Automated testing and integration validation
- **Documentation** - System documentation and API specifications

#### 6. Data Layer
Persistent data storage:

- **MySQL** - Relational database for structured data storage
- **Redis** - In-memory data store for caching and session management

---

## Cloud Infrastructure Architecture

The system is deployed on Kubernetes, providing scalability, high availability, and automated orchestration.

![Cloud Infrastructure Architecture](./cloud.png)

### External Layer

#### Users & DNS
- Users access the system through various devices (laptops, mobile phones)
- DNS resolution directs traffic to the cloud load balancer

#### Cloud Load Balancer
- Entry point for all external traffic
- Distributes load across multiple instances
- Provides high availability and fault tolerance
- Connects to:
  - **Container Registry** - Stores container images
  - **Persistent Storage** - Cloud-backed persistent volumes for durable data storage

### Kubernetes Cluster - Ingress & API Gateway

#### Ingress
- Routes external traffic into the Kubernetes cluster
- Handles SSL/TLS termination
- Provides path-based and host-based routing

#### API Gateway (namespace: aether-defense)
- **user-api** - HTTP/REST API service
  - Receives requests from external services
  - Communicates with internal RPC services
  - Accesses data stores for user information

### Kubernetes Cluster - Internal Services

#### API Gateway Layer
- **user-api** - Deployed as Kubernetes Deployment + Service
  - Exposes REST endpoints
  - Communicates with internal RPC services via gRPC

#### Internal Services (cmd/rpc)
Microservices implementing core business logic:

- **user-rpc** - User domain RPC service
- **trade-rpc** - Trading domain RPC service
- **promotion-rpc** - Promotion domain RPC service

These services communicate via gRPC and are deployed as separate microservices within the cluster.

### Data Storage

#### PVC + PV (Persistent Volume Claim + Persistent Volume)
- Provides **strong consistency** guarantees
- Manages **Secrets** and **ConfigMaps** for secure configuration
- Backed by cloud disk storage for durability

#### Redis StatefulSet
High-availability Redis cluster:

- **Redis Master** - Primary node handling write operations
  - Uses Secret and Persistent Volume for data persistence
- **Redis Replica** - Replica nodes for read scaling and high availability
  - Uses Secret and Persistent Volume for replication

Both master and replica nodes are monitored and send metrics/alerts to the observability stack.

### DevOps & Observability

#### CI/CD Pipeline

**GitHub Actions**
- Builds container images from source code
- Pushes images to the Image Registry
- Executes `helm upgrade` using charts from `deploy/helm/` directory
- Automates deployment to the Kubernetes cluster

**Image Registry**
- Stores container images
- Serves as the source for Kubernetes deployments
- Integrated with Helm for release management

**Helm Release Management**
- Manages application releases and deployments
- Provides versioning and rollback capabilities
- Configures Kubernetes resources declaratively

#### Observability Stack

**Prometheus**
- Collects metrics from Redis Master and Redis Replica
- Monitors system health and performance
- Stores time-series data for analysis

**Grafana (Optional)**
- Visualizes metrics and system dashboards
- Provides real-time monitoring views
- Customizable dashboards for different stakeholders

**Alertmanager**
- Receives alerts from Redis components
- Routes alerts to appropriate channels
- Manages alert grouping and silencing

---

## Key Design Principles

1. **Microservices Architecture** - Services are independently deployable and scalable
2. **API Gateway Pattern** - Single entry point for external traffic with centralized concerns
3. **Service Mesh Ready** - gRPC communication enables future service mesh integration
4. **Cloud-Native** - Built for Kubernetes with cloud-native patterns
5. **Observability First** - Comprehensive monitoring and alerting from the ground up
6. **CI/CD Automation** - Fully automated deployment pipeline
7. **High Availability** - Redis StatefulSet with master-replica configuration
8. **Strong Consistency** - Persistent volumes ensure data durability

---

## Communication Protocols

- **External**: HTTP REST for client-to-service communication
- **Internal**: gRPC/RPC for inter-service communication
- **Data**: Direct database connections (MySQL) and Redis protocol

---

## Deployment

The system is deployed using:
- **Kubernetes** for container orchestration
- **Helm** for package management and deployment
- **GitHub Actions** for CI/CD automation
- **Container Registry** for image storage and distribution

All deployment configurations are managed in the `deploy/helm/` directory.
