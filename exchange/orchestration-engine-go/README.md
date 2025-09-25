# Introduction

This directory contains the source code for the Orchestration Engine, a Go-based service that manages
and orchestrates various tasks and workflows within the Exchange platform. The Orchestration Engine is
responsible for coordinating complex operations, ensuring that tasks are executed in the correct order,
and handling dependencies between different components.

## How it works

The Orchestration Engine (OE) is a Go-based service that orchestrates data requests from consumers to multiple data
providers. It handles authorization and consent checks, argument mapping, and data aggregation.

### Key Features
- **GraphQL API**: The OE exposes a GraphQL API for consumers to request data.
- **Multiple Data Providers**: It can fetch data from multiple providers based on the consumer's request.
- **Authorization Checks**: Before fetching data, the OE checks with the Policy Decision Point (PDP) to ensure the
  consumer is authorized to access the requested fields.
- **Consent Management**: The OE interacts with the Consent Engine (CE) to verify that the consumer has the necessary consents for
  accessing certain data fields.

## Setting Up the Development Environment

To set up the development environment for the Orchestration Engine, follow these steps:

1. **Install Go**: Ensure you have Go installed on your machine. You can download it from the
   official [Go website](https://golang.org/dl/).
2. **GraphQL Specification**: The Orchestration Engine uses GraphQL for its API. Familiarize yourself with the GraphQL
   specification by visiting the [GraphQL official site](https://graphql.org/).
3. **`schema.graphql` Schema File**: The GraphQL schema file is currently located in the `schemas` directory. These
   files define the structure of the API and the types of data that can be queried.
   We have placed the sample schema in it.
    - It should include `@sourceInfo` the directives in each of its leaf fields along with the following fields.
        - `providerKey` - A unique identifier for the data provider.
        - `providerField` - The field name in the provider's schema that corresponds to this field.
4. **`config.json` File**: Refer to the sample `config.example.json` file
   and create your own `config.json` file based on it. This file lists out the following information.
    - `pdpUrl` - The URL of the Policy Decision Point which handles authorization.
    - `ceUrl` - The URL of the Consent Engine which handles consent management.
    - `providers` - An array of data providers, each with a `providerKey` and `providerUrl`. 
      For detailed provider integration steps, see the [Provider Onboarding Guide](PROVIDER_CONFIGURATION.md).

5. **Run the Server**: You can run the Orchestration Engine server using the following command:
   ```bash
   go run main.go
   ```
   The server will start and listen for incoming requests.