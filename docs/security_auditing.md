# Security and Auditing

The PUM system is designed for high-security environments, incorporating multiple layers of activity tracking and access control.

## 1. Activity Logging

Activity tracking is implemented through several mechanisms:

### 1.1. Middleware Tracking
The `ActivityLogMiddleware` (from `django-user-activity-log2`) automatically logs all POST and GET requests to the database, capturing:
- **Actor:** The user who performed the action.
- **Action:** The URL and HTTP method.
- **Context:** Source IP address and timestamp.

### 1.2. Manual SSH Auditing
As described in the WebSSH analysis, every terminal session is linked to an `ActivityLog` entry, which points to the persistent recording of the terminal interaction.

## 2. Authentication and SSO

The system relies on external identity providers for robust authentication:
- **Dual LDAP Backends:** Provides redundancy for enterprise directory integration.
- **Kerberos Support:** Enables secure service-to-service communication and potential SSO.
- **CAS (Central Authentication Service):** Integrated via `django_mama_cas` to act as an identity provider or service.

## 3. Access Control and Brute-Force Protection

### 3.1. `django-axes`
The system uses `axes` to monitor login attempts and lock out users or IPs after multiple failed attempts (`AXES_FAILURE_LIMIT = 5`). A custom lockout view (`accounts.views.axes`) is used to provide a tailored user experience during lockouts.

### 3.2. Hardware Access Control
The `AccessControlledMixin` (used by service modules) ensures that hardware commands (like rebooting a PDU or resetting a server via IPMI) are gated by fine-grained permissions. These commands are often recorded in the `ActivityLog` to ensure that critical hardware changes are always attributable to a specific user.
