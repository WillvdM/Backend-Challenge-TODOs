# Todo Backend Challenge
This project is a TODO management API that was built using GoFiber and PostgreSQL. This guide will explain how to set up the project, run it locally,
and interact with the API.

## 1. Project Setup
Clone the repository:
```
git clone <your-repo-url>
cd Backend-Challenge-TODOs
```
 2. Install Go modules:
```
go mod tidy
```
This downloads all dependencies specified in `go.mod`.

 3. Configure your environment:

This project uses a `config.yaml` file to store environment specific settings. These settings include database URL, sort fields, and the deletion mode.

This is the layout:
``` 
deletion_mode: SOFT (or HARD) 

database_url: "postgres://postgres:postgres@localhost:5432/TODOs?sslmode=disable"

sort_fields:
  - id
  - title
  - completed
  - due_date
  - created_at
  - completed_at
  - assignee  
 ```
 Replace `database_url` with your PostgreSQL credentials.

 ## 2. Database setup
 1. Ensure that PostgreSQL is running on your machine.
 2. Run the mirgration file to create the schema:
    * The migration file is located at `migrations/<epoch>_init_db.sql`
    * Execute it in your PostgreSQL client (for example, DbBeaver or psql)
    ` \i path/to/migrations/1771579988_init_db.sql`
    This will create the `users` table, the `todos` table, and the `set_updated_at` trigger. It will also install
    `pgcrypto`, which is a required exxtension.
3. Verify tables:
```
SELECT * FROM users;
SELECT * FROM todos;
```

## 3. Run the API
1. Run the GO  server:
    `run go main.go`
2. The default server address:
    `http://localhost:3000`

## 4. API endpoints
The API supports `users` and `todos`. All reponses are JSON.

### Users:
| Method | Endpoint     | Description                         | Body                                                       |
| ------ | ------------ | ----------------------------------- | ---------------------------------------------------------- |
| POST   | `/users`     | Create a new user (single or batch) | `[{"name": "John", "surname": "Doe", "username": "jdoe"}]` |
| GET    | `/users`     | Get all users                       | None                                                       |
| GET    | `/users/:id` | Get user by ID                      | None                                                       |
| PATCH  | `/users/:id` | Update user info                    | `{"name":"Jane","surname":"Doe","username":"janedoe"}`     |
| DELETE | `/users/:id` | Soft delete a user                  | None                                                       |

### Todos
| Method | Endpoint         | Description                           | Body                                                                                    |
| ------ | ---------------- | ------------------------------------- | --------------------------------------------------------------------------------------- |
| POST   | `/todos`         | Create todos (batch)                  | `[{"title":"Do homework","completed":false,"assignee":"John","due_date":"20-02-2026"}]` |
| GET    | `/todos`         | List todos (supports pagination)      | Query params: `?offset=0&limit=10&sort=id&order=asc`                                    |
| GET    | `/todos/:id`     | Get todo by ID                        | None                                                                                    |
| GET    | `/todos/expired` | List todos past due date & incomplete | None                                                                                    |
| PATCH  | `/todos/:id`     | Update todo                           | `{"title":"Do math homework","completed":true,"due_date":"22-02-2026"}`                 |
| DELETE | `/todos/:id`     | Delete todo (soft/hard per config)    | None                                                                                    |

## 5. Additional notes
1. *Soft delete*: marked in the `deleted_at` column.
2. *Triggers*: `set_updated_at` automatically updates `updated_at` for `todos`.
3. *UUID generation*: `users.id` uses `gen_random_uuid` (pgcrypto).

## 6. Development tips
1. Always run `go mod tidy` after adding new packages.
2. Make sure the database URL in `config.yaml` points to a valid database.
3. Use tools like DBeaver or psql to inspect your data and verify migrations.
