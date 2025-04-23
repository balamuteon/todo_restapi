package repository

import (
	"testing"
	"github.com/stretchr/testify/assert"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestPostgres_NewPostgresDB(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "successful connection",
			cfg: Config{
				Host:     "localhost",
				Port:     "5437",
				Username: "postgres",
				Password: "123",
				DBName:   "todo_rest",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "invalid configuration",
			cfg: Config{
				Host:     "",
				Port:     "",
				Username: "",
				Password: "",
				DBName:   "",
				SSLMode:  "",
			},
			wantErr: true,
		},
		{
			name: "database ping error",
			cfg: Config{
				Host:     "localhost",
				Port:     "5432",
				Username: "testuser",
				Password: "wrongpassword",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDB(tt.cfg)

			if tt.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, db, "db should be nil or error")
			} else {
				assert.NoError(t, db.Ping(), "failed to ping db")
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, db, "db should not be nil")
				assert.NoError(t, db.Close(), "failed to close db")
			}
		})
	}
}
