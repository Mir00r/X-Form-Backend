# X-Form Collaboration Service - WebSocket API Documentation

## 📖 Overview

The X-Form Collaboration Service provides real-time collaboration features for form editing through WebSocket connections. This service enables multiple users to collaborate on forms simultaneously with live cursor tracking, real-time updates, and instant messaging.

## 🚀 Getting Started

### Base URL
```
ws://localhost:8080/api/v1/ws
wss://localhost:8080/api/v1/ws (production)
```

### Authentication
All WebSocket connections require JWT authentication. You can authenticate using either:

1. **Authorization Header** (Recommended):
   ```javascript
   const headers = {
     'Authorization': 'Bearer your_jwt_token_here'
   };
   ```

2. **Query Parameter**:
   ```javascript
   const wsUrl = 'ws://localhost:8080/api/v1/ws?token=your_jwt_token_here';
   ```

## 🔌 Connection Example

### JavaScript Client
```javascript
// Using Authorization header (recommended)
const socket = new WebSocket('ws://localhost:8080/api/v1/ws', [], {
  headers: {
    'Authorization': 'Bearer your_jwt_token_here'
  }
});

// Using query parameter
const socket = new WebSocket('ws://localhost:8080/api/v1/ws?token=your_jwt_token_here');

socket.onopen = function(event) {
  console.log('Connected to collaboration service');
  
  // Join a form collaboration session
  socket.send(JSON.stringify({
    type: 'join:form',
    payload: {
      form_id: 'form_123',
      user_id: 'user_456'
    }
  }));
};

socket.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
  
  // Handle different message types
  switch(message.type) {
    case 'user:joined':
      handleUserJoined(message.payload);
      break;
    case 'cursor:move':
      updateCursorPosition(message.payload);
      break;
    case 'question:update':
      updateQuestion(message.payload);
      break;
    // ... handle other events
  }
};

socket.onerror = function(error) {
  console.error('WebSocket error:', error);
};

socket.onclose = function(event) {
  console.log('Connection closed:', event.code, event.reason);
};
```

## 📨 Message Format

All WebSocket messages use the following JSON structure:

```json
{
  "type": "event_type",
  "payload": {
    // Event-specific data
  },
  "timestamp": 1672531200000,
  "user_id": "user_123",
  "message_id": "msg_abc123"
}
```

### Message Fields
- **type**: Event type identifier (required)
- **payload**: Event-specific data object (optional)
- **timestamp**: Unix timestamp in milliseconds (auto-generated)
- **user_id**: ID of the user sending the message (auto-populated)
- **message_id**: Unique message identifier (auto-generated)

## 📋 WebSocket Events

### 🏠 Room Management Events

#### `join:form`
User joins a form collaboration session.

**Send:**
```json
{
  "type": "join:form",
  "payload": {
    "form_id": "form_123",
    "user_id": "user_456"
  }
}
```

**Receive Confirmation:**
```json
{
  "type": "joined:form",
  "payload": {
    "form_id": "form_123",
    "user": {
      "user_id": "user_456",
      "username": "john.doe",
      "email": "john@example.com",
      "role": "editor"
    },
    "room_info": {
      "active_users": 3,
      "max_users": 10,
      "created_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "user:joined",
  "payload": {
    "form_id": "form_123",
    "user": {
      "user_id": "user_456",
      "username": "john.doe",
      "email": "john@example.com",
      "role": "editor"
    }
  }
}
```

#### `leave:form`
User leaves a form collaboration session.

**Send:**
```json
{
  "type": "leave:form",
  "payload": {
    "form_id": "form_123"
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "user:left",
  "payload": {
    "form_id": "form_123",
    "user_id": "user_456"
  }
}
```

### 🖱️ Cursor Tracking Events

#### `cursor:move`
User moves their cursor position.

**Send:**
```json
{
  "type": "cursor:move",
  "payload": {
    "x": 120.5,
    "y": 75.2,
    "element_id": "question_1",
    "viewport": {
      "width": 1920,
      "height": 1080
    }
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "cursor:update",
  "payload": {
    "user_id": "user_456",
    "username": "john.doe",
    "x": 120.5,
    "y": 75.2,
    "element_id": "question_1",
    "color": "#FF6B6B"
  }
}
```

#### `cursor:hide`
User hides their cursor (e.g., when inactive).

**Send:**
```json
{
  "type": "cursor:hide",
  "payload": {}
}
```

