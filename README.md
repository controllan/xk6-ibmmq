# xk6-ibmmq

Grafana k6 extension for IBM MQ (IBM WebSphere MQ) load testing. This extension allows you to put and get messages from IBM MQ queues to implement comprehensive load tests using Grafana k6.

## Features

- **Put Messages**: Send messages to IBM MQ queues with custom headers/properties
- **Get Messages**: Retrieve messages from queues with configurable wait intervals
- **Message Properties**: Support for IBM MQ message properties (headers)
- **Connection Management**: Secure connection with username/password authentication and SSL/TLS support
- **Load Testing**: Full integration with k6 for performance and load testing

## Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local builds)
- IBM MQ client libraries (automatically installed via Docker)

## Quick Start

### 1. Start IBM MQ (Local Development)

Use the provided Docker Compose setup to run a local IBM MQ instance:

```bash
# Start IBM MQ container
docker compose up -d

# Wait for the queue manager to be ready (about 30-60 seconds)
docker compose ps
```

This will start:
- Queue Manager: `QM1`
- Queue: `DEV.QUEUE.1`
- Channel: `DEV.APP.SVRCONN` (with authentication disabled for development)
- Port: `1414`

### 2. Build the Extension

Use the provided build script:

```bash
# Make the script executable and run it
chmod +x build.sh
./build.sh
```

This creates a custom `k6` binary with the IBM MQ extension.

### 3. Run Tests

```bash
# Run the example test
./k6 run test_ibmmq.js

# Or use the provided script
chmod +x run-test.sh
./run-test.sh
```

## Manual Build

If you prefer to build manually:

```bash
# Install xk6
go install go.k6.io/xk6/cmd/xk6@latest

# Build with the extension
xk6 build --with github.com/controllan/xk6-ibmmq@latest

# Or build from local source
xk6 build --with github.com/controllan/xk6-ibmmq=.
```

**Note**: Building requires IBM MQ client libraries. See the Dockerfile for installation details.

## Configuration

### Connection Parameters

```javascript
import { connectQueueManager } from 'k6/x/ibmmq';

const qm = connectQueueManager({
  connectionName: 'localhost(1414)',  // host(port) format
  queueManager: 'QM1',                 // Queue manager name
  channel: 'DEV.APP.SVRCONN',         // Server connection channel
  username: 'app',                     // Optional: username for authentication
  password: 'password',               // Optional: password for authentication
  sslCipher: 'TLS_RSA_WITH_AES_256_CBC_SHA256'  // Optional: SSL cipher spec
});
```

### Environment Variables

- `MQ_HOST`: MQ hostname (default: `ibmmq-dev` when using Docker Compose)

### Docker Compose Configuration

The `docker-compose.yml` configures IBM MQ with:
- Queue Manager: `QM1`
- Application Channel: `DEV.APP.SVRCONN`
- Pre-configured queues: `DEV.QUEUE.1`, `DEV.QUEUE.2`, `DEV.QUEUE.3`
- Web Console available at: `https://localhost:9444/ibmmq/console/`
  - Username: `admin`
  - Password: `admin123`

## API Reference

### connectQueueManager(config)

Creates a connection to an IBM MQ queue manager.

**Parameters:**
- `config` (object): Connection configuration
  - `connectionName` (string): Connection name in `host(port)` format
  - `queueManager` (string): Queue manager name
  - `channel` (string): Server connection channel name
  - `username` (string, optional): Username for authentication
  - `password` (string, optional): Password for authentication
  - `sslCipher` (string, optional): SSL cipher specification

**Returns:** QueueManager instance

### QueueManager.Put(queueName, message, headers)

Puts a message to a queue.

**Parameters:**
- `queueName` (string): Name of the target queue
- `message` (string): Message payload
- `headers` (object, optional): Message properties/headers as key-value pairs

**Returns:** `null` on success, error on failure

### QueueManager.Get(queueName, waitInterval)

Gets a message from a queue.

**Parameters:**
- `queueName` (string): Name of the source queue
- `waitInterval` (number): Wait interval in milliseconds (0 for no wait)

**Returns:** Message object or `null` if no message available

**Message object structure:**
```javascript
{
  data: ArrayBuffer,        // Message payload
  messageId: ArrayBuffer,   // Message ID
  correlId: ArrayBuffer,    // Correlation ID
  headers: object           // Message properties
}
```

### QueueManager.Close()

