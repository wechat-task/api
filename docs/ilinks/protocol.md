# WeChat Clawbot iLink Protocol

## Overview

The WeChat Clawbot iLink protocol provides QR code-based authentication and real-time messaging for bot integrations. This is an official Tencent product backed by the "WeChat ClawBot Usage Terms" (微信ClawBot功能使用条款), with legal jurisdiction in Nanshan District, Shenzhen.

**Base URL**: `https://ilinkai.weixin.qq.com`

**Protocol**: HTTP/JSON, no SDK required, can be called directly with `fetch`.

---

## Authentication

### Request Headers

All authenticated requests must include the following headers:

```http
Content-Type: application/json
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
Authorization: Bearer <bot_token>
```

`X-WECHAT-UIN` is regenerated as a random uint32 on every request (decimal string → base64 encoded). It serves as an anti-replay mechanism.

### ID Format Convention

| Type | Format | Example |
|------|--------|---------|
| User ID | `xxx@im.wechat` | `o9cq80_-YQ4SsQEctLny00QWNWd4@im.wechat` |
| Bot ID | `xxx@im.bot` | `ac84a3ae8e58@im.bot` |

---

## Endpoints

### 1. Create QR Code

Generates a QR code for bot authentication.

#### Request

```http
GET /ilink/bot/get_bot_qrcode?bot_type=3
```

**Query Parameters:**

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| bot_type  | number | Yes      | Bot type identifier (default: 3) |

#### Response

**Success Response (200 OK)**

```json
{
  "qrcode": "20b6099d6aafc51079afbd4f96ebae0a",
  "qrcode_img_content": "https://liteapp.weixin.qq.com/q/7GiQu1?qrcode=20b6099d6aafc51079afbd4f96ebae0a&bot_type=3",
  "ret": 0
}
```

**Response Fields:**

| Field              | Type   | Description |
|--------------------|--------|-------------|
| qrcode             | string | Unique QR code identifier |
| qrcode_img_content | string | URL to the QR code image |
| ret                | number | Return code (0 = success) |

---

### 2. Query QR Code Status

Checks the status of a previously generated QR code.

#### Request

```http
GET /ilink/bot/get_qrcode_status?qrcode={qrcode_id}
```

**Query Parameters:**

| Parameter | Type   | Required | Description                |
|-----------|--------|----------|----------------------------|
| qrcode    | string | Yes      | QR code identifier from Create QR Code response |

#### Response

The response varies based on the current status of the QR code.

**Status: Expired**

The QR code has expired and is no longer valid.

```json
{
  "ret": 0,
  "status": "expired"
}
```

**Status: Wait**

Waiting for user to scan the QR code.

```json
{
  "ret": 0,
  "status": "wait"
}
```

**Status: Confirmed**

User has successfully scanned and confirmed the QR code.

```json
{
  "baseurl": "https://ilinkai.weixin.qq.com",
  "bot_token": "ac84a3ae8e58@im.bot:06000083cd68bc60c51089a9fc8c1862b2803a",
  "ilink_bot_id": "ac84a3ae8e58@im.bot",
  "ilink_user_id": "o9cq80_-YQ4SsQEctLny00QWNWd4@im.wechat",
  "ret": 0,
  "status": "confirmed"
}
```

**Response Fields:**

| Field         | Type   | Description |
|---------------|--------|-------------|
| ret           | number | Return code (0 = success) |
| status        | string | QR code status: `expired`, `wait`, or `confirmed` |
| baseurl       | string | Base URL for iLink API (only when confirmed) |
| bot_token     | string | Authentication token for the bot (only when confirmed) |
| ilink_bot_id  | string | Bot identifier (only when confirmed) |
| ilink_user_id | string | User identifier (only when confirmed) |

---

### 3. Get Updates (Long Polling)

Core endpoint for receiving messages. Identical design to Telegram Bot API's `getUpdates`.

#### Request

```http
POST /ilink/bot/getupdates
Authorization: Bearer <bot_token>
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
```

```json
{
  "get_updates_buf": "<cursor from last response, empty string on first call>",
  "base_info": { "channel_version": "1.0.2" }
}
```

**Request Fields:**

| Field            | Type   | Description |
|------------------|--------|-------------|
| get_updates_buf  | string | Cursor from previous response; empty string on first call |
| base_info        | object | Client metadata |
| base_info.channel_version | string | Channel version (e.g. "1.0.2") |

#### Response

The server **holds the connection for up to 35 seconds** until new messages arrive.

```json
{
  "ret": 0,
  "msgs": [ ...WeixinMessage[] ],
  "get_updates_buf": "<new cursor for next request>",
  "longpolling_timeout_ms": 35000
}
```

**Response Fields:**

| Field                  | Type    | Description |
|------------------------|---------|-------------|
| ret                    | number  | Return code (0 = success) |
| msgs                   | array   | Array of `WeixinMessage` objects |
| get_updates_buf        | string  | Cursor for next request (must be updated each time) |
| longpolling_timeout_ms | number  | Long polling timeout in milliseconds (35000) |

> **Important**: `get_updates_buf` acts like a database cursor. You must pass the new value with each subsequent request, otherwise you will receive duplicate messages.

---

### 4. Send Message

Send a message (text/image/file/video/voice) to a user.

#### Request

```http
POST /ilink/bot/sendmessage
Authorization: Bearer <bot_token>
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
```

```json
{
  "msg": {
    "to_user_id": "o9cq800kum_xxx@im.wechat",
    "message_type": 2,
    "message_state": 2,
    "context_token": "<from inbound message>",
    "item_list": [
      { "type": 1, "text_item": { "text": "Hello!" } }
    ]
  }
}
```

