# Authentication & Authorization Guide for Open Data Exchange Sandbox

This guide details the authentication and authorization mechanisms implemented in the Open Data Exchange Sandbox, leveraging Asgardeo as the central Identity Provider, WSO2 Choreo for deployment, and GraphQL with JWTs for secure data access.

#### 1\. Introduction

Secure identity verification and access control are paramount for a digital data exchange. This sandbox implements a robust framework to authenticate various entity types (Citizens, Businesses, Government Employees) and authorize their access to sensitive data based on their roles and permissions.

#### 2\. Core Components Involved

  * **Asgardeo:** The Identity Provider (IdP) and Authorization Server. It manages user accounts, handles login flows, and issues secure JSON Web Tokens (JWTs).
  * **WSO2 Choreo:** The platform where all backend services (GraphQL Supergraph, Provider Services, Mock Services) are deployed. Choreo integrates seamlessly with Asgardeo for API security.
  * **Consumer Web Portal (`apps/portal/`):** The client-facing application where users initiate authentication.
  * **GraphQL Supergraph Gateway (`apps/graphql-supergraph/`):** The primary API endpoint that receives requests, validates JWTs, and enforces initial authorization policies.
  * **Provider Services (`services/dmt-provider-service/`, `services/rop-provider-service/`):** Backend services that serve specific data domains and perform fine-grained authorization.
  * **Mock SLUDI Service (`mocks/mock-sludi/`):** A simulated service for Citizen identity verification.
  * **Mock ROC Service (`mocks/mock-roc/`):** A simulated service for Business registration verification.

#### 3\. Authentication Flow Overview

All authentication flows leverage OAuth 2.0 / OpenID Connect (OIDC) via Asgardeo, which then issues JWTs.

1.  **User Initiates Action:** A user (Citizen, Business, or Government Employee) accesses the `Consumer Web Portal` to use a service.
2.  **Authentication Delegation:** The `Portal` redirects the user's browser to `Asgardeo` for authentication (using the OAuth 2.0 Authorization Code Flow).
3.  **Entity-Specific Verification (Orchestrated by Asgardeo):**
      * **For Citizens:** `Asgardeo` redirects to the `Mock SLUDI Service`. The user provides simulated SLUDI credentials (e.g., NIC, mock biometrics). `Mock SLUDI Service` processes these and returns a mock verification result to `Asgardeo`.
      * **For Businesses:** `Asgardeo` initiates a multi-factor authentication process. First, it verifies the company's registration details against the `Mock ROC Service`. Second, the individual user's credentials (username/password) are authenticated, and their association/authority within that company is verified (e.g., through user roles managed in Asgardeo, potentially informed by mock ROC data).
      * **For Government Employees:** `Asgardeo` directly handles their authentication (username/password, potentially MFA configured in Asgardeo).
4.  **JWT Issuance:** Upon successful verification, `Asgardeo` issues **JSON Web Tokens (JWTs)** – an ID Token (containing identity claims) and an Access Token (for accessing protected resources) – back to the `Portal`. A Refresh Token may also be issued.

#### 4\. Authorization Flow Overview

Authorization occurs at multiple layers, primarily driven by the claims within the JWTs issued by Asgardeo.

1.  **API Request with JWT:** The `Consumer Web Portal` (or a server-side application acting on behalf of an authenticated user) includes the Access Token (JWT) in the `Authorization: Bearer <JWT>` header of all GraphQL API requests to the `GraphQL Supergraph Gateway`.

2.  **GraphQL Supergraph Gateway (`apps/graphql-supergraph/`) - Initial Authorization:**

      * **JWT Validation:** The `GraphQL Supergraph Gateway` (a Ballerina service) is the first component to receive the request. It rigorously validates the incoming JWT:
          * **Signature Verification:** Ensures the token hasn't been tampered with, using Asgardeo's public keys.
          * **Expiration Check:** Verifies the token is still valid.
          * **Issuer and Audience Validation:** Confirms the token was issued by Asgardeo and is intended for this API.
      * **Claim Extraction & Central Policies:** If the JWT is valid, the Gateway extracts crucial claims (e.g., `user_id`, `roles`, `scopes` like `citizen.read`, `business.write`). It then applies **centralized authorization policies** based on these claims (e.g., "only users with a `provider` role can access mutation operations").
      * **Context Propagation:** The extracted and validated identity and authorization information (claims, roles, scopes) are propagated into the GraphQL context object, which is then passed to the downstream subgraphs.

3.  **Provider Services (GraphQL Subgraphs) - Fine-Grained Authorization:**

      * **Resolver-Level Checks:** Each `Provider Service` (e.g., `DMT_Provider`, `ROP_Provider` in `services/`) receives the request from the `GraphQL Supergraph Gateway` with the user's identity and permissions in the GraphQL context.
      * Resolvers within these Ballerina services perform **fine-grained, resource-level authorization checks**. For example:
          * A resolver for `citizen.vehicleRegistration` might check if the user has the `police` role *and* the `dmt:vehicle:read` scope before returning data.
          * A mutation for `business.updateDetails` would verify that the authenticated user represents the specific business being updated, and has the `business:details:write` scope.
      * **Shared Authorization Logic:** The `common/lib-auth/` Ballerina module houses reusable functions for JWT claim parsing, scope validation, and role checks, ensuring consistent authorization logic across all subgraphs.

#### 5\. Entity-Specific Permissions

Permissions are managed in Asgardeo and reflected in the JWT `scopes` and `roles` claims.

  * **Citizen (Consumer Only):**

      * **Scopes:** Primarily `citizen.read` (limited fields), allowing access to public or their own personal data.
      * **Authorization:** Access to read public data, or specific personal records (e.g., their own vehicle registration, person details) verified against their authenticated identity. Cannot perform write operations to core data.

  * **Business (Consumer & Provider):**

      * **Scopes:** `business.read`, `business.write` (for their own business data), `citizen.read` (limited, for related services like employee verification), `business.manage`.
      * **Authorization (Consumer):** Can read general business data and specific `citizen.read` data related to their business operations.
      * **Authorization (Provider):** Can create, update, and manage data pertaining to their own business entity (e.g., updating company profile, potentially registering new employees via a provided service endpoint).

  * **Government (Consumer & Provider):**

      * **Scopes:** `gov.read`, `gov.write`, `gov.manage`, `citizen.read`, `citizen.create`, `citizen.update`, `business.read`, `business.create`, `business.update`, `audit.read`.
      * **Authorization (Consumer):** Broad read access to citizen, business, and other government data, typically segmented by the agency's specific responsibilities (e.g., Police access certain ROP/DMT data for law enforcement, but not all citizen health records).
      * **Authorization (Provider):** Can perform creation, update, and management operations on relevant data models as per their agency's role and data ownership (e.g., DMT updating vehicle registrations, ROP managing person records, Police updating incident reports).

#### 6\. Service-to-Service Authentication

  * For internal communication between provider services (e.g., `DMT_Provider` querying `ROP_Provider`), WSO2 Choreo's managed environment ensures secure communication using internal network policies (e.g., Mutual TLS where applicable).
  * Requests are routed via Choreo's internal service discovery, and fine-grained authorization policies can still be applied at the receiving subgraph based on the identity of the calling service (e.g., if `DMT_Provider` is authorized to access `ROP_Provider`'s `vehicleOwner` data).