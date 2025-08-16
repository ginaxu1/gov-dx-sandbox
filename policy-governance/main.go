// main.go
package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// Database connection string (replace with actual credentials and host)
	// TODO: add environment variables for credentials for production
	connStr := "user=user password=password host=localhost port=5432 dbname=policy_db sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close() // Ensure the connection is closed when main exits

	// Ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	// Initialize the concrete database fetcher
	dbFetcher := &DatabasePolicyFetcher{DB: db}

	// Initialize the PolicyGovernanceService with the database fetcher
	// This is the line that caused the error previously.
	policyService := &PolicyGovernanceService{Fetcher: dbFetcher}

	// Register the HTTP handler, passing the initialized service
	http.HandleFunc("/evaluate-policy", HandlePolicyRequest(policyService))
	port := ":8081" // Policy Governance service typically runs on its own port
	log.Printf("Policy Governance Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

/*
To run this example with PostgreSQL:

1.  Ensure PostgreSQL is Running and Configured
    * You need a PostgreSQL server running.
    * Create a database (e.g., `policy_db`).
    * Create a user (e.g., `user` with password `password`) and grant privileges.
    * Create the `policies` table and insert sample data as specified in the previous response.
    ```sql
    CREATE TABLE policies (
        id SERIAL PRIMARY KEY,
        subgraph_name VARCHAR(255) NOT NULL,
        type_name VARCHAR(255) NOT NULL,
        field_name VARCHAR(255) NOT NULL,
        classification VARCHAR(50) NOT NULL,
        UNIQUE (subgraph_name, type_name, field_name)
    );

    INSERT INTO policies (subgraph_name, type_name, field_name, classification) VALUES
    ('dmt', 'VehicleInfo', 'engineNumber', 'ALLOW_PROVIDER_CONSENT'),
    ('drp', 'PersonData', 'photo', 'ALLOW_CITIZEN_CONSENT'),
    ('dmt', 'VehicleInfo', 'id', 'ALLOW'),
    ('dmt', 'VehicleInfo', 'make', 'ALLOW'),
    ('dmt', 'VehicleInfo', 'model', 'ALLOW'),
    ('dmt', 'VehicleInfo', 'yearOfManufacture', 'ALLOW'),
    ('dmt', 'VehicleInfo', 'ownerNic', 'ALLOW_PROVIDER_CONSENT'),
    ('dmt', 'VehicleInfo', 'conditionAndNotes', 'ALLOW_PROVIDER_CONSENT'),
    ('dmt', 'VehicleInfo', 'registrationNumber', 'ALLOW'),
    ('dmt', 'VehicleInfo', 'vehicleClass', 'ALLOW'),
    ('dmt', 'DriverLicense', 'id', 'ALLOW'),
    ('dmt', 'DriverLicense', 'licenseNumber', 'ALLOW_PROVIDER_CONSENT');
    ```

2.  Install Go PostgreSQL Driver
    If you haven't already: `go get github.com/lib/pq`

3.  Navigate to Project Root
    Open your terminal and go into the `policy-governance` directory.

4.  Run the Go Server
    `go run .` (The `.` tells Go to run the current module)

5.  Test with Curl
    ```bash
    curl -X POST \
     -H "Content-Type: application/json" \
     -d '{ "consumerId": "consumer-123", "requestedFields": [ { "subgraphName": "dmt", "typeName": "VehicleInfo", "fieldName": "engineNumber", "classification": "ALLOW", "context": {} }, { "subgraphName": "drp", "typeName": "PersonData", "fieldName": "photo", "classification": "ALLOW", "context": { "citizenId": "citizen-uuid-from-token" } } ] }' \
     http://localhost:8081/evaluate-policy
    ```
    The output should reflect the classifications retrieved from the PostgreSQL database.
*/