Closes the connection to the queue manager.

**Returns:** `null` on success, error on failure

## Example Test Script

```javascript
import { connectQueueManager } from 'k6/x/ibmmq';
import { check } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

// Initialize queue manager in init context
const mqHost = __ENV.MQ_HOST || 'localhost';
const qm = connectQueueManager({
  connectionName: `${mqHost}(1414)`,
  queueManager: 'QM1',
  channel: 'DEV.APP.SVRCONN',
  // Optional: username and password for authenticated connections
  // username: 'app',
  // password: 'password',
});

export default function () {
  // Put message with custom headers
  const messageData = JSON.stringify({
    timestamp: Date.now(),
    payload: 'Test message from k6',
    vu: __VU,
    iteration: __ITER
  });
  
  const headers = {
    'X-Custom-Header': 'test-value',
    'X-VU': `${__VU}`,
    'X-Iteration': `${__ITER}`,
  };
  
  const putResult = qm.Put('DEV.QUEUE.1', messageData, headers);
  check(putResult, {
    'message sent successfully': (r) => r === null,
  });
  
  // Get message from queue (wait up to 5 seconds)
  const message = qm.Get('DEV.QUEUE.1', 5000);
  check(message, {
    'message received': (m) => m !== null,
    'message has data': (m) => m && m.data.length > 0,
    'message has headers': (m) => m && Object.keys(m.headers).length > 0,
  });
  
  if (message) {
    // Convert ArrayBuffer to string
    const decoder = new TextDecoder();
    const messageText = decoder.decode(message.data);
    console.log(`Received message: ${messageText}`);
    console.log(`Headers: ${JSON.stringify(message.headers)}`);
  }
}

export function teardown() {
  qm.Close();
}
```

## Docker Setup Details

### Dockerfile

The `Dockerfile` includes:
- Debian 13 slim base image
- Go 1.24.11
- IBM MQ Client libraries (version 9.4.4.1)
- xk6 build tool

### IBM MQ Container Configuration

The `mq-config.mqsc` file customizes the MQ configuration:
- Disables channel authentication for development
- Makes client authentication optional
- Configures the DEV.APP.SVRCONN channel
- Starts the channel automatically

### Stopping the Environment

```bash
# Stop IBM MQ
docker compose down

# Stop and remove volumes (data will be lost)
docker compose down -v
```

## Troubleshooting

### Connection Errors

**Error: `MQRC_CHANNEL_CONFIG_ERROR [2539]`**

This usually means the channel is not running or authentication is failing:

```bash
# Check channel status
docker exec ibmmq-dev bash -c "echo 'DISPLAY CHSTATUS(DEV.APP.SVRCONN)' | runmqsc QM1"

# Start the channel if needed
docker exec ibmmq-dev bash -c "echo 'START CHANNEL(DEV.APP.SVRCONN)' | runmqsc QM1"
```

**Error: `MQRC_CHANNEL_NOT_AVAILABLE [2537]`**

The channel is not enabled. Start it manually:

```bash
docker exec ibmmq-dev bash -c "echo 'START CHANNEL(DEV.APP.SVRCONN)' | runmqsc QM1"
```

**Error: `MQRC_NOT_AUTHORIZED [2035]`**

Authentication is enabled but credentials are incorrect or missing. Either:
1. Provide correct username/password in the connection config
2. Disable authentication in the MQ configuration (see `mq-config.mqsc`)

### Build Errors

**Missing IBM MQ headers (`cmqc.h`)**

Ensure the Dockerfile properly installs IBM MQ client libraries. The libraries must be in:
- `/opt/mqm/lib64`
- `/opt/mqm/inc`

### Performance Issues

- Use `teardown()` function to properly close connections
- Consider connection pooling for high-throughput scenarios
- Monitor MQ channel status and queue depth during tests

## IBM MQ Resources

- [IBM MQ Documentation](https://www.ibm.com/docs/en/ibm-mq)
- [IBM MQ Docker Image](https://github.com/ibm-messaging/mq-container)
- [mq-golang library](https://github.com/ibm-messaging/mq-golang)

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

Contributions are welcome! Please ensure:
1. Code follows Go best practices
2. Tests pass successfully
3. Documentation is updated for new features

## Support

For issues and questions:
- Open an issue on GitHub
- Check existing issues for solutions
- Refer to IBM MQ documentation for MQI-specific questions
