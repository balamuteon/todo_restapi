package repository

import (
	"testing"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/stretchr/testify/assert"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestAuthPostgres_NewAuthPostgres(t *testing.T) {
	t.Run("valid db", func(t *testing.T) {
		cfg := Config{
			Host:     "localhost",
			Port:     "5437",
			Username: "postgres",
			Password: "123",
			DBName:   "todo_rest",
			SSLMode:  "disable",
		}
		db, err := NewPostgresDB(cfg)
		assert.NoError(t, err, "failed to create test db")
		defer db.Close()

		auth := NewAuthPostgres(db)
		assert.NotNil(t, auth, "NewAuthPostgres returned nil, expected *AuthPostgres")
		assert.NotNil(t, auth.db, "NewAuthPostgres db is nil, expected *sqlx.DB")
		assert.Equal(t, db, auth.db, "NewAuthPostgres db does not match input db")
	})

	t.Run("nil db", func(t *testing.T) {
		auth := NewAuthPostgres(nil)
		assert.NotNil(t, auth, "NewAuthPostgres returned nil, expected *AuthPostgres")
		assert.Nil(t, auth.db, "NewAuthPostgres db is not nil, expected nil")
	})
}

func TestAuthPostgres_CreateUser(t *testing.T) {
	t.Run("create user", func(t *testing.T) {
		cfg := Config{
			Host:     "localhost",
			Port:     "5437",
			Username: "postgres",
			Password: "123",
			DBName:   "todo_rest_test",
			SSLMode:  "disable",
		}
		db, err := NewPostgresDB(cfg)
		assert.NoError(t, err, "failed to connect to test database")
		defer db.Close()

		_, err = db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate users table")

		repo := &AuthPostgres{db: db}
		user := todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		}
		id, err := repo.CreateUser(user)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, 1, id, "expected ID=1")

		var dbUser todo.User

		err = db.Get(&dbUser, "SELECT id, name, username, password_hash FROM users WHERE id=$1", id)
		assert.NoError(t, err, "failed to fetch user")
		assert.Equal(t, user.Name, dbUser.Name, "name mismatch")
		assert.Equal(t, user.Username, dbUser.Username, "username mismatch")
		assert.Equal(t, user.Password, dbUser.Password, "password mismatch")
	})

	t.Run("error creating user caused by duplicate username", func(t *testing.T) {
		cfg := Config{
			Host:     "localhost",
			Port:     "5437",
			Username: "postgres",
			Password: "123",
			DBName:   "todo_rest_test",
			SSLMode:  "disable",
		}
		db, err := NewPostgresDB(cfg)
		assert.NoError(t, err, "failed to connect to test database")
		defer db.Close()

		repo := &AuthPostgres{db: db}
		user := todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		}
		id, err := repo.CreateUser(user)

		assert.Error(t, err, "expected error")
		assert.NotEqual(t, 1, id, "expected ID=0")

		var dbUser todo.User

		err = db.Get(&dbUser, "SELECT id, name, username, password_hash FROM users WHERE id=$1", id)
		assert.Error(t, err, "failed to fetch user")
		assert.NotEqual(t, user.Name, dbUser.Name, "name match")
		assert.NotEqual(t, user.Username, dbUser.Username, "username match")
		assert.NotEqual(t, user.Password, dbUser.Password, "password match")
	})
}

func TestAuthPostgres_GetUser(t *testing.T) {
	t.Run("succefully get user", func(t *testing.T) {
		cfg := Config{
			Host:     "localhost",
			Port:     "5437",
			Username: "postgres",
			Password: "123",
			DBName:   "todo_rest_test",
			SSLMode:  "disable",
		}
		db, err := NewPostgresDB(cfg)
		assert.NoError(t, err, "failed to connect to test database")
		defer db.Close()

		repo := &AuthPostgres{db: db}

		db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

		user := todo.User{
			Name:     "Lev Yashin",
			Username: "levyashin",
			Password: "hashedpassword",
		}

		userId, err := repo.CreateUser(user)
		assert.NoError(t, err, "expected no error")
		assert.NotNil(t, user, "expected user")

		testUser, err := repo.GetUser(user.Username, user.Password)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, user.Name, testUser.Name, "user mismatch")
		assert.Equal(t, userId, testUser.Id, "user id mismatch")
		assert.Equal(t, user.Password, testUser.Password, "password mismatch")
	})


	t.Run("succefully get user", func(t *testing.T) {
		cfg := Config{
			Host:     "localhost",
			Port:     "5437",
			Username: "postgres",
			Password: "123",
			DBName:   "todo_rest_test",
			SSLMode:  "disable",
		}
		db, err := NewPostgresDB(cfg)
		assert.NoError(t, err, "failed to connect to test database")
		defer db.Close()

		repo := &AuthPostgres{db: db}

		db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

		invalidUser := todo.User{
			Id:       123,
			Name:     "random,",
			Username: "random",
			Password: "random",
		}

		testUser, err := repo.GetUser(invalidUser.Username, invalidUser.Password)
		assert.Error(t, err, "expected error")
		assert.NotEqual(t, invalidUser.Name, testUser.Name, "user match")
		assert.NotEqual(t, invalidUser.Id, testUser.Id, "user id match")
		assert.NotEqual(t, invalidUser.Password, testUser.Password, "password match")
	})
}