**Message Fields:**

| Field          | Type   | Description |
|----------------|--------|-------------|
| to_user_id     | string | Recipient user ID |
| message_type   | number | 1 = inbound (user), 2 = outbound (bot) |
| message_state  | number | 2 = FINISH (complete message) |
| context_token  | string | **Required** — must match the token from the inbound message |
| item_list      | array  | Array of message items (see Message Item Types below) |

> **Critical**: Every inbound message carries a `context_token`. You **must** include this token unchanged in your reply, or the message will not be associated with the correct conversation.

---

### 5. Get Upload URL

Get a pre-signed CDN upload URL for media files.

#### Request

```http
POST /ilink/bot/getuploadurl
Authorization: Bearer <bot_token>
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
```

#### Media Upload Flow

1. Generate a random AES-128 key
2. Encrypt the file with AES-128-ECB
3. Call `getuploadurl` to get a pre-signed URL
4. `PUT` the encrypted file to CDN (`https://novac2c.cdn.weixin.qq.com/c2c`)
5. In `sendmessage`, include `aes_key` (base64) and CDN reference parameters

---

### 6. Get Config

Retrieve configuration including `typing_ticket`.

#### Request

```http
POST /ilink/bot/getconfig
Authorization: Bearer <bot_token>
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
```

---

### 7. Send Typing

Send "typing" indicator to the user.

#### Request

```http
POST /ilink/bot/sendtyping
Authorization: Bearer <bot_token>
AuthorizationType: ilink_bot_token
X-WECHAT-UIN: <base64(random_uint32)>
```

---

## Complete API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ilink/bot/get_bot_qrcode` | GET | Get login QR code (`?bot_type=3`) |
| `/ilink/bot/get_qrcode_status` | GET | Poll QR code scan status (`?qrcode=xxx`) |
| `/ilink/bot/getupdates` | POST | **Long poll for messages** (core) |
| `/ilink/bot/sendmessage` | POST | Send message (text/image/file/video/voice) |
| `/ilink/bot/getuploadurl` | POST | Get CDN pre-signed upload URL |
| `/ilink/bot/getconfig` | POST | Get typing_ticket |
| `/ilink/bot/sendtyping` | POST | Send "typing" indicator |

**CDN Domain**: `https://novac2c.cdn.weixin.qq.com/c2c`

---

## Message Structure

### WeixinMessage

```json
{
  "from_user_id": "o9cq800kum_xxx@im.wechat",
  "to_user_id": "e06c1ceea05e@im.bot",
  "message_type": 1,
  "message_state": 2,
  "context_token": "AARzJWAFAAABAAAAAAAp...",
  "item_list": [
    {
      "type": 1,
      "text_item": { "text": "Hello" }
    }
  ]
}
```

### Message Item Types

| type | Description |
|------|-------------|
| 1    | Text |
| 2    | Image (CDN encrypted storage) |
| 3    | Voice (silk encoding, with speech-to-text) |
| 4    | File attachment |
| 5    | Video |

---

## Media Encryption

All media files on WeChat CDN are encrypted with **AES-128-ECB**:

- **Upload**: `encryptAesEcb(fileBuffer, aesKey)`
- **Download**: `decryptAesEcb(encryptedBuffer, aesKey)`

---

## Authentication Flow

```
Developer          iLink Server          WeChat User
   │                    │                     │
   │── GET get_bot_qrcode ──▶                │
   │◀── { qrcode, url } ──│                  │
   │                    │◀── User scans ─────│
   │── GET get_qrcode_status ──▶ (long poll) │
   │◀── { status: "confirmed",              │
   │      bot_token, baseurl } ──│           │
   │                    │                     │
   │  Persist bot_token, use Bearer auth     │
   │  for all subsequent requests            │
```

1. **Create QR Code**: Call `get_bot_qrcode` to generate a QR code
2. **Display QR Code**: Show the QR code image to the user via `qrcode_img_content`
3. **Poll Status**: Periodically call `get_qrcode_status` until confirmed
4. **Receive Confirmation**: Extract `bot_token`, `ilink_bot_id`, `ilink_user_id`
5. **Authenticate**: Use `bot_token` as Bearer token for all subsequent API calls

---

## Status Codes

| Status    | Description |
|-----------|-------------|
| expired   | QR code has expired |
| wait      | Waiting for user to scan |
| confirmed | User successfully scanned and confirmed |

## Error Handling

All endpoints return `ret: 0` on success. Non-zero return codes indicate errors (specific error codes to be documented).

---

## Legal & Compliance

Tencent's positioning: **iLink is a message channel only**. Tencent does not store your input/output content, does not provide AI services, and is not responsible for AI output.

Key constraints from the WeChat ClawBot Usage Terms:
- Tencent can rate-limit or block specific AI service connections at any time
- Content filtering/interception is applied
- Tencent can terminate the service at any time
- IP addresses, operation logs, and device info are collected for security auditing
- Reverse engineering or bypassing WeChat's technical protection measures is prohibited

**Do not build core business entirely on this API** — have a fallback plan.

---

## Known Limitations

1. `bot_type=3` meaning is not fully documented — hardcoded in source, may correspond to specific account types
2. Requires OpenClaw account system — login needs Tencent iLink server connection
3. Group chat support — source has `group_id` field and `ChatType: "direct"`, group chat may require extra permissions
4. No message history API — only `get_updates_buf` cursor mechanism
5. Rate limits — not publicly documented, requires testing

---

*Based on analysis of `@tencent-weixin/openclaw-weixin@1.0.2` source code and live testing, as of March 2026.*
*Last Updated: 2026-03-30*
