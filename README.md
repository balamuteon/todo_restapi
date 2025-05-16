# ToDo REST API

ToDo REST API — это приложение для управления списком задач (ToDo list). Оно предоставляет RESTful интерфейс для создания, чтения, обновления и удаления задач.

## Основные возможности (CRUD)

- **C**reate: Создание задач
- **R**ead: Получение списка задач
- **U**pdate: Обновление задач
- **D**elete: Удаление задач

## Технологии

- **Язык программирования:** Go
- **База данных:** PostgreSQL
- **Кэширование:** Redis
- **Веб-фреймворк:** Gin
- **Управление конфигурацией:** Viper
- **Логирование:** Logrus
- **Миграции:**

## Установка и запуск

1.  **Клонируйте репозиторий:**

    ```bash
    git clone <URL репозитория>
    cd todo_restapi
    ```

2.  **Инициализируйте модуль Go (если еще не сделано):**

    ```bash
    go mod init [github.com/balamuteon/todo_restapi](https://github.com/leo/todo_restapi)
    ```

3.  **Установите зависимости:**

    ```bash
    go mod tidy
    ```

4.  **Настройте конфигурацию:**

    - Скопируйте файл `config.example.yaml` в `config.yaml` и отредактируйте его, указав настройки подключения к PostgreSQL и Redis.

      ```bash
      cp config.example.yaml config.yaml
      nano config.yaml
      ```

    - Убедитесь, что указаны правильные параметры для `database`, `redis` и `server`.

5.  **Настройте базу данных:**

    - Убедитесь, что у вас установлен и запущен сервер PostgreSQL.
    - Создайте базу данных, указанную в `config.yaml`.
    - Примените миграции, расположенные в папке `migrations/`, используя инструмент для миграций (например, `golang-migrate/migrate`). Пример использования `migrate`:

      ```bash
      # Убедитесь, что путь к директории миграций указан правильно
      migrate -database "postgres://user:password@host:port/dbname?sslmode=disable" -path migrations/ up
      ```

      *(Замените `postgres://user:password@host:port/dbname` на строку подключения к вашей базе данных)*

6.  **Запустите приложение:**

    ```bash
    go run cmd/apiserver/main.go
    ```

    Приложение будет доступно по адресу, указанному в конфигурации (`server.port`).

## Примеры API запросов

### Создание списка

**POST /api/lists**

```json
{
  "title": "Купить продукты",
  "description": "Купить молоко, хлеб и яйца"
}
```

### Успешный ответ (HTTP 201 Created):

```json
{
  "id": 1,
  "title": "Купить продукты",
  "description": "Купить молоко, хлеб и яйца"
}
```

### Получение списка задач

**GET /api/lists**

### Успешный ответ (HTTP 200 OK):

```json
{
  "data": [
    {
      "id": 1,
      "title": "Купить продукты",
      "description": "Купить молоко, хлеб и яйца"
    },
    {
      "id": 2,
      "title": "Сделать отчет",
      "description": "Подготовить ежемесячный отчет"
    }
    // ... другие задачи
  ]
}
```

### Получение списка по ID

**GET /api/lists/{id}**

### Успешный ответ (HTTP 200 OK):

```json
{
  "id": 1,
  "title": "Купить продукты",
  "description": "Купить молоко, хлеб и яйца"
}
```

### Ответ при отсутствии списка (HTTP 404 Not Found):

```json
{
  "error": "list not found"
}
```

### Обновление задачи

**PUT /api/lists/{id}**

*Пример: PUT /api/lists/1*

```json
{
  "title": "Купить продукты",
  "description": "Купить молоко, хлеб, яйца и сыр"
}
```

### Успешный ответ (HTTP 200 OK):

```json
{
  "status": "ok"
}
```

### Ответ при отсутствии задачи (HTTP 404 Not Found):

```json
{
  "error": "list not found"
}
```

### Удаление задачи

**DELETE /api/lists/{id}**

*Пример: DELETE /api/lists/1*

### Успешный ответ (HTTP 200 OK):

```json
{
  "status": "ok"
}
```

### Ответ при отсутствии задачи (HTTP 404 Not Found):

```json
{
  "error": "list not found"
}
```