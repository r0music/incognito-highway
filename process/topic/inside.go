package topic

import (
	"fmt"
	"highway/common"
)

// InsideTopic is topic inside a committee, using for communicate between proxy node and INC node
// MessageType is defined as type of message
// CommitteeID is defined as type of committee (0->254: shardID, 255: beacon)
// SelfID (optional) is ID of Proxy node
type InsideTopic struct {
	MessageType string
	CommitteeID byte
	SelfID      string
}

func (topic *InsideTopic) ToString() string {
	return fmt.Sprintf("%s-%x-%s", topic.MessageType, topic.CommitteeID, topic.SelfID)
}

func (topic *InsideTopic) FromMessageType(
	validator string,
	messageType string,
) error {
	// committeeID := common.GetCommitteeIDOfValidator(validator)
	// if (committeeID) < 0 {
	// 	return errors.New("Candidate not found")
	// }
	// topic.CommitteeID = byte(committeeID)
	topic.CommitteeID = byte(0x00)
	if isBroadcastMessage(messageType) {
		topic.SelfID = ""
	} else {
		topic.SelfID = common.SelfID
	}
	//Validate correctness of messageType
	topic.MessageType = messageType
	return nil
}
