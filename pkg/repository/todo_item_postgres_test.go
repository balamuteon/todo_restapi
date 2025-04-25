package repository

import (
	// "fmt"
	"testing"

	todo "github.com/balamuteon/todo_restapi"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTodoItemPostgres_NewTodoItemPostgres(t *testing.T) {
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
				DBName:   "todo_rest_test",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := NewPostgresDB(tt.cfg)
			todoItem := NewTodoItemPostgres(db)

			if tt.wantErr {
				assert.Nil(t, todoItem.db, "db should be nil or error")
			} else {
				assert.NoError(t, todoItem.db.Ping(), "failed to ping db")
				assert.NotNil(t, todoItem.db, "db should not be nil")
				assert.NoError(t, todoItem.db.Close(), "failed to close db")
			}
		})
	}
}

func TestTodoItemPostgres_Create(t *testing.T) {

	t.Run("successfully create list", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// создаем юзера
		userId := createTestUser(t, authRepo, db)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, _ := createTestList(t, todoListRepo, userId)

		// создаем айтем
		itemId, item := createTestItem(t, todoItemRepo, listId)

		// проверяем что айтем создан
		dbItem := todo.TodoItem{}
		err = db.Get(&dbItem, "SELECT id, title, description, done FROM todo_items WHERE id=$1", itemId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, item, dbItem, "item in db doesn't mathc")

		// Проверяем связь в таблице lists_items
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM lists_items WHERE list_id=$1 AND item_id=$2", listId, itemId)
		assert.NoError(t, err, "failed to check lists_items")
		assert.Equal(t, 1, count, "expected one record in lists_items")
		// добавим еще один итем
		itemId, err = todoItemRepo.Create(listId, item)
		err = db.Get(&count, "SELECT COUNT(*) FROM lists_items WHERE list_id=$1", listId)
		assert.NoError(t, err, "failed to check lists_items")
		assert.Equal(t, 2, count, "expected two record in lists_items")
	})

	t.Run("invalid listId", func(t *testing.T) {
		db, _, todoItemRepo, _, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Пытаемся создать элемент с несуществующим listId
		item := todo.TodoItem{}
		invalidListId := 999
		itemId, err := todoItemRepo.Create(invalidListId, item)
		assert.Error(t, err, "expected error for invalid listId")
		assert.Equal(t, 0, itemId, "expected item ID=0")
	})
}

func TestTodoItemPostgres_GetAll(t *testing.T) {
	t.Run("successful get all items", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// создаем юзера
		userId := createTestUser(t, authRepo, db)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, _ := createTestList(t, todoListRepo, userId)

		items := []todo.TodoItem{
			{
				Title:       "1",
				Description: "One",
				Done:        false,
			},
			{
				Title:       "2",
				Description: "Two",
				Done:        false,
			},
			{
				Title:       "3",
				Description: "Three",
				Done:        false,
			},
		}
		// создаем 3 элемента
		for idx := range items {
			itemId, err := todoItemRepo.Create(listId, items[idx])
			assert.NoError(t, err, "failed to create item %d", idx+1)
			assert.Equal(t, idx+1, itemId, "expected item ID=%d", idx+1)
			items[idx].Id = itemId
		}

		// Получаем элементы
		dbItems, err := todoItemRepo.GetAll(userId, listId)
		assert.NoError(t, err, "failed to get items")
		assert.Equal(t, len(items), len(dbItems), "expected len of item arrays to be equal")
		assert.NotNil(t, dbItems, "expected items to be not nil")
		assert.Equal(t, items, dbItems, "expected items arrays to be equal")
	})

	t.Run("no items found", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		userId := createTestUser(t, authRepo, db)
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, list := createTestList(t, todoListRepo, userId)
		assert.Equal(t, 1, listId, "expected list ID=1")
		assert.NotZero(t, list, "expected non-zero TodoList")

		// Ожидаем что элементов нет
		dbItems, err := todoItemRepo.GetAll(userId, listId)
		assert.Equal(t, len(dbItems), 0, "expected len of item arrays to be zero")
	})
}

func TestTodoItemPostgres_GetById(t *testing.T) {
	t.Run("successful get by id", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// создаем юзера
		userId := createTestUser(t, authRepo, db)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, _ := createTestList(t, todoListRepo, userId)

		// создаем айтем
		itemId, item := createTestItem(t, todoItemRepo, listId)
		assert.NoError(t, err, "failed to create item")
		assert.Equal(t, 1, itemId, "expected item ID=1")
		assert.NotNil(t, item, "expected item to be not nil")

		// Получаем элемент
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.NoError(t, err, "failed to get items")
		assert.NotNil(t, dbItem, "expected item to be not nil")
		assert.Equal(t, item, dbItem, "expected items to be equal")
	})

	t.Run("no items found", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		userId := createTestUser(t, authRepo, db)
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, list := createTestList(t, todoListRepo, userId)
		assert.Equal(t, 1, listId, "expected list ID=1")
		assert.NotZero(t, list, "expected non-zero TodoList")

		// Ожидаем что элемента нет
		randomItemId := 123
		dbItem, err := todoItemRepo.GetById(userId, randomItemId)
		assert.Error(t, err, "expected error")
		assert.Equal(t, dbItem, todo.TodoItem{}, "expected item to be empty")
	})
}

