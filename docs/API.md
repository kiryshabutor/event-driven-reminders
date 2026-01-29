# API Documentation

Все запросы к API проходят через **API Gateway**.
Base URL: `/` (обычно `http://localhost:8080`)

## Auth Service

### Регистрация
`POST /auth/register`

Базовая регистрация нового пользователя.

**Request:**
```json
{
  "username": "user123",
  "password": "strongPassword123"
}
```

**Response (201 Created):**
```json
{
  "user_id": "uuid-string"
}
```

### Вход (Login)
`POST /auth/login`

Аутентификация пользователя и получение пары токенов.

**Request:**
```json
{
  "username": "user123",
  "password": "strongPassword123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "jwt-token-string",
  "refresh_token": "refresh-token-string"
}
```

### Обновление токена (Refresh)
`POST /auth/refresh`

Получение новой пары токенов с использованием Refresh токена.

**Request:**
```json
{
  "refresh_token": "refresh-token-string"
}
```

**Response (200 OK):**
```json
{
  "access_token": "new-jwt-token-string",
  "refresh_token": "new-refresh-token-string"
}
```

### Профиль пользователя
`GET /auth/profile`

Получение информации о текущем пользователе.

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
{
  "user_id": "uuid-string",
  "username": "user123"
}
```

### Выход (Logout)
`POST /auth/logout`

Выход пользователя и инвалидация текущей сессии.

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
{
  "message": "Successfully logged out"
}
```

---

## Reminder Service

### Создать напоминание
`POST /reminders`

**Headers:**
`Authorization: Bearer <access_token>`

**Request:**
```json
{
  "title": "Meeting",
  "description": "Project discussion",
  "remind_at": "2024-12-31T15:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "id": "uuid-string",
  "title": "Meeting",
  "description": "Project discussion",
  "remind_at": "2024-12-31T15:00:00Z",
  "user_id": "uuid-string",
  "status": "pending"
}
```

### Список напоминаний
`GET /reminders`

Получение списка напоминаний с возможностью фильтрации.

**Query Parameters:**
- `status` (optional): Фильтр по статусу (например, `pending`, `sent`).

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "title": "Meeting",
    "status": "pending"
  }
]
```

### Получить напоминание
`GET /reminders/:id`

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
{
  "id": "uuid-string",
  "title": "Meeting",
  "description": "Project discussion",
  "remind_at": "2024-12-31T15:00:00Z",
  "user_id": "uuid-string",
  "status": "pending"
}
```

### Обновить напоминание
`PUT /reminders/:id`

**Headers:**
`Authorization: Bearer <access_token>`

**Request:**
```json
{
  "title": "Updated Meeting",
  "description": "Updated discussion",
  "remind_at": "2024-12-31T16:00:00Z"
}
```

**Response (200 OK):**
```json
{
  "id": "uuid-string",
  "title": "Updated Meeting",
  "description": "Updated discussion",
  "remind_at": "2024-12-31T16:00:00Z",
  "status": "pending"
}
```

### Удалить напоминание
`DELETE /reminders/:id`

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
{
  "message": "Reminder deleted successfully"
}
```

---

## Analytics Service

### Статистика пользователя
`GET /analytics/me`

Получение статистики по напоминаниям текущего пользователя.

**Headers:**
`Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
{
  "total_reminders": 10,
  "sent_reminders": 5,
  "pending_reminders": 5
}
```
