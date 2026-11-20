## WebSocket API

Connect to `wss://{apiId}.execute-api.eu-central-1.amazonaws.com/prod` while the app is in the foreground,
with the JWT in the `Authorization: Bearer <token>` header or the `?token=<token>` query parameter.
An invalid token is rejected with 401.

API Gateway closes a connection after 10 minutes without messages and after 2 hours at most,
reconnect with backoff, and refetch chats and the open chat after a reconnect, events sent meanwhile are lost.

### App → server
```json
{"action": "typing", "friendId": 12, "typing": true}
{"action": "ping"}
```
- `typing` – send `true` when the user starts typing and `false` when they stop or send the message,
  the receiver hides the indicator after ~5 s without an update
- `ping` – any unknown action is accepted and ignored, send one every ~5 minutes to keep the connection open

### Server → app
Every event is `{"type": "...", "data": {...}}`. `friendId` is the other user of the chat from the receiver's view.
Message events also reach the sender's devices, the device which sent the request deduplicates by message `id`.

| type | data |
|---|---|
| `message.new` | `{friendId, message}`, `message` as in `GET /chats/:friendId/messages` |
| `message.deleted` | `{friendId, id}` |
| `message.reaction` | `{friendId, id, reaction}`, `reaction` is `null` when removed |
| `chat.read` | `{friendId, readerId, lastReadMessageId, lastReadAt}` |
| `typing` | `{friendId, typing}` |
| `presence` | `{userId, online, lastOnline}` |

### Initial state
- `GET /last-online/:id` – `{time, online}`, 403 when the user isn't a friend

Online status is shown only in the chat header and only for friends: `presence`, `typing` and `GET /last-online/:id`
skip ex-friends, even when their chat still exists. Close the socket already on AppState `inactive`,
a socket left open by a suspended app counts as online until the 10 minute idle timeout.

FCM notifications are sent as before, hide them in the foreground when their chat is open.
