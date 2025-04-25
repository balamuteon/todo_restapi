package repository

import (
	"testing"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTodoListPostgres_NewTodoListPostgres(t *testing.T) {
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
			todoList := NewTodoListPostgres(db)

			if tt.wantErr {
				assert.Nil(t, todoList.db, "db should be nil or error")
			} else {
				assert.NoError(t, todoList.db.Ping(), "failed to ping db")
				assert.NotNil(t, todoList.db, "db should not be nil")
				assert.NoError(t, todoList.db.Close(), "failed to close db")
			}
		})
	}
}

func TestTodoListPostgres_Create(t *testing.T) {
	t.Run("successfully create list", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		userId := createTestUser(t, authRepo, db)

		// Создаем список
		listId, list := createTestList(t, todoListRepo, userId)

		// Проверяем запись в todo_lists
		var dbList todo.TodoList
		err = db.Get(&dbList, "SELECT id, title, description FROM todo_lists WHERE id=$1", listId)
		assert.NoError(t, err, "failed to fetch list")
		assert.Equal(t, list.Title, dbList.Title, "title mismatch")
		assert.Equal(t, list.Description, dbList.Description, "description mismatch")
		assert.Equal(t, listId, dbList.Id, "list ID mismatch")

		// Проверяем связь в users_lists
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM users_lists WHERE user_id=$1 AND list_id=$2", userId, listId)
		assert.NoError(t, err, "failed to check users_lists")
		assert.Equal(t, 1, count, "expected one record in users_lists")
	})

	t.Run("error: non-existent userId", func(t *testing.T) {
		db, todoListRepo, _, _, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Пытаемся создать список с несуществующим userId
		non_existing_userId := 999
		list := todo.TodoList{}
		listId, err := todoListRepo.Create(non_existing_userId, list)
		assert.Error(t, err, "expected error")
		assert.Equal(t, 0, listId, "expected list ID=0")

		// Проверяем, что список не создан
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM todo_lists")
		assert.NoError(t, err, "failed to check todo_lists")
		assert.Equal(t, 0, count, "expected no records in todo_lists")

		// Проверяем, что связь не создана
		err = db.Get(&count, "SELECT COUNT(*) FROM users_lists")
		assert.NoError(t, err, "failed to check users_lists")
		assert.Equal(t, 0, count, "expected no records in users_lists")
	})
}

func TestTodoListPostgres_GetAll(t *testing.T) {
	t.Run("successfully get all lists", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		user := todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		}
		userId, err := authRepo.CreateUser(user)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем списки
		lists := []todo.TodoList{
			{
				Title:       "First List",
				Description: "My tasks",
			},
			{
				Title:       "Second List",
				Description: "Other tasks",
			},
			{
				Title:       "Third List",
				Description: "Additional tasks",
			},
		}
		for idx := range lists {
			listId, err := todoListRepo.Create(userId, lists[idx])
			assert.NoError(t, err, "expected no error")
			assert.Equal(t, idx+1, listId, "expected list ID=1")
		}

		toCompareList, err := todoListRepo.GetAll(userId)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, len(lists), len(toCompareList), "have to be equal")
	})

	t.Run("non-existent userId", func(t *testing.T) {
		db, todoListRepo, _, _, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Проверяем GetAll для несуществующего userId
		randomUserId := 123
		lists, err := todoListRepo.GetAll(randomUserId)
		assert.NoError(t, err, "expected no error")
		assert.Empty(t, lists, "expected empty list")
	})
}

func TestTodoListPostgres_GetById(t *testing.T) {
	t.Run("successfull getting by id", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		user := todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		}
		userId, err := authRepo.CreateUser(user)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// Создаем списки
		lists := []todo.TodoList{
			{
				Title:       "First List",
				Description: "My tasks",
			},
			{
				Title:       "Second List",
				Description: "Other tasks",
			},
			{
				Title:       "Third List",
				Description: "Additional tasks",
			},
		}
		for idx := range lists {
			listId, err := todoListRepo.Create(userId, lists[idx])
			assert.NoError(t, err, "expected no error")
			assert.Equal(t, idx+1, listId, "expected list ID=1")
		}
		// check lists by id
		for i := 1; i <= len(lists); i++ {
			todoList, err := todoListRepo.GetById(userId, i)
			assert.NoError(t, err, "expected no error")
			assert.Equal(t, todoList.Title, lists[i-1].Title, "expected to be equal")
			assert.Equal(t, todoList.Description, lists[i-1].Description, "expected to be equal")
			assert.Equal(t, todoList.Id, i, "expected to be equal")
			assert.NotEmpty(t, todoList, "expected not empty list")
		}
	})
}

func TestTodoListPostgres_Delete(t *testing.T) {
	t.Run("successfull deleting by id", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Создаем пользователя
		user := todo.User{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "hashedpassword",
		}
		userId, err := authRepo.CreateUser(user)
		assert.NoError(t, err, "failed to create user")
		assert.Equal(t, 1, userId, "expected user ID=1")

		// создаем список
		list := todo.TodoList{
			Title:       "First List",
			Description: "My tasks",
		}
		listId, err := todoListRepo.Create(userId, list)
		assert.NoError(t, err, "expected no error")
		assert.Equal(t, 1, listId, "expected list ID=1")

		// удаляем список
		err = todoListRepo.Delete(userId, listId)
		assert.NoError(t, err, "expected no error")

		// проверяем, что список удален
		_, err = todoListRepo.GetById(userId, listId)
		assert.Error(t, err, "expected error")
	})
}

