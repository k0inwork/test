# Codebase Analysis Documentation

This set of documents provides a technical analysis of the system for developers.

## Contents

1. [Architecture Overview](architecture.md)
   - High-level components and communication flows.
2. [Data Models](data_models.md)
   - Detailed explanation of core Django models and state management.
3. [Celery and RabbitMQ](celery_rabbitmq.md)
   - Analysis of task queues, connection singletons, and the "special worker" pattern.
4. [Redis Usage](redis.md)
   - Custom storage abstractions (RedisStore, RedisVar) and infrastructure roles.
5. [Module Interaction System](modules.md)
   - Analysis of the registry-based module architecture and dynamic command mapping.
6. [Module Catalog](module_catalog.md)
   - Comprehensive guide to external service APIs (PDU, Switch, Routing, etc.).
7. [Critical Workflows](workflows.md)
   - Step-by-step breakdown of Multi-Master failover, VXLAN creation, and sync processes.
8. [External Integrations](external_integrations.md)
   - Integration details for LDAP, Zabbix, GLPI, and QEMU.
