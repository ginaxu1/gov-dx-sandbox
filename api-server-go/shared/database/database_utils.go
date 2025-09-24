package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/pkg/errors"
)

// ValidateDBConnection checks if the database connection is valid
func ValidateDBConnection(db *sql.DB) error {
	if db == nil {
		return errors.InternalError("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return errors.HandleDatabaseError(err, "connection validation")
	}
	return nil
}

// ValidateDBConnectionSimple checks if the database connection is valid (simple version)
func ValidateDBConnectionSimple(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
