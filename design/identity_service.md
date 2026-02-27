# Identity Service

**Identity Service** is the authentication and authorization authority for the PUM system.

## Key Responsibilities
- User account management
- Role-Based Access Control (RBAC)
- Token issuance (JWT - planned)
- LDAP integration (Mocked)

## Security Note (Prototype)
The current implementation is a **prototype** and does not enforce full authentication for REST/GraphQL endpoints. Seeding creates default `admin` and `operator` accounts with no passwords for demonstration purposes.

## API Endpoints
- `GET /users`: List all system users.
- `POST /users`: Create a new user.
- `POST /query`: GraphQL endpoint for user queries.

## Data Model
Users are stored in `identity.db` (SQLite) with fields for `Username`, `Role`, and `CreatedAt/UpdatedAt`.
