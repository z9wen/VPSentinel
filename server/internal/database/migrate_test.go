package database

import "testing"

func TestEmbeddedMigrationsAreValid(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("migration count = %d, want 2", len(migrations))
	}
	for index, version := range []string{"000001", "000002"} {
		if migrations[index].version != version || migrations[index].checksum == "" {
			t.Fatalf("migration = %#v, want version %s with checksum", migrations[index], version)
		}
	}
}
