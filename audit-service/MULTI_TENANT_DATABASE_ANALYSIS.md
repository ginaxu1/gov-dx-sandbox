# Database Choice for Reusable, Configurable Audit Service

## New Requirement: Company-Wide Reusable Service

The audit service must be **reusable and configurable by different teams** in the company, not just for OpenDIF. Each team deploys and configures their own instance for their specific audit logging needs. This is **not a multi-tenant shared service**, but rather a **generalized, reusable service** that teams can adopt and configure independently.

## Key Requirements for Reusable Audit Service

### 1. **Reusability & Generalization**
- Service can be deployed by any team for their own audit logging
- Teams configure the service for their specific needs
- No shared infrastructure or data isolation concerns
- Each team manages their own instance

### 2. **Ease of Adoption**
- Teams should integrate via simple API calls
- Minimal configuration required
- No deep database knowledge needed
- JSON-based requests (natural for REST APIs)
- Easy to deploy and run

### 3. **Operational Simplicity**
- Teams can manage their own database infrastructure (or use managed services)
- Simple deployment and maintenance
- Easy to scale for individual team needs
- Simple backup/restore procedures
- Low operational overhead

### 4. **Schema Flexibility**
- Different teams may log different event types
- Teams may have different metadata requirements
- Should accommodate schema evolution without migrations
- Support for custom fields per team's use case
- General enough to handle various audit scenarios

### 5. **Performance**
- Handle individual team's workload efficiently
- Scaling capability (vertical or horizontal) as needed
- Efficient querying for audit log retrieval
- Good performance for compliance queries

### 6. **Compliance (Still Important)**
- Audit logs must be tamper-resistant
- Configurable data retention policies
- Query capabilities for compliance audits
- Data integrity guarantees

## Database Comparison for Reusable Audit Service

### PostgreSQL (with TimescaleDB)

#### ✅ Advantages
- **ACID guarantees**: Strong data integrity for compliance
- **SQL familiarity**: Most teams understand SQL for custom queries
- **Mature ecosystem**: Well-understood by operations teams
- **Rich querying**: Complex analytics and reporting
- **JSONB support**: Flexible metadata storage
- **TimescaleDB**: Excellent for time-series audit data
- **Managed services**: AWS RDS, Google Cloud SQL reduce ops burden

#### ❌ Disadvantages
- **Schema rigidity**: Adding new fields requires migrations
- **Operational complexity**: Requires DBAs, connection pooling, backups
- **Vertical scaling limits**: May need read replicas, sharding
- **Integration overhead**: Teams need to understand SQL concepts
- **Configuration complexity**: Schema changes affect all data

#### Implementation Considerations
```sql
-- Standard audit logs table (no tenant_id needed)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    trace_id UUID,
    event_name VARCHAR(100) NOT NULL,
    event_type VARCHAR(20),
    status VARCHAR(10) NOT NULL CHECK (status IN ('SUCCESS', 'FAILURE')),
    actor_type VARCHAR(10) NOT NULL CHECK (actor_type IN ('USER', 'SERVICE')),
    actor_service_name VARCHAR(100),
    actor_user_id UUID,
    actor_user_type VARCHAR(20),
    actor_metadata JSONB,
    target_type VARCHAR(10) NOT NULL CHECK (target_type IN ('RESOURCE', 'SERVICE')),
    target_service_name VARCHAR(100),
    target_resource VARCHAR(100),
    target_resource_id UUID,
    target_metadata JSONB,
    requested_data JSONB,
    response_metadata JSONB,
    event_metadata JSONB
);

-- Standard indexes for performance
CREATE INDEX idx_audit_logs_trace_id ON audit_logs(trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_event_name ON audit_logs(event_name);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_actor_service ON audit_logs(actor_service_name) WHERE actor_type = 'SERVICE';
CREATE INDEX idx_audit_logs_actor_user_id ON audit_logs(actor_user_id) WHERE actor_type = 'USER';
```

**Benefits:**
- Simpler schema (no tenant isolation needed)
- Standard indexes work well
- Query performance is straightforward
- Schema migrations are simpler (single instance)
- Teams can customize indexes based on their queries

---

### MongoDB

#### ✅ Advantages
- **Schema flexibility**: Teams can have different document structures for their use case
- **Horizontal scaling**: Shard easily if needed for large teams
- **JSON-native**: Natural fit for REST API integration
- **Operational simplicity**: Managed services (Atlas) reduce ops burden
- **Easy adoption**: Teams just send JSON, no SQL knowledge needed
- **Dynamic schema**: New event types don't require migrations
- **Document model**: Matches API request/response structure
- **Configuration flexibility**: Teams can customize indexes and retention per their needs

#### ❌ Disadvantages
- **Eventual consistency**: Default read concern may not be strict enough
- **ACID limitations**: Multi-document transactions are complex
- **Query limitations**: Complex joins/aggregations harder than SQL
- **Compliance concerns**: Eventual consistency may not meet strict requirements

