-- Удаление таблицы пользователей
DROP TABLE IF EXISTS users;

-- Триггер для удаления старых сообщений (> 30 дней)
CREATE TRIGGER clean_old_chat
AFTER INSERT ON chat_messages
BEGIN
    DELETE FROM chat_messages 
    WHERE created_at < datetime('now', '-30 days');
END;