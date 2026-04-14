# Bot Channel Abstraction Design

## Goal

Refactor bot functionality to support multiple messaging channels (WeChat clawbot, Lark, future platforms) instead of being tightly coupled to iLink only.

## Design Principles

- All channels are equal — any channel can receive/send messages
- Bot and channel have independent lifecycles — create bot first, add channels separately
- Separate API endpoints per channel type for type-specific creation flows
- Single `channels` table with JSONB config for type-specific data

## Data Model

### `bots` table (refactored)

| Column       | Type      | Constraints              |
|--------------|-----------|--------------------------|
| id           | serial    | PK                       |
| user_id      | bigint    | FK → users, indexed      |
| name         | varchar   | NOT NULL                 |
| description  | text      |                          |
| status       | varchar   | pending/active/disconnected/expired |
| created_at   | timestamp |                          |
| updated_at   | timestamp |                          |

Dropped columns: `bot_token`, `base_url`, `ilink_bot_id`, `ilink_user_id`, `last_cursor`, `qrcode_id`, `qrcode_image`.

### `channels` table (new)

| Column      | Type      | Constraints                          |
|-------------|-----------|--------------------------------------|
| id          | serial    | PK                                   |
| bot_id      | bigint    | FK → bots, indexed                   |
| type        | varchar   | `wechat_clawbot` / `lark`            |
| status      | varchar   | pending/active/disconnected/expired  |
| config      | jsonb     | type-specific data                   |
| last_cursor | text      | for polling-based channels           |
| created_at  | timestamp |                                      |
| updated_at  | timestamp |                                      |

### Config structure by type

**wechat_clawbot:**
```json
{
  "ilink_bot_id": "xxx",
  "ilink_user_id": "xxx",
  "bot_token": "xxx",
  "base_url": "https://ilinkai.weixin.qq.com",
  "qrcode_id": "xxx",
  "qrcode_image": "base64..."
}
```

**lark (future):**
```json
{
  "app_id": "xxx",
  "app_secret": "xxx",
  "oauth_url": "xxx"
}
```

## API Design

### Bot CRUD

| Method | Path                    | Description          |
|--------|-------------------------|----------------------|
| POST   | /api/v1/bots            | Create bot (name required) |
| GET    | /api/v1/bots            | List user's bots (with channels) |
| GET    | /api/v1/bots/:id        | Get bot detail (with channels) |
| PUT    | /api/v1/bots/:id        | Update name/description |
| DELETE | /api/v1/bots/:id        | Delete bot + cascade channels |

### Channel management

| Method | Path                                              | Description                    |
|--------|---------------------------------------------------|--------------------------------|
| POST   | /api/v1/bots/:botId/channels/wechat-clawbot       | Create wechat clawbot channel  |
| POST   | /api/v1/bots/:botId/channels/lark                 | Create lark channel (future)   |
| GET    | /api/v1/bots/:botId/channels                      | List bot's channels            |
| DELETE | /api/v1/bots/:botId/channels/:channelId           | Delete a channel               |

Channel creation responses vary by type:
- WeChat clawbot: returns QR code image
- Lark: returns OAuth URL (future)

## Code Structure

```
internal/
├── model/
│   ├── bot.go          # Bot model (simplified)
│   └── channel.go      # Channel model with Config JSONB
├── repository/
│   ├── bot.go          # BotRepository (CRUD only)
│   └── channel.go      # ChannelRepository
├── service/
│   ├── bot.go          # BotService (pure CRUD)
│   └── channel.go      # ChannelService (dispatches by type)
├── handler/
│   ├── bot.go          # BotHandler (refactored)
│   └── channel.go      # ChannelHandler (routes by type)
└── ilink/
    └── ...             # Unchanged
```

Key changes:
- BotService becomes pure CRUD — no QR code polling logic
- ChannelService dispatches by channel type — wechat_clawbot delegates to iLink logic
- Existing iLink package is reused as-is

## Migration Strategy

1. New migration adds `channels` table
2. Migration strips iLink-specific columns from `bots`
3. Data migration: existing iLink data from `bots` migrates to `channels` rows with `type = 'wechat_clawbot'`
4. `bots.name` gets NOT NULL constraint; existing rows default to "My Bot"
