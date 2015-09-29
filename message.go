package main

import (
	"fmt"
	"github.com/ugorji/go/codec"
	"reflect"
)

var (
	msgpack codec.MsgpackHandle
)

type MessagePayload struct {
	Target   string `json:"t"` // Target of the query or event
	Argument string `json:"a"` // Optional argument (i.e. version, sha, etc)
}

func decodeMessagePayload(msg []byte) (MessagePayload, error) {
	MessagePayload := MessagePayload{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoderBytes(msg, &msgpack)
	err := decoder.Decode(&MessagePayload)
	if err != nil {
		return MessagePayload, err
	}

	return MessagePayload, nil
}

func encodeMessagePayload(m MessagePayload) ([]byte, error) {
	b := []byte{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	encoder := codec.NewEncoderBytes(&b, &msgpack)
	err := encoder.Encode(m)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (m MessagePayload) String() string {
	return fmt.Sprintf("target:%s arg:%s", m.Target, m.Argument)
}
