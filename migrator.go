package main

import (
	"fmt"
	"github.com/frain-dev/convoy/database/hooks"
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
	projects []*datastore.Project

	newDB *sqlx.DB
	oldDB *sqlx.DB
}

var defaultPageable = datastore.Pageable{
	PerPage:    1000,
	Direction:  datastore.Next,
	NextCursor: "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
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

	fmt.Println("Finished user migration")
	err = m.RunOrgMigration()
	if err != nil {
		return fmt.Errorf("failed to run org migration: %v", err)
	}
	fmt.Println("Finished org migration")

	err = m.RunProjectMigration()
	if err != nil {
		return fmt.Errorf("failed to run project migration: %v", err)
	}
	fmt.Println("Finished project migration")

	err = m.RunAPIKeyMigration()
	if err != nil {
		return fmt.Errorf("failed to run api key migration: %v", err)
	}
	fmt.Println("Finished api key migration")

	err = m.RunEndpointMigration()
	if err != nil {
		return fmt.Errorf("failed to run endpoint migration: %v", err)
	}
	fmt.Println("Finished endpoint migration")

	err = m.RunSourceMigration()
	if err != nil {
		return fmt.Errorf("failed to run source migration: %v", err)
	}
	fmt.Println("Finished source migration")

	err = m.RunSubscriptionMigration()
	if err != nil {
		return fmt.Errorf("failed to run subsription migration: %v", err)
	}
	fmt.Println("Finished subscription migration")

	if m.MigrateEvents {
		err = m.RunEventMigration()
		if err != nil {
			return fmt.Errorf("failed to run event migration: %v", err)
		}
		fmt.Println("Finished event migration")

		err = m.RunEventDeliveriesMigration()
		if err != nil {
			return fmt.Errorf("failed to run event deliveries migration: %v", err)
		}
		fmt.Println("Finished event delivery migration")
	}

	return nil
}

func (p *Migrator) GetDB() *sqlx.DB {
	return p.oldDB
}

func (p *Migrator) Close() error {
	return nil
}

func (p *Migrator) GetHook() *hooks.Hook {
	return nil
}
