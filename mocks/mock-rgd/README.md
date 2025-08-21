# Mock RGD (Registrar General's Department) GraphQL Service

This service provides a GraphQL subgraph for person information including address and profession data, queryable by National Identity Card (NIC) number.

## Features

- GraphQL API with person information queries
- Mock data for testing and development
- Health check endpoint
- FastAPI-based with Strawberry GraphQL

## Setup and Installation

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Run the service:
```bash
python main.py
```

The service will start on `http://localhost:4005`

## API Endpoints

- **GraphQL Playground**: `http://localhost:4005/graphql`
- **Health Check**: `http://localhost:4005/health`
- **Service Info**: `http://localhost:4005/`

## GraphQL Schema

### Types

```graphql
type PersonInfo {
  nic: String!
  address: String!
  profession: String!
}

type Query {
  person(nic: String!): PersonInfo
  allPersons: [PersonInfo!]!
}
```

### Sample Queries

#### Get person by NIC:
```graphql
query GetPerson($nic: String!) {
  person(nic: $nic) {
    nic
    address
    profession
  }
}
```

Variables:
```json
{
  "nic": "123456789V"
}
```

#### Get all persons:
```graphql
query GetAllPersons {
  allPersons {
    nic
    address
    profession
  }
}
```

## Mock Data

The service includes sample data for the following NICs:
- `123456789V` - Software Engineer in Colombo
- `987654321V` - Medical Doctor in Kandy
- `456789123V` - Teacher in Galle
- `321654987V` - Business Analyst in Negombo

## Federation Support

This service is designed to work as a subgraph in a federated GraphQL setup, providing person address and profession data that can be composed with other services.
