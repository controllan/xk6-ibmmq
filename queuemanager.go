package ibmmq

import (
	"fmt"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

// QueueManagerConfig holds connection parameters
type QueueManagerConfig struct {
	ConnectionName string `json:"connectionName"` // host(port)
	QueueManager   string `json:"queueManager"`
	Channel        string `json:"channel"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	SSLCipher      string `json:"sslCipher,omitempty"`
}

// QueueManager wraps IBM MQ connection
type QueueManager struct {
	config *QueueManagerConfig
	qMgr   ibmmq.MQQueueManager
}

// NewQueueManager creates a new queue manager connection
func (mi *ModuleInstance) NewQueueManager(config QueueManagerConfig) (*QueueManager, error) {
	qm := &QueueManager{
		config: &config,
	}

	if err := qm.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return qm, nil
}

// connect establishes connection to IBM MQ
func (qm *QueueManager) connect() error {
	cno := ibmmq.NewMQCNO()
	csp := ibmmq.NewMQCSP()
	cd := ibmmq.NewMQCD()

	// Set connection details
	cd.ChannelName = qm.config.Channel
	cd.ConnectionName = qm.config.ConnectionName

	// SSL Configuration
	if qm.config.SSLCipher != "" {
		cd.SSLCipherSpec = qm.config.SSLCipher
		cno.ClientConn = cd
	} else {
		cno.ClientConn = cd
	}

	// Authentication
	if qm.config.Username != "" {
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = qm.config.Username
		csp.Password = qm.config.Password
		cno.SecurityParms = csp
	}

	cno.Options = ibmmq.MQCNO_CLIENT_BINDING

	// Connect to queue manager
	qMgr, err := ibmmq.Connx(qm.config.QueueManager, cno)
	if err != nil {
		return fmt.Errorf("MQCONNX failed: %w", err)
	}

	qm.qMgr = qMgr
	return nil
}

// Put writes a message to a queue
func (qm *QueueManager) Put(queueName string, message []byte, headers map[string]string) error {
	// Open queue for output
	mqod := ibmmq.NewMQOD()
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = queueName

	openOptions := ibmmq.MQOO_OUTPUT
	qObject, err := qm.qMgr.Open(mqod, openOptions)
	if err != nil {
		return fmt.Errorf("failed to open queue: %w", err)
	}
	defer qObject.Close(0)

	// Create message descriptor
	putmqmd := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT

	// Set message properties/headers if provided
	if len(headers) > 0 {
		pmho := ibmmq.NewMQPMHO()
		pd := ibmmq.NewMQPD()

		for key, value := range headers {
			pd.Options = ibmmq.MQPD_NONE
			pd.CopyOptions = ibmmq.MQCOPY_NONE

			// Set property
			err = qObject.SetMP(pmho, key, pd, value)
			if err != nil {
				return fmt.Errorf("failed to set property %s: %w", key, err)
			}
		}

		pmo.OriginalMsgHandle = pmho.NewMsgHandle
	}

	// Put message
	err = qObject.Put(putmqmd, pmo, message)
	if err != nil {
		return fmt.Errorf("failed to put message: %w", err)
	}

	return nil
}

// Get reads a message from a queue
func (qm *QueueManager) Get(queueName string, waitInterval int32) (*Message, error) {
	// Open queue for input
	mqod := ibmmq.NewMQOD()
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = queueName

	openOptions := ibmmq.MQOO_INPUT_AS_Q_DEF
	qObject, err := qm.qMgr.Open(mqod, openOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to open queue: %w", err)
	}
	defer qObject.Close(0)

	// Create message descriptor and get options
	getmqmd := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT | ibmmq.MQGMO_PROPERTIES_IN_HANDLE

	if waitInterval > 0 {
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.WaitInterval = waitInterval
	} else {
		gmo.Options |= ibmmq.MQGMO_NO_WAIT
	}

	// Create message handle for properties
	cmho := ibmmq.NewMQCMHO()
	msgHandle, err := qm.qMgr.CrtMH(cmho)
	if err == nil {
		gmo.MsgHandle = msgHandle
		defer qm.qMgr.DltMH(&msgHandle)
	}

	// Allocate buffer for message
	buffer := make([]byte, 100000) // 100KB max message size
	datalen, err := qObject.Get(getmqmd, gmo, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Extract headers/properties
	headers := make(map[string]string)
	if msgHandle.IsValid() {
		impo := ibmmq.NewMQIMPO()
		pd := ibmmq.NewMQPD()

		for {
			name, value, err := msgHandle.InqMP(impo, pd, "%")
			if err != nil {
				break // No more properties
			}
			headers[name] = fmt.Sprintf("%v", value)
		}
	}

	return &Message{
		Data:      buffer[:datalen],
		MessageID: getmqmd.MsgId,
		CorrelID:  getmqmd.CorrelId,
		Headers:   headers,
	}, nil
}

// Close disconnects from queue manager
func (qm *QueueManager) Close() error {
	if err := qm.qMgr.Disc(); err != nil {
		return fmt.Errorf("disconnect failed: %w", err)
	}
	return nil
}

// Message represents an IBM MQ message
type Message struct {
	Data      []byte            `json:"data"`
	MessageID []byte            `json:"messageId"`
	CorrelID  []byte            `json:"correlId"`
	Headers   map[string]string `json:"headers"`
}
