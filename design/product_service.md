# Product Service Design

## 1. Overview
The Product Service manages "Physical Nodes" (previously called Products or PDUs). It tracks the location, state, and type of equipment nodes in the facility.

## 2. Data Model (SQLite)

### Node Table (Product)
| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary Key |
| name | String | e.g. "Rack 1" |
| description | String | |
| address | String | Physical address |
| state | String | Current health/status string |
| lat | Float | Latitude |
| long | Float | Longitude |
| region | String | Geographical region |
| pou_type | String | Node type (POU, OU, etc.) |
| seq_num | Integer | Sequential identifier |
| glpi_uuid | UUID | Link to external CMDB |
| operate | Boolean | Enabled/Disabled status |

## 3. API Contracts

### 3.1 REST API

#### GET `/nodes`
- **Response**: List of all nodes.

#### POST `/nodes`
- **Request**: Create a new physical node.

#### GET `/nodes/:id`
- **Response**: Single node detail.

#### POST `/nodes/:id/start` / `/nodes/:id/stop`
- **Description**: Trigger operational state changes.

### 3.2 GraphQL API

```graphql
type Node {
  id: ID!
  name: String!
  description: String
  address: String
  state: String
  lat: Float
  long: Float
  region: String
  pouType: String
  seqNum: Int
  operate: Boolean
}

type Query {
  nodes(region: String): [Node!]!
  node(id: ID!): Node
}

type Mutation {
  createNode(input: NewNode!): Node!
  updateNode(id: ID!, input: UpdateNode!): Node!
  deleteNode(id: ID!): Boolean!
}

input NewNode {
  name: String!
  region: String!
  pouType: String!
}
```

## 4. Operational Logic
- **Geo-Mapping**: Logic to parse/generate `lat;long` strings for compatibility with existing GIS systems.
- **Sync Logic**: Interface for GLPI synchronization (Mocked in Phase 1).

## 5. Mocks
- **Inventory Service**: To check if a node has attached switches.
- **Monitoring Service**: To fetch Zabbix status.
