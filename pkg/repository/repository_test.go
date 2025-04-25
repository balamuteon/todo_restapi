package repository

import (
	// "reflect"
	"testing"
	// "github.com/jmoiron/sqlx"
	todo "github.com/balamuteon/todo_restapi"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestRepositiry_NewRepository(t *testing.T) {
	t.Run("valid db", func(t *testing.T) {
		db, _, _, _, cleanup := setupTestDB(t)
		defer cleanup()

		_, err	:= db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		repo := NewRepository(db)

		assert.IsType(t, repo.Authorization, &AuthPostgres{})
		assert.IsType(t, repo.TodoList, &TodoListPostgres{})
		assert.IsType(t, repo.TodoItem, &TodoItemPostgres{})

		id, err := repo.Authorization.CreateUser(todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		})
		
		// проверяем, что метод вернул валидный бд, вызвав соответствующие методы
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, 1, id, "expected ID=1")

		listId, err := repo.TodoList.Create(id, todo.TodoList{
			Title:       "Test List",
			Description: "Test tasks",
		})
		assert.NoError(t, err, "expected no error from Create list")
		assert.Equal(t, 1, listId, "expected list ID=1")

		itemId, err := repo.TodoItem.Create(listId, todo.TodoItem{
			Title:       "Test Item",
			Description: "Test description",
		})
		assert.NoError(t, err, "expected no error from Create item")
		assert.Equal(t, 1, itemId, "expected item ID=1")
	})
}