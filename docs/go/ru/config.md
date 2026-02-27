# Конфигурация Системы (system.yaml)

Централизованный файл `system.yaml` определяет состав и поведение системы PUM (Go).

## Разделы Конфигурации

### 1. system (Общая информация)
- `name`: Название системы.
- `version`: Версия (1.0.0-prototype).
- `env`: Окружение (development/production).

### 2. core_services (Ядро Системы)
Перечень сервисов, без которых приложение не может функционировать. Если `registry`, `identity` или `product` отключены, GUI переходит в режим ожидания.

### 3. optional_services (Дополнительные модули)
Сервисы, которые расширяют функционал системы (Inventory, Network, Task и др.). GUI автоматически обнаруживает их через **Registry Service**.

### 4. external_modules (Внешние интеграции)
Определяет статус внешних систем:
- `mode: "mock"`: Используются локальные заглушки для тестирования.
- `mode: "remote"`: Используются реальные сетевые соединения (LDAP, GLPI).

### 5. discovery (Обнаружение)
Настройки для `Registry`:
- `registry_url`: URL сервиса реестра.
- `heartbeat_interval`: Интервал отправки пульса (30 сек).
- `timeout`: Максимальное время отсутствия пульса до удаления сервиса (60 сек).

## Пример Конфигурации
```yaml
core_services:
  - registry
  - identity
  - product
optional_services:
  - inventory
  - network
external_modules:
  glpi:
    mode: "mock"
```
