package main

import (
	"fmt"
	"github.com/ugorji/go/codec"
	"reflect"
)

var (
	msgpack codec.MsgpackHandle
)

type DeployMessage struct {
	RequiredTags map[string]string `json:"rt"`
	Target       string            `json:"t"`
	Argument     string            `json:"a"`
	Name         string            `json:"n"`
}

func decodeDeployMessage(msg []byte) (DeployMessage, error) {
	deployMessage := DeployMessage{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoderBytes(msg, &msgpack)
	err := decoder.Decode(&deployMessage)
	if err != nil {
		return deployMessage, err
	}

	return deployMessage, nil
}

func encodeDeployMessage(m DeployMessage) ([]byte, error) {
	b := []byte{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	encoder := codec.NewEncoderBytes(&b, &msgpack)
	err := encoder.Encode(m)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (m DeployMessage) String() string {
	return fmt.Sprintf("target:%s arg:%s tags:%v name:%s", m.Target, m.Argument, m.RequiredTags, m.Name)
}
