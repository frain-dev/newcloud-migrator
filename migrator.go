package main

import (
	"fmt"
	"github.com/frain-dev/convoy/datastore"
	"github.com/jmoiron/sqlx"
)

type Migrator struct {
	OldBaseURL     string
	OldPostgresDSN string
	NewPostgresDSN string
	PAT            string
	MigrateEvents  bool

	user     *datastore.User
	userOrgs []datastore.Organisation
	projects []datastore.Project

	newDB *sqlx.DB
	oldDB *sqlx.DB
}

func NewMigrator(oldBaseURL string, oldPostgresDSN string, newPostgresDSN string, PAT string, migrateEvents bool) (*Migrator, error) {
	oldDB, err := sqlx.Connect("postgres", oldPostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open oldPostgresDSN: %v", err)
	}

	newDB, err := sqlx.Connect("postgres", newPostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open newPostgresDSN: %v", err)
	}

	return &Migrator{
		OldBaseURL:     oldBaseURL,
		OldPostgresDSN: oldPostgresDSN,
		NewPostgresDSN: newPostgresDSN,
		PAT:            PAT,
		MigrateEvents:  migrateEvents,

		oldDB: oldDB,
		newDB: newDB,
	}, nil
}

func (m *Migrator) Run() error {
	err := m.RunUserMigration()
	if err != nil {
		return fmt.Errorf("failed to run user migration: %v", err)
	}

	err = m.RunOrgMigration()
	if err != nil {
		return fmt.Errorf("failed to run org migration: %v", err)
	}

	err = m.RunProjectMigration()
	if err != nil {
		return fmt.Errorf("failed to run project migration: %v", err)
	}

	err = m.RunAPIKeyMigration()
	if err != nil {
		return fmt.Errorf("failed to run api key migration: %v", err)
	}

	err = m.RunEndpointMigration()
	if err != nil {
		return fmt.Errorf("failed to run endpoint migration: %v", err)
	}

	err = m.RunSourceMigration()
	if err != nil {
		return fmt.Errorf("failed to run source migration: %v", err)
	}

	err = m.RunSubscriptionMigration()
	if err != nil {
		return fmt.Errorf("failed to run subsription migration: %v", err)
	}

	if m.MigrateEvents {
		err = m.RunEventMigration()
		if err != nil {
			return fmt.Errorf("failed to run event migration: %v", err)
		}

		err = m.RunEventDeliveriesMigration()
		if err != nil {
			return fmt.Errorf("failed to run event deliveries migration: %v", err)
		}
	}

	return nil
}
