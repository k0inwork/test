# External Integrations Guide

The system relies on several external services for identity, monitoring, and infrastructure management.

## 1. Authentication (LDAP & Kerberos)

### 1.1. LDAP
- **Implementation:** Uses `django-auth-ldap`.
- **Configuration:** Two LDAP backends are configured in `settings.py` (`LDAPBackend1`, `LDAPBackend2`) for redundancy.
- **Group Mapping:** LDAP groups (e.g., `tsumadm`, `trafadm`) are mapped to Django permissions. This allows for fine-grained access control based on organizational roles.
- **Search:** Configured to search for users in `ou=users,dc=strela` and groups in `ou=groups,dc=strela`.

### 1.2. Kerberos
- **Usage:** Referenced in `variables.py` (`KERBEROS_USER`). Typically used for Single Sign-On (SSO) or secure service-to-service communication.

## 2. Monitoring (Zabbix)

- **Integration:** The system interacts with Zabbix to monitor the health of managed devices.
- **Naming Convention:** Products (PDUs) generate their Zabbix names using a prefix defined in `GlobalSettings`.
- **Alerting:** The system can display Zabbix maps on its dashboard and reacts to problem severity levels.

## 3. Inventory Management (GLPI)

- **Role:** Source of truth for physical infrastructure.
- **Integration:** Synchronized via a custom library (mentioned in `requirements.txt`) and periodic Celery tasks.
- **Data Points:** Syncs datacenter names, rack positions, and device status.

## 4. Virtualization (QEMU)

- **Usage:** Used for hosting virtual gateways or testing environments.
- **Network Control:** Custom scripts (`qemu-ifup-brcontrol`) manage bridge interfaces for QEMU VMs, allowing them to participate in the VXLAN network managed by the ARMs.

## 5. Logging (Graylog)

- **Role:** Centralized log aggregation.
- **Implementation:** Django is configured to send logs to a Rotating File Handler, which is typically picked up by a log collector and forwarded to Graylog (as configured in `GlobalSettings`).
