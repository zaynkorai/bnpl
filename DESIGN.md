## High Level System Architecture

Adopting an event-driven microservices architecture with a clear separation of concerns.

    graph TD
        UserApp[Consumer Mobile App / Web] --> |API Calls| API_Gateway(API Gateway / Load Balancer)
        MerchantEcom[Merchant E-commerce Site] --> |API Calls & Widgets| API_Gateway
        API_Gateway --> |gRPC/HTTP| UserAuthSvc(User Auth Service)
        API_Gateway --> |gRPC/HTTP| OnboardingSvc(Onboarding Service)
        API_Gateway --> |gRPC/HTTP| CreditDecisionSvc(Credit Decisioning Service)
        API_Gateway --> |gRPC/HTTP| TransactionSvc(Transaction Service)
        API_Gateway --> |gRPC/HTTP| PaymentSvc(Payment Service)
        API_Gateway --> |gRPC/HTTP| InstallmentSvc(Installment Service)
        API_Gateway --> |gRPC/HTTP| NotificationSvc(Notification Service)
        API_Gateway --> |gRPC/HTTP| MerchantSvc(Merchant Service)

        UserAuthSvc -- DB --> UserDB(User Database)
        OnboardingSvc -- DB --> UserDB
        OnboardingSvc --> |Integrate with| IDVerification(3rd Party ID Verification)
        CreditDecisionSvc --> |Integrate with| CreditBureau(3rd Party Credit Bureaus)
        CreditDecisionSvc --> |Internal rules/ML| RiskDB(Risk & ML Model Database)
        TransactionSvc -- DB --> TransactionDB(Transaction Database)
        TransactionSvc --> |Pub/Sub| EventsQueue(Events Queue - Kafka/RabbitMQ)
        PaymentSvc --> |Pub/Sub| EventsQueue
        PaymentSvc --> |Integrate with| PaymentGateway(3rd Party Payment Gateways)
        InstallmentSvc --> |Pub/Sub| EventsQueue
        InstallmentSvc -- DB --> InstallmentDB(Installment Database)
        NotificationSvc --> |SMS/Email/Push| NotificationProviders(3rd Party Notification Services)
        MerchantSvc -- DB --> MerchantDB(Merchant Database)
        EventsQueue --> CreditDecisionSvc
        EventsQueue --> InstallmentSvc
        EventsQueue --> NotificationSvc
        EventsQueue --> DataLake(Data Lake / Analytics)
        EventsQueue --> AuditLogSvc(Audit Log Service)

        AuditLogSvc -- DB --> AuditDB(Audit Database)


### Key Components

**API Gateway (NGINX, or AWS API Gateway)** 

Entry point for all external requests. Handles authentication, rate limiting, and routes requests to the appropriate microservices.

**User Authentication Service** 

Manages user registration, login, session management, and KYC checks.

**Onboarding Service**

Handles merchant onboarding, verification, and configuration.

**Credit Decisioning Service** 

Brain ofsystem. Takes consumer data, queries credit bureaus, applies internal rules, and runs ML models to approve/decline applications and set dynamic spending limits. Needs to be extremely fast.

**Transaction Service** 

Records all purchase requests and their initial status (pending, approved, rejected).

**Payment Service** 

Orchestrates payments, integrates with payment gateways, processes down payments, and handles failed payment attempts and retries.

**Installment Service** 

Manages the lifecycle of installment plans: schedules payments, tracks balances, applies late fees, processes prepayments, and marks payments as complete.

**Notification Service** 

Sends transactional notifications (payment reminders, approval/rejection, late payment alerts) via SMS, email, or push.

**Merchant Service** 

Manages merchant profiles, configurations, transaction reporting for merchants, and settlement.

**Events Queue (Kafka)** 

For asynchronous communication between services, enabling high throughput and decoupling.

**Databases** 

PostgreSQL for relational data, 
DynamoDB for high-throughput transaction logs.

**Audit Log Service**

Centralized logging of all critical actions and transactions for compliance and debugging.

### Technology Stack 

**Backend** 
Go (Golang) for all microservices. Its concurrency model (goroutines, channels) and performance are ideal for high-throughput systems.

**Frameworks/Libraries**
Go-kit (for service structure), g
RPC (for inter-service communication), 
HTTP/REST for external APIs, 
Gorilla Mux/Echo (for HTTP routing), 
GORM (for database interactions), 
Sarama (Kafka client), Redis client.

**PostgreSQL** For core relational data (users, merchants, installment plans, audit logs). Provides strong ACID properties.

**Redis** For caching, session management, rate limiting, and potentially temporary storage for real-time credit decisioning.

**DynamoDB** For append-only data like raw transaction logs, if PostgreSQL becomes a bottleneck for all transaction data.

**Message Broker** Apache Kafka

**Cloud Provider** AWS. All offer managed services that simplify deployment, scaling, and operations (managed Kubernetes, managed databases).

**Monitoring & Logging** Prometheus/Grafana, ELK Stack

**Containerization** Docker

**Orchestration** Kubernetes