#### Implementation Considerations
```javascript
// Standard audit logs collection
{
  _id: ObjectId("..."),
  id: "uuid-123",          // Application-level UUID
  timestamp: ISODate("2024-01-20T10:00:00Z"),
  trace_id: "abc-123",
  event_name: "POLICY_CHECK",
  event_type: "READ",
  status: "SUCCESS",
  actor_type: "SERVICE",
  actor_service_name: "ORCHESTRATION_ENGINE",
  // ... rest of fields
}

// Core indexes for audit queries
db.audit_logs.createIndex({ trace_id: 1, timestamp: 1 });
db.audit_logs.createIndex({ timestamp: -1 });
db.audit_logs.createIndex({ event_name: 1 });
db.audit_logs.createIndex({ status: 1 });
db.audit_logs.createIndex({ actor_service_name: 1 });
db.audit_logs.createIndex({ target_service_name: 1 });

// Teams can add custom indexes based on their query patterns
// Example: db.audit_logs.createIndex({ custom_field: 1 });
```

**Benefits:**
- Simple document structure
- Flexible schema evolution
- Teams can add custom fields without coordination
- Easy to configure per team's needs

---

### Hybrid Approach: PostgreSQL + MongoDB

#### Architecture
- **PostgreSQL**: Store critical compliance data (id, timestamp, event_name, status)
- **MongoDB**: Store flexible metadata (requested_data, response_metadata, event_metadata)
- **Reference**: PostgreSQL row has `metadata_id` pointing to MongoDB document

#### ✅ Advantages
- Best of both worlds: ACID for compliance, flexibility for metadata
- Teams get structured queries on core fields, flexible JSON for metadata
- Can migrate teams gradually

#### ❌ Disadvantages
- **Complexity**: Two databases to manage
- **Consistency**: Need to keep both in sync
- **Operational overhead**: Double the infrastructure
- **Query complexity**: May need to join across databases

---

## Recommendation: **MongoDB** (with Strong Read Concern)

### Why MongoDB for Reusable Audit Service?

1. **API-First Architecture**
   - Teams interact via REST API, not directly with database
   - JSON requests map naturally to MongoDB documents
   - No SQL knowledge required for teams
   - Simple integration for any team

2. **Schema Flexibility**
   - Different teams can log different event structures for their use cases
   - New event types don't require schema migrations
   - Teams can extend metadata to fit their specific needs
   - General model accommodates various audit scenarios

3. **Operational Simplicity**
   - Use MongoDB Atlas (managed service) for easy deployment
   - Automatic scaling, backups, monitoring
   - Teams can manage their own instance or use managed service
   - Low operational overhead

4. **Configuration Flexibility**
   - Teams can customize indexes based on their query patterns
   - Configurable retention policies per team's compliance needs
   - Teams can optimize for their specific workload
   - No one-size-fits-all constraints

5. **Performance**
   - Efficient indexes for common audit queries
   - Horizontal scaling available if teams need it
   - Good performance for trace correlation and time-range queries
   - Teams can tune performance for their workload

6. **Compliance Mitigation**
   - Use `readConcern: "majority"` for strong consistency
   - Use `writeConcern: { w: "majority" }` for durability
   - Document versioning for audit trail
   - Encryption at rest and in transit
   - Teams can configure retention policies

### Standard Model (No Tenant Isolation Needed)

```javascript
{
  _id: ObjectId("..."),
  id: "uuid-123",          // Application-level UUID
  timestamp: ISODate("2024-01-20T10:00:00Z"),
  trace_id: "abc-123",
  event_name: "POLICY_CHECK",
  event_type: "READ",
  status: "SUCCESS",
  actor_type: "SERVICE",
  actor_service_name: "ORCHESTRATION_ENGINE",
  actor_user_id: null,
  actor_user_type: null,
  actor_metadata: {},
  target_type: "SERVICE",
  target_service_name: "POLICY_DECISION_POINT",
  target_resource: null,
  target_resource_id: null,
  target_metadata: {},
  requested_data: {
    "fields": ["name", "email"],
    "applicationId": "app-123"
  },
  response_metadata: {
    "authorized": true,
    "consentRequired": false
  },
  event_metadata: {}
}
```

### Standard Indexes (Teams Can Customize)

```javascript
// Core indexes for audit queries
db.audit_logs.createIndex({ trace_id: 1, timestamp: 1 });
db.audit_logs.createIndex({ timestamp: -1 });
db.audit_logs.createIndex({ event_name: 1 });
db.audit_logs.createIndex({ status: 1 });
db.audit_logs.createIndex({ actor_service_name: 1 });
db.audit_logs.createIndex({ target_service_name: 1 });

// Teams can add custom indexes based on their query patterns
// Example: db.audit_logs.createIndex({ custom_field: 1 });
```

