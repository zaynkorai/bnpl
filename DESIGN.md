## BNPL System Design: 10 Million Transactions Per Day

Handling 10 million transactions per day, with 10% month-on-month growth, At an average of **approximately 115 transactions per second**, with significant peak loads, the architecture must be highly distributed, resilient, and optimized for high throughput and low latency.

### High Level System Architecture

Adopting an event-driven microservices architecture with a clear separation of concerns.

    graph TD
        UserApp[Consumer Mobile App / Web] --> |API Calls| CDN(CDN / Edge Network)
        MerchantEcom[Merchant E-commerce Site] --> |API Calls & Widgets| CDN
        CDN --> |Load Balanced Traffic| API_Gateway(API Gateway / Load Balancer Cluster)

        API_Gateway --> |gRPC/HTTP/Async| UserAuthSvc(User Auth Service)
        API_Gateway --> |gRPC/HTTP/Async| OnboardingSvc(Onboarding Service)
        API_Gateway --> |gRPC/HTTP/Async| CreditDecisionSvc(Credit Decisioning Service)
        API_Gateway --> |gRPC/HTTP/Async| TransactionSvc(Transaction Service)
        API_Gateway --> |gRPC/HTTP/Async| PaymentSvc(Payment Service)
        API_Gateway --> |gRPC/HTTP/Async| InstallmentSvc(Installment Service)
        API_Gateway --> |gRPC/HTTP/Async| NotificationSvc(Notification Service)
        API_Gateway --> |gRPC/HTTP/Async| MerchantSvc(Merchant Service)
        API_Gateway --> |gRPC/HTTP/Async| ReportingSvc(Reporting Service)

        UserAuthSvc -- DB --> UserDB(User Database)
        OnboardingSvc -- DB --> UserDB
        OnboardingSvc --> |Integrate with| IDVerification(3rd Party ID Verification)
        CreditDecisionSvc --> |Integrate with| CreditBureau(3rd Party Credit Bureaus)
        CreditDecisionSvc --> |Internal rules/ML| RiskDB(Risk & ML Feature Store - DynamoDB / Redis)
        TransactionSvc -- DB --> TransactionDB(Transaction Database - Sharded PostgreSQL / DynamoDB)
        TransactionSvc --> |Pub/Sub| EventsQueue(Events Queue - Kafka Cluster)
        PaymentSvc --> |Pub/Sub| EventsQueue
        PaymentSvc --> |Integrate with| PaymentGateway(3rd Party Payment Gateways)
        InstallmentSvc --> |Pub/Sub| EventsQueue
        InstallmentSvc -- DB --> InstallmentDB(Installment Database - Sharded PostgreSQL / DynamoDB)
        NotificationSvc --> |SMS/Email/Push| NotificationProviders(3rd Party Notification Services)
        MerchantSvc -- DB --> MerchantDB(Merchant Database - PostgreSQL)
    
        EventsQueue --> CreditDecisionSvc
        EventsQueue --> InstallmentSvc
        EventsQueue --> NotificationSvc
        EventsQueue --> DataLake(Data Lake / Analytics)
        EventsQueue --> AuditLogSvc(Audit Log Service)
        EventsQueue --> ReportingSvc
        EventsQueue --> MonitoringMetrics(Monitoring & Metrics Pipelines)

        AuditLogSvc -- DB --> AuditDB(Audit Database - DynamoDB / S3 with Athena)
        ReportingSvc -- DB --> ReportingDB(Reporting Data Store - Read Replicas / Data Marts)

