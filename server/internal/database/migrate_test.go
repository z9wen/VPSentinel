package database

import "testing"

func TestEmbeddedMigrationsAreValid(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("migration count = %d, want 1", len(migrations))
	}
	if migrations[0].version != "000001" || migrations[0].checksum == "" {
		t.Fatalf("migration = %#v, want version 000001 with checksum", migrations[0])
	}
}