### Write Concern for Compliance

```go
// Strong write concern for compliance
opts := options.InsertOne().SetWriteConcern(
    writeconcern.New(writeconcern.WMajority(), writeconcern.J(true)),
)
result, err := collection.InsertOne(ctx, auditLog, opts)
```

### Read Concern for Consistency

```go
// Strong read concern
opts := options.Find().SetReadConcern(readconcern.Majority())
cursor, err := collection.Find(ctx, filter, opts)
```

---

## Alternative: Managed PostgreSQL (if ACID is Strict Requirement)

If compliance **absolutely requires** ACID guarantees that eventual consistency cannot satisfy:

### Use **AWS RDS PostgreSQL** or **Google Cloud SQL**

#### Why Managed PostgreSQL?
- **Managed operations**: Automatic backups, patching, scaling
- **Multi-AZ**: High availability built-in
- **Read replicas**: Scale reads horizontally
- **Parameter groups**: Optimize for audit workload

#### Implementation Strategy
- **Standard table**: Single `audit_logs` table per team instance
- **Row-level security**: Optional additional isolation if needed
- **JSONB fields**: Flexible metadata storage

#### Challenges
- Schema migrations affect all data in instance
- Less flexible for different team requirements
- More operational overhead than MongoDB Atlas

---

## Final Recommendation Matrix

| Criteria | PostgreSQL | MongoDB | Winner |
|----------|-----------|---------|--------|
| **Ease of Adoption** | SQL knowledge needed | JSON API, no SQL | 🏆 MongoDB |
| **Schema Flexibility** | Requires migrations | Dynamic schema | 🏆 MongoDB |
| **Configuration Flexibility** | Schema changes affect all | Teams can customize | 🏆 MongoDB |
| **Operational Simplicity** | Requires DBAs | Managed service (Atlas) | 🏆 MongoDB |
| **Compliance (ACID)** | Full ACID | Eventual → Strong with config | 🏆 PostgreSQL* |
| **Query Flexibility** | Rich SQL | Limited joins | 🏆 PostgreSQL |
| **API Integration** | ORM/SQL mapping | Native JSON | 🏆 MongoDB |
| **Deployment Ease** | Standard SQL setup | Simple JSON config | 🏆 MongoDB |
| **Cost** | Managed RDS affordable | Atlas free tier available | 🏆 MongoDB |

*PostgreSQL wins on strict ACID, but MongoDB with `readConcern: majority` and `writeConcern: majority` provides strong enough guarantees for most compliance needs.

---

## Implementation Strategy

### Phase 1: MongoDB with Strong Consistency
1. Use MongoDB Atlas (managed service) or self-hosted
2. Configure `writeConcern: majority` and `readConcern: majority`
3. Deploy standard audit_logs collection
4. Create core indexes for common queries
5. Provide configuration options for teams

### Phase 2: Compliance Hardening (if needed)
1. Enable encryption at rest
2. Enable audit logging in MongoDB
3. Implement document versioning
4. Set up automated backups with retention
5. Provide configurable data retention policies

### Phase 3: Team Customization
1. Allow teams to add custom indexes based on their queries
2. Support configurable retention policies per team's compliance needs
3. Enable teams to extend metadata fields for their use cases
4. Provide documentation for team-specific optimizations

---

## Conclusion

**For a reusable, configurable audit service: MongoDB is the better choice.**

### Key Reasons:
1. **Teams integrate via API** → JSON-native MongoDB is natural fit
2. **Schema flexibility** → Different teams can configure different event structures
3. **Operational simplicity** → MongoDB Atlas reduces ops burden, easy to deploy
4. **Configuration flexibility** → Teams can customize indexes and retention per their needs
5. **Generalized model** → Accommodates various audit scenarios without complexity

### Compliance Concerns Addressed:
- Use `readConcern: majority` and `writeConcern: majority` for strong consistency
- Encryption at rest and in transit
- Document versioning for audit trail
- Configurable retention policies per team's compliance requirements

### When to Choose PostgreSQL Instead:
- **Strict regulatory requirement** for ACID transactions across all operations
- **Complex analytical queries** that require SQL joins and aggregations
- **Existing PostgreSQL expertise** in operations team
- **Teams require direct SQL access** for custom reporting
- **Teams prefer SQL** for querying audit logs

### Architecture Model:
- **Not multi-tenant**: Each team deploys their own instance
- **Reusable**: Same service code, teams configure for their needs
- **Generalized**: Flexible model accommodates various audit scenarios
- **Configurable**: Teams can customize indexes, retention, and metadata

For most use cases, **MongoDB with strong consistency settings provides the best balance of flexibility, ease of adoption, and operational simplicity for a reusable, configurable audit service that teams can deploy and configure independently.**