func TestTodoListPostgres_Update(t *testing.T) {
	t.Run("update only title", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, originalList := createTestList(t, todoListRepo, userId)

		// Обновление
		newTitle := "Updated Title"
		input := todo.UpdateListInput{
			Title: &newTitle,
		}
		err = todoListRepo.Update(userId, listId, input)
		assert.NoError(t, err, "expected no error")

		// Проверка
		expectedList := originalList
		expectedList.Title = newTitle
		checkList(t, db, listId, expectedList)
	})

	t.Run("update only description", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, originalList := createTestList(t, todoListRepo, userId)

		// Обновление
		newDescription := "Updated tasks"
		input := todo.UpdateListInput{
			Description: &newDescription,
		}
		err = todoListRepo.Update(userId, listId, input)
		assert.NoError(t, err, "expected no error")

		// Проверка
		expectedList := originalList
		expectedList.Description = newDescription
		checkList(t, db, listId, expectedList)
	})

	t.Run("update both title and description", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, originalList := createTestList(t, todoListRepo, userId)

		// Обновление
		newTitle := "Updated Title"
		newDescription := "Updated tasks"
		input := todo.UpdateListInput{
			Title:       &newTitle,
			Description: &newDescription,
		}
		err = todoListRepo.Update(userId, listId, input)
		assert.NoError(t, err, "expected no error")

		// Проверка
		expectedList := originalList
		expectedList.Title = newTitle
		expectedList.Description = newDescription
		checkList(t, db, listId, expectedList)
	})

	t.Run("no fields to update", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, originalList := createTestList(t, todoListRepo, userId)

		// Обновление
		input := todo.UpdateListInput{}
		err = todoListRepo.Update(userId, listId, input)
		assert.Error(t, err, "expected error due to invalid query")
		assert.Contains(t, err.Error(), "syntax error", "expected SQL syntax error")

		// Проверка, что список не изменился
		checkList(t, db, listId, originalList)
	})

	t.Run("non-existent userId or listId", func(t *testing.T) {
		db, todoListRepo, _, authRepo, cleanup := setupTestDB(t)
		defer cleanup()

		// Очистка таблиц
		_, err := db.Exec("TRUNCATE TABLE users, todo_lists, users_lists RESTART IDENTITY CASCADE")
		assert.NoError(t, err, "failed to truncate tables")

		// Подготовка данных
		userId := createTestUser(t, authRepo, db)
		listId, originalList := createTestList(t, todoListRepo, userId)

		// Обновление с несуществующим userId
		newTitle := "Updated Title"
		input := todo.UpdateListInput{Title: &newTitle}
		err = todoListRepo.Update(999, listId, input)
		assert.NoError(t, err, "expected no error, but no rows affected")

		// Проверка, что список не изменился
		checkList(t, db, listId, originalList)

		// Обновление с несуществующим listId
		err = todoListRepo.Update(userId, 999, input)
		assert.NoError(t, err, "expected no error, but no rows affected")

		// Проверка, что список не изменился
		checkList(t, db, listId, originalList)
	})
}


func setupTestDB(t *testing.T) (*sqlx.DB, *TodoListPostgres, *TodoItemPostgres, *AuthPostgres, func()) {
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

	todoListRepo := &TodoListPostgres{db: db}
	todoItemRepo := &TodoItemPostgres{db: db}
	authRepo := &AuthPostgres{db: db}
	cleanup := func() {
		db.Close()
	}

	return db, todoListRepo, todoItemRepo, authRepo, cleanup
}


func createTestUser(t *testing.T, authRepo *AuthPostgres, db *sqlx.DB) int {
	_, err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	assert.NoError(t, err, "failed to truncate users")
	user := todo.User{
		Name:     "John Doe",
		Username: "johndoe",
		Password: "hashedpassword",
	}
	userId, err := authRepo.CreateUser(user)
	assert.NoError(t, err, "failed to create user")
	assert.Equal(t, 1, userId, "expected user ID=1")
	return userId
}

func createTestList(t *testing.T, todoListRepo *TodoListPostgres, userId int) (int, todo.TodoList) {
	list := todo.TodoList{
		Title:       "Original List",
		Description: "Original tasks",
	}
	listId, err := todoListRepo.Create(userId, list)
	assert.NoError(t, err, "failed to create list")
	assert.Equal(t, 1, listId, "expected list ID=1")
	list.Id = listId
	return listId, list
}

func createTestItem(t *testing.T, todoItemRepo *TodoItemPostgres, listId int) (int, todo.TodoItem) {
	item := todo.TodoItem{
		Title:       "Important",
		Description: "Make something important",
		Done: false,
	}
	itemId, err := todoItemRepo.Create(listId, item)
	assert.NoError(t, err, "failed to create item")
	assert.Equal(t, 1, itemId, "expected list ID=1")
	item.Id = itemId
	return itemId, item
}

func checkList(t *testing.T, db *sqlx.DB, listId int, expected todo.TodoList) {
	var dbList todo.TodoList
	err := db.Get(&dbList, "SELECT id, title, description FROM todo_lists WHERE id=$1", listId)
	assert.NoError(t, err, "failed to fetch list")
	assert.Equal(t, expected, dbList, "list mismatch")
}
