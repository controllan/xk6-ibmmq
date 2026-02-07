// test_ibmmq.js
import { connectQueueManager } from 'k6/x/ibmmq';
import { check } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

// Initialize queue manager in init context
// Use 'ibmmq-dev' as hostname when running inside Docker network
// Use 'localhost' when running from host machine
const mqHost = __ENV.MQ_HOST || 'ibmmq-dev';
const qm = connectQueueManager({
  connectionName: `${mqHost}(1414)`,
  queueManager: 'QM1',
  channel: 'DEV.APP.SVRCONN',
  // username: 'app',
  // password: 'password',
  // sslCipher: 'TLS_RSA_WITH_AES_256_CBC_SHA256', // Optional SSL
});

export default function () {
  // Put message with headers
  const messageData = JSON.stringify({
    timestamp: Date.now(),
    payload: 'Test message from k6',
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
    'message has data': (m) => m.data.length > 0,
    'message has headers': (m) => Object.keys(m.headers).length > 0,
  });
  
  if (message) {
    console.log(`Received message: ${String.fromCharCode.apply(null, message.data)}`);
    console.log(`Headers: ${JSON.stringify(message.headers)}`);
  }
}

export function teardown() {
  qm.Close();
}
