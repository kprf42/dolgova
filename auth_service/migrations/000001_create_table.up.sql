-- Создание таблицы пользователей
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Посты (связь с пользователями через author_id)
CREATE TABLE posts (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    author_id   TEXT NOT NULL,
    category_id TEXT,
    is_pinned   INTEGER DEFAULT 0, -- 0 = false, 1 = true
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- Комментарии
CREATE TABLE comments (
    id         TEXT PRIMARY KEY,
    content    TEXT NOT NULL,
    post_id    TEXT NOT NULL,
    author_id  TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (author_id) REFERENCES users(id)
);

-- Сообщения чата
CREATE TABLE chat_messages (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    text       TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);