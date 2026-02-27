# Codebase Analysis Documentation

This set of documents provides a technical analysis of the system for developers.

## Contents

1. [Architecture Overview](architecture.md)
   - High-level components and communication flows.
2. [Application Summary](apps_summary.md)
   - Concise overview of all apps and major project components.
3. [Data Models](data_models.md)
   - Detailed explanation of core Django models and state management.
4. [Celery and RabbitMQ](celery_rabbitmq.md)
   - Analysis of task queues, connection singletons, and the "special worker" pattern.
5. [Redis Usage](redis.md)
   - Custom storage abstractions (RedisStore, RedisVar) and infrastructure roles.
6. [Module Interaction System](modules.md)
   - Analysis of the registry-based module architecture and dynamic command mapping.
7. [Module Catalog](module_catalog.md)
   - Comprehensive guide to external service APIs (PDU, Switch, Routing, etc.).
8. [Critical Workflows](workflows.md)
   - Step-by-step breakdown of Multi-Master failover, VXLAN creation, and sync processes.
9. [External Integrations](external_integrations.md)
   - Integration details for LDAP, Zabbix, GLPI, and QEMU.
10. [WebSSH Terminal](webssh.md)
    - Analysis of the web terminal, session recording, and playback system.
11. [Advanced Utilities](advanced_utilities.md)
    - Deep dive into custom decorators, recursive serialization, and the alarm system.
12. [Security and Auditing](security_auditing.md)
    - Overview of activity logging, session auditing, and brute-force protection.
13. [Master/Slave Middleware](master_slave_middleware.md)
    - Analysis of distributed read-only enforcement and address discovery.
14. [Distributed Caching](distributed_caching.md)
    - Deep dive into periodic cache refresh and change detection mechanisms.
15. [Data Sync Engine](data_sync_engine.md)
    - Detailed look at object rename detection and external data mapping.
16. [JSON Dynamic API](json_middleware.md)
    - Analysis of the transparent conversion of Django views into JSON APIs.
