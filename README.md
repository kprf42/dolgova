# Go Package Utils

Common Go packages and utilities.

## Logger Package

Пакет для логирования на основе zap.Logger с дополнительной функциональностью.

### Установка

```bash
go get github.com/kprf42/dolgova/pkg
```

### Использование

```go
package main

import (
    "github.com/kprf42/dolgova/pkg/logger"
)

func main() {
    log, err := logger.New()
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // Базовое логирование
    log.Info("Hello world")
    log.Debug("Debug message")
    log.Error("Error occurred", logger.Error(err))

    // С дополнительными полями
    log.Info("User action",
        logger.String("user_id", "123"),
        logger.Int("action_id", 456))
}
```

## Конфигурация

```go
config := logger.LogConfig{
    Level:      "debug",  // debug, info, warn, error, fatal
    OutputPath: "stdout", // путь к файлу или stdout
    Format:     "json",   // json или console
}

log, err := logger.NewWithConfig(config)
```

## License

MIT License 