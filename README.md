# test-task-org-api

REST API для управления организационной структурой компании: подразделения и сотрудники с поддержкой иерархического дерева.

## Стек

- **Go 1.26** — net/http (без сторонних роутеров)
- **PostgreSQL 16** — основная БД
- **GORM** — ORM для работы с БД
- **goose** — миграции
- **Docker + docker-compose** — контейнеризация

---

## Быстрый старт

### Через Docker (рекомендуется)

```bash
git clone https://github.com/austyuzhaninov/test-task-org-api.git
cd test-task-org-api

cp .env.example .env

docker-compose up --build
```

Приложение доступно на `http://localhost:8080`

### Локальная разработка

```bash
# 1. Поднять только БД
make docker-db

# 2. Загрузить переменные окружения
cp .env.example .env
export $(cat .env | grep -v '#' | xargs)

# 3. Запустить приложение
make run
```

---

## Переменные окружения

| Переменная    | Описание              | По умолчанию |
|---------------|-----------------------|--------------|
| `APP_PORT`    | Порт HTTP сервера     | `8080`       |
| `DB_HOST`     | Хост PostgreSQL       | `localhost`  |
| `DB_PORT`     | Порт PostgreSQL       | `5432`       |
| `DB_USER`     | Пользователь БД       | `postgres`   |
| `DB_PASSWORD` | Пароль БД             | `postgres`   |
| `DB_NAME`     | Имя базы данных       | `org`        |

---

## API

### Подразделения

| Метод    | Путь                  | Описание                         |
|----------|-----------------------|----------------------------------|
| `POST`   | `/departments`        | Создать подразделение            |
| `GET`    | `/departments/{id}`   | Получить подразделение с деревом |
| `PATCH`  | `/departments/{id}`   | Переименовать / переместить      |
| `DELETE` | `/departments/{id}`   | Удалить подразделение            |

### Сотрудники

| Метод  | Путь                          | Описание                    |
|--------|-------------------------------|-----------------------------|
| `POST` | `/departments/{id}/employees` | Добавить сотрудника в отдел |

---

## Примеры запросов

**Создать подразделение**
```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering"}'
```

**Создать дочернее подразделение**
```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name": "Backend", "parent_id": 1}'
```

**Получить подразделение с деревом глубиной 3**
```bash
curl "http://localhost:8080/departments/1?depth=3&include_employees=true"
```

**Переместить подразделение**
```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{"parent_id": 5}'
```

**Удалить каскадно**
```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=cascade"
```

**Удалить с переводом сотрудников**
```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1"
```

**Добавить сотрудника**
```bash
curl -X POST http://localhost:8080/departments/1/employees \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Иван Иванов", "position": "Senior Developer", "hired_at": "2024-01-15"}'
```

---

## Структура проекта

```
test-task-org-api/
├── cmd/
│   └── api/
│       └── main.go                  # Точка входа, DI, запуск сервера
├── internal/
│   ├── config/
│   │   └── config.go                # Конфигурация через env
│   ├── domain/
│   │   ├── department.go            # Сущность + интерфейсы репозитория и сервиса
│   │   ├── employee.go              # Сущность + интерфейсы репозитория и сервиса
│   │   └── errors.go                # Доменные ошибки
│   ├── repository/
│   │   ├── department.go            # GORM реализация DepartmentRepository
│   │   ├── employee.go              # GORM реализация EmployeeRepository
│   │   └── errors.go                # mapDBError — маппинг ошибок БД
│   ├── service/
│   │   ├── department.go            # Бизнес-логика подразделений
│   │   └── employee.go              # Бизнес-логика сотрудников
│   ├── handler/
│   │   ├── dto/
│   │   │   ├── department.go        # Request/Response DTO + конвертеры
│   │   │   └── employee.go          # Request/Response DTO + конвертеры
│   │   ├── respond/
│   │   │   └── respond.go           # Responder — JSON ответы и маппинг ошибок
│   │   ├── testhelper/
│   │   │   ├── mocks.go             # Моки репозиториев для тестов
│   │   │   └── setup.go             # Сборка тестового стека
│   │   ├── department.go            # HTTP хендлеры подразделений
│   │   ├── employee.go              # HTTP хендлеры сотрудников
│   │   ├── helpers.go               # pathID, queryInt, queryBool
│   │   └── router.go                # Регистрация маршрутов
│   └── middleware/
│       └── logger.go                # Логирование HTTP запросов
├── migrations/
│   ├── embed.go                     # embed.FS для вшивания SQL в бинарник
│   ├── 001_create_departments.sql
│   └── 002_create_employees.sql
├── pkg/
│   └── logger/
│       └── logger.go                # slog JSON логгер
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

---

## Тесты

```bash
make test
```

Тесты покрывают handler слой через `httptest` с моками репозиториев:

- Создание подразделения (успех, пустое имя, несуществующий parent)
- Получение подразделения (успех, not found)
- Обновление (переименование, защита от цикла в дереве)
- Удаление (cascade, reassign, без mode)

---

## Бизнес-логика

- Имя подразделения уникально в пределах одного родителя
- Нельзя переместить подразделение в своё собственное поддерево (409 Conflict)
- Нельзя сделать подразделение родителем самого себя (409 Conflict)
- Удаление `cascade` — удаляет отдел, всех сотрудников и дочерние отделы
- Удаление `reassign` — переводит сотрудников в другой отдел атомарно (транзакция), дочерние отделы должны отсутствовать