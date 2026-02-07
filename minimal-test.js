// minimal-test.js
import { connectQueueManager } from 'k6/x/ibmmq';

const mqHost = __ENV.MQ_HOST || 'ibmmq-dev';

export default function () {
  console.log(`Attempting to connect to ${mqHost}(1414)...`);
  
  try {
    const qm = connectQueueManager({
      connectionName: `${mqHost}(1414)`,
      queueManager: 'QM1',
      channel: 'DEV.APP.SVRCONN',
    });
    
    console.log('Connection successful!');
    qm.Close();
    console.log('Disconnected successfully!');
  } catch (e) {
    console.log(`Connection failed: ${e}`);
  }
}
