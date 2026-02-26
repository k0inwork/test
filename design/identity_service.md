# Identity Service Design

## 1. Overview
The Identity service handles user authentication, session management, and authorization roles. It is the gatekeeper for all other services.

## 2. Data Model (SQLite)

### User Table
| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary Key |
| username | String | Unique, non-null |
| password_hash | String | BCrypt hashed password |
| role | String | Admin, Operator, Viewer |
| full_name | String | Optional display name |
| created_at | DateTime | Auto-set |
| updated_at | DateTime | Auto-update |

## 3. API Contracts

### 3.1 REST API

#### POST `/auth/login`
- **Request**: `{"username": "...", "password": "..."}`
- **Response**: `{"token": "JWT_TOKEN", "user": {...}}`

#### GET `/users`
- **Description**: List all users (Admin only).

#### POST `/users`
- **Description**: Create new user.

#### GET `/users/:id`
- **Description**: Get user details.

### 3.2 GraphQL API

```graphql
type User {
  id: ID!
  username: String!
  role: String!
  fullName: String
  createdAt: String!
}

type Query {
  me: User
  users: [User!]!
  user(id: ID!): User
}

type Mutation {
  createUser(username: String!, role: String!): User!
  updateUser(id: ID!, fullName: String): User!
  deleteUser(id: ID!): Boolean!
}
```

## 4. Authentication Mechanism
- **JWT (JSON Web Tokens)**: Issued upon login. Services will validate the JWT signature using a shared secret or public key.
- **Middleware**: A Go middleware will extract and validate the `Authorization: Bearer <token>` header.

## 5. Future Integrations
- **CAS**: Support for Central Authentication Service.
- **LDAP**: Passive sync/auth against enterprise directory.