**Broadcast to Other Users:**
```json
{
  "type": "cursor:hidden",
  "payload": {
    "user_id": "user_456"
  }
}
```

### ❓ Question Management Events

#### `question:update`
User updates question content.

**Send:**
```json
{
  "type": "question:update",
  "payload": {
    "question_id": "q1",
    "content": "What is your favorite programming language?",
    "type": "multiple_choice",
    "options": [
      { "id": "opt1", "text": "JavaScript" },
      { "id": "opt2", "text": "Python" },
      { "id": "opt3", "text": "Go" }
    ],
    "required": true
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "question:updated",
  "payload": {
    "question_id": "q1",
    "content": "What is your favorite programming language?",
    "type": "multiple_choice",
    "options": [
      { "id": "opt1", "text": "JavaScript" },
      { "id": "opt2", "text": "Python" },
      { "id": "opt3", "text": "Go" }
    ],
    "required": true,
    "updated_by": {
      "user_id": "user_456",
      "username": "john.doe"
    }
  }
}
```

#### `question:focus`
User focuses on a specific question.

**Send:**
```json
{
  "type": "question:focus",
  "payload": {
    "question_id": "q1"
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "question:focused",
  "payload": {
    "question_id": "q1",
    "user": {
      "user_id": "user_456",
      "username": "john.doe",
      "color": "#FF6B6B"
    }
  }
}
```

#### `question:blur`
User unfocuses from a question.

**Send:**
```json
{
  "type": "question:blur",
  "payload": {
    "question_id": "q1"
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "question:blurred",
  "payload": {
    "question_id": "q1",
    "user_id": "user_456"
  }
}
```

### 💾 Form Management Events

#### `form:save`
User saves the form.

**Send:**
```json
{
  "type": "form:save",
  "payload": {
    "form_id": "form_123",
    "auto_save": false
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "form:saved",
  "payload": {
    "form_id": "form_123",
    "saved_by": {
      "user_id": "user_456",
      "username": "john.doe"
    },
    "saved_at": "2024-01-15T10:30:00Z",
    "version": "1.2.3"
  }
}
```

### ⌨️ Typing Indicators

#### `user:typing`
User starts typing in a question.

**Send:**
```json
{
  "type": "user:typing",
  "payload": {
    "element_id": "question_1"
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "user:typing",
  "payload": {
    "user_id": "user_456",
    "username": "john.doe",
    "element_id": "question_1"
  }
}
```

#### `user:stopped_typing`
User stops typing.

**Send:**
```json
{
  "type": "user:stopped_typing",
  "payload": {
    "element_id": "question_1"
  }
}
```

**Broadcast to Other Users:**
```json
{
  "type": "user:stopped_typing",
  "payload": {
    "user_id": "user_456",
    "element_id": "question_1"
  }
}
```

### 💓 Connection Management

#### `heartbeat`
Maintain connection and check server status.

**Send:**
```json
{
  "type": "heartbeat",
  "payload": {
    "client_time": 1672531200000
  }
}
```

**Receive:**
```json
{
  "type": "heartbeat:ack",
  "payload": {
    "server_time": 1672531201000,
    "latency": 15
  }
}
```

## ⚠️ Error Handling

### Error Message Format
```json
{
  "type": "error",
  "payload": {
    "code": "INVALID_FORM_ID",
    "message": "The specified form ID does not exist or you don't have access",
    "details": {
      "form_id": "invalid_form_123",
      "user_id": "user_456"
    }
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `AUTHENTICATION_FAILED` | Invalid or expired JWT token |
| `INVALID_FORM_ID` | Form ID doesn't exist or no access |
| `ROOM_FULL` | Maximum number of users reached |
| `INVALID_MESSAGE_FORMAT` | Malformed JSON message |
| `RATE_LIMIT_EXCEEDED` | Too many messages sent |
| `PERMISSION_DENIED` | User doesn't have required permissions |

## 🔒 Security Considerations

### Authentication
- All connections require valid JWT tokens
- Tokens are validated on connection and periodically refreshed
- Invalid tokens result in immediate connection termination

### Rate Limiting
- Maximum 100 messages per minute per user
- Cursor movement messages limited to 30 per second
- Heartbeat messages limited to 1 per 30 seconds

### Data Validation
- All incoming messages are validated against schemas
- Malformed messages are rejected with error responses
- User permissions are checked for each action

## 📊 Connection States

### Connection Lifecycle
1. **Connecting**: WebSocket handshake in progress
2. **Connected**: Authentication successful, ready for messages
3. **Joined**: User has joined a form collaboration session
4. **Disconnecting**: Connection closing gracefully
5. **Disconnected**: Connection terminated

### Reconnection Strategy
```javascript
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectDelay = 1000; // 1 second