### Key Components

  * **CDN / Edge Network (New Consideration):** To reduce latency, offload API Gateway, improves performance for static and cacheable content.
  * **API Gateway (NGINX, or AWS API Gateway):** The entry point. Clustered setup with rate limiting, auth, and advanced routing. Heavily leveraging features like rate limiting, circuit breaking, and caching for API endpoints that can benefit.
  * **User Authentication Service:** Managing user registration, login, session management, and KYC checks. 
  * **Onboarding Service:** Handles merchant onboarding, verification, and configuration. 
  * **Credit Decisioning Service:**  Real-time approvals. Uses a Feature Store and pre-computed ML features. Optimized for low latency and minimal reliance on external bureaus.
  * **Transaction Service:** Records all purchase requests and their initial status (pending, approved, rejected). It will publish events to the `EventsQueue` for subsequent processing by other services.
  * **Payment Service:** Orchestrates payments, integrates with payment gateways, processes down payments, and handles failed payment attempts and retries. This service needs sophisticated retry mechanisms with exponential backoff and idempotent operations to ensure financial consistency at high volumes. It interacts heavily with external, potentially slower payment gateways.
  * **Installment Service:** Managing the lifecycle of installment plans: schedules payments, tracks balances, applies late fees, processes prepayments, and marks payments as complete. This service will handle a large volume of scheduled events.
  * **Notification Service:** Sends transactional notifications (payment reminders, approval/rejection, late payment alerts) via SMS, email, or push. Scalability here means efficient queuing and dispatching to third-party providers.
  * **Merchant Service:** Managing merchant profiles, configurations, transaction reporting for merchants, and settlement. Will also need to provide highly scalable APIs for merchant integration, potentially leveraging webhooks for real-time updates.
  * **Reporting Service:** Real-time aggregation and reporting using a separate optimized store. This service would consume events from Kafka, process them, and store aggregated data in an optimized Reporting Data Store.
  * **Events Queue (Apache Kafka Cluster):** Handles async workflows and decouples services. Requires a multi-broker, partitioned setup.
  * **Databases** 
      * **PostgreSQL (Sharded / AWS Aurora with multiple read replicas):** For core relational data (users, merchants, configuration settings, smaller lookup tables every where ACID consistency important).
      * **Redis (Cluster):** Caching frequently accessed data (e.g., user profiles, merchant configurations, credit decisioning features), real-time session management, and high-speed rate limiting. 
      * **DynamoDB (or other managed NoSQL like Cassandra/ScyllaDB):**  Append-only, high-volume data like transaction logs or audit trails.
      
  * **Audit Log Service:** Centralized logging of all critical actions and transactions for compliance and debugging. storing in DynamoDB or S3.

### Technology Stack

  1. **Backend:** Golang for performance and concurrency.
      * **Frameworks** Go-kit (for service structure), gRPC (for high-performance inter-service communication and potential client-side load balancing), HTTP/REST for external APIs, Gorilla Mux/Echo (for HTTP routing).
      * **Clients/Libraries** Database drivers optimized for connection pooling (e.g., `pgx` for PostgreSQL, official AWS SDK for DynamoDB), Sarama (Kafka client), Redis client.
  2. **Databases:**
      * **PostgreSQL:**
      * **Redis/AWS ElastiCaches** for high availability and distributed caching.
      * **DynamoDB** For high-throughput, flexible schema data, especially for transaction logs and ephemeral data.
  3. **Message Broker:** (Apache Kafka Cluster) With careful topic partitioning and replication strategy.
  4. **Cloud Provider:** AWS. All listed services (managed Kubernetes EKS, managed databases like Aurora/DynamoDB/ElastiCache, Kafka MSK) simplify deployment, scaling, and operations at this scale.
  5. **Containerization:** Docker.
  6. **Orchestration:** Kubernetes (AWS EKS) with advanced autoscaling strategies:
      * **Horizontal Pod Autoscaler (HPA):** Based on CPU utilization or custom metrics (e.g., Kafka consumer lag).
      * **Cluster Autoscaler:** To automatically adjust the number of nodes in the cluster based on HPA demands.
      * **Vertical Pod Autoscaler (VPA):** For optimizing resource requests/limits for individual pods (though often used with HPA cautiously).
      * **Kubernetes Event-driven Autoscaling:** To scale Kafka consumers based on the number of messages in topics, ensuring efficient processing of asynchronous workloads.