func TestTodoItemPostgres_Delete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// создаем юзера
		userId := createTestUser(t, authRepo, db)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем список
		listId, _ := createTestList(t, todoListRepo, userId)

		// создаем айтем
		itemId, item := createTestItem(t, todoItemRepo, listId)
		assert.NoError(t, err, "failed to create item")
		assert.Equal(t, 1, itemId, "expected item ID=1")
		assert.NotNil(t, item, "expected item to be not nil")

		err = todoItemRepo.Delete(userId, itemId)
		assert.NoError(t, err, "failed to delete item")
		// Ожидаем что элемента нет
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.Error(t, err, "expected error")
		assert.Equal(t, dbItem, todo.TodoItem{}, "expected item to be empty")
	})
}

func TestUpdateItem(t *testing.T) {
	t.Run("update only title", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, _ := createTestList(t, todoListRepo, userId)
		itemId, originalItem := createTestItem(t, todoItemRepo, listId)

		// Обновление
		newTitle := "Updated Title"
		input := todo.UpdateItemInput{
			Title: &newTitle,
		}

		// Проверка
		err = todoItemRepo.Update(userId, itemId, input)
		assert.NoError(t, err, "expected no error")
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, dbItem.Title, newTitle, "expected title to be updated")
		assert.NotEqual(t, dbItem.Title, originalItem.Title, "expected title to be updated")
		assert.Equal(t, dbItem.Description, originalItem.Description, "expected description not to be updated")
	})

	t.Run("update only done", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, _ := createTestList(t, todoListRepo, userId)
		itemId, originalItem := createTestItem(t, todoItemRepo, listId)

		// Обновление
		newDone := true
		input := todo.UpdateItemInput{
			Done: &newDone,
		}
		err = todoItemRepo.Update(userId, itemId, input)
		assert.NoError(t, err, "expected no error")

		// update
		err = todoItemRepo.Update(userId, itemId, input)
		assert.NoError(t, err, "expected no error")
		
		// Проверка
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, dbItem.Done, newDone, "expected done to be updated")
		assert.Equal(t, dbItem.Title, originalItem.Title, "expected title to stay same")
		assert.NotEqual(t, dbItem.Done, originalItem.Done, "expected done to be updated")
	})

	t.Run("update title and description", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, _ := createTestList(t, todoListRepo, userId)
		itemId, originalItem := createTestItem(t, todoItemRepo, listId)

		// Обновление
		newTitle := "Updated Title"
		newDescription := "Updated Description"
		input := todo.UpdateItemInput{
			Title:       &newTitle,
			Description: &newDescription,
		}
		err = todoItemRepo.Update(userId, itemId, input)
		assert.NoError(t, err, "expected no error")

		// update
		err = todoItemRepo.Update(userId, itemId, input)
		assert.NoError(t, err, "expected no error")

		// Проверка
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, dbItem.Description, newDescription, "expected description to be updated")
		assert.Equal(t, dbItem.Title, newTitle, "expected title to be updated")

		assert.NotEqual(t, dbItem.Description, originalItem.Description, "expected title to be updated")
		assert.NotEqual(t, dbItem.Title, originalItem.Title, "expected title to be updated")
	})

	t.Run("no fields to update", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, _ := createTestList(t, todoListRepo, userId)
		itemId, _ := createTestItem(t, todoItemRepo, listId)

		// Обновление
		input := todo.UpdateItemInput{}
		err = todoItemRepo.Update(userId, itemId, input)
		assert.Error(t, err, "expected error due to nil fields in structure")		
	})

	t.Run("non-existent userId or itemId", func(t *testing.T) {
		db, todoListRepo, todoItemRepo, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists, todo_items, lists_items RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, _ := createTestList(t, todoListRepo, userId)
		itemId, originalItem := createTestItem(t, todoItemRepo, listId)

		// Обновление с несуществующим userId
		newTitle := "Updated Title"
		input := todo.UpdateItemInput{Title: &newTitle}
		err = todoItemRepo.Update(999, itemId, input)
		assert.NoError(t, err, "expected no error, but no rows affected")

		// // Проверка, что элемент не изменился

		// Обновление с несуществующим itemId
		err = todoItemRepo.Update(userId, 999, input)
		assert.NoError(t, err, "expected no error, but no rows affected")

		// Проверка, что элемент не изменился
		dbItem, err := todoItemRepo.GetById(userId, itemId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, dbItem, originalItem, "expected item to be unchanged")
		
	})
}