function connect() {
  const socket = new WebSocket('ws://localhost:8080/api/v1/ws');
  
  socket.onopen = function() {
    reconnectAttempts = 0;
    console.log('Connected successfully');
  };
  
  socket.onclose = function(event) {
    if (reconnectAttempts < maxReconnectAttempts) {
      setTimeout(() => {
        reconnectAttempts++;
        console.log(`Reconnecting... (${reconnectAttempts}/${maxReconnectAttempts})`);
        connect();
      }, reconnectDelay * reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
    }
  };
}
```

## 🧪 Testing with JavaScript

### Complete Example Client
```javascript
class CollaborationClient {
  constructor(wsUrl, token) {
    this.wsUrl = wsUrl;
    this.token = token;
    this.socket = null;
    this.currentFormId = null;
    this.userId = null;
  }

  connect() {
    this.socket = new WebSocket(`${this.wsUrl}?token=${this.token}`);
    
    this.socket.onopen = (event) => {
      console.log('🔗 Connected to collaboration service');
    };

    this.socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.socket.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };

    this.socket.onclose = (event) => {
      console.log('🔌 Connection closed:', event.code, event.reason);
    };
  }

  joinForm(formId, userId) {
    this.currentFormId = formId;
    this.userId = userId;
    this.send('join:form', {
      form_id: formId,
      user_id: userId
    });
  }

  leaveForm() {
    if (this.currentFormId) {
      this.send('leave:form', {
        form_id: this.currentFormId
      });
      this.currentFormId = null;
    }
  }

  moveCursor(x, y, elementId) {
    this.send('cursor:move', {
      x: x,
      y: y,
      element_id: elementId
    });
  }

  updateQuestion(questionId, content, type, options) {
    this.send('question:update', {
      question_id: questionId,
      content: content,
      type: type,
      options: options
    });
  }

  send(type, payload) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({
        type: type,
        payload: payload
      }));
    } else {
      console.warn('⚠️ WebSocket not connected');
    }
  }

  handleMessage(message) {
    console.log('📨 Received:', message);
    
    switch (message.type) {
      case 'joined:form':
        console.log('✅ Joined form successfully:', message.payload);
        break;
      case 'user:joined':
        console.log('👤 User joined:', message.payload.user);
        break;
      case 'user:left':
        console.log('👋 User left:', message.payload.user_id);
        break;
      case 'cursor:update':
        console.log('🖱️ Cursor updated:', message.payload);
        break;
      case 'question:updated':
        console.log('❓ Question updated:', message.payload);
        break;
      case 'error':
        console.error('❌ Error:', message.payload);
        break;
      default:
        console.log('📭 Unknown message type:', message.type);
    }
  }

  disconnect() {
    if (this.socket) {
      this.leaveForm();
      this.socket.close();
    }
  }
}

// Usage example
const client = new CollaborationClient('ws://localhost:8080/api/v1/ws', 'your_jwt_token');
client.connect();

// Join a form
setTimeout(() => {
  client.joinForm('form_123', 'user_456');
}, 1000);

// Move cursor
setTimeout(() => {
  client.moveCursor(150, 200, 'question_1');
}, 2000);

// Update a question
setTimeout(() => {
  client.updateQuestion('q1', 'Updated question text', 'text', []);
}, 3000);
```

## 📈 Monitoring and Debugging

### WebSocket Connection Events
- Monitor connection events in browser developer tools
- Check Network tab for WebSocket frames
- Use console logs for message flow debugging

### Server-Side Logs
- Connection establishment and termination
- Authentication failures
- Rate limiting violations
- Message processing errors

### Health Check Endpoint
```bash
curl http://localhost:8080/api/v1/health
```

### Metrics Endpoint
```bash
curl http://localhost:8080/api/v1/metrics
```

---

## 📞 Support

For technical support or questions about the WebSocket API:

- **Email**: dev@xform.com
- **GitHub**: [X-Form-Backend Issues](https://github.com/Mir00r/X-Form-Backend/issues)
- **Documentation**: [Full API Documentation](http://localhost:8080/swagger/index.html)

---

*This documentation is for X-Form Collaboration Service v1.0.0*
