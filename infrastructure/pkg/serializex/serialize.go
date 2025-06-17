package serializex

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
)

func Unmarshal(msgType string, message []byte, req interface{}) error {
	switch msgType {
	case "json":
		return json.Unmarshal(message, req)
	case "proto":
		return proto.Unmarshal(message, req.(proto.Message))
	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}
}

func Marshal(msgType string, resp interface{}) ([]byte, error) {
	switch msgType {
	case "json":
		return json.Marshal(resp)
	case "proto":
		return proto.Marshal(resp.(proto.Message))
	default:
		return nil, fmt.Errorf("unsupported message type: %s", msgType)
	}
}
