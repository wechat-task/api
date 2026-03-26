# WeChat Clawbot iLink Protocol

## Overview

The WeChat Clawbot iLink protocol provides QR code-based authentication for bot integrations. This document describes the API endpoints for creating QR codes and querying their status.

**Base URL**: `https://ilinkai.weixin.qq.com`

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

## Status Codes

| Status    | Description |
|-----------|-------------|
| expired   | QR code has expired |
| wait      | Waiting for user to scan |
| confirmed | User successfully scanned and confirmed |

## Authentication Flow

1. **Create QR Code**: Call the Create QR Code endpoint to generate a QR code
2. **Display QR Code**: Show the QR code image to the user via `qrcode_img_content`
3. **Poll Status**: Periodically call Query QR Code Status to check if user has scanned
4. **Receive Confirmation**: When status changes to `confirmed`, extract bot credentials
5. **Authenticate**: Use `bot_token`, `ilink_bot_id`, and `ilink_user_id` for subsequent API calls

## Error Handling

All endpoints return `ret: 0` on success. Non-zero return codes indicate errors (specific error codes to be documented).

---

*Document Version: 1.0*
*Last Updated: 2026-03-26*
