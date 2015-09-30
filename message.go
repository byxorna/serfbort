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
	//Action string `json:"action"` // Action is command to run (i.e. deploy, verify)
	Target string `json:"t"` // Target of the query or event (i.e.

	// Optional argument that will be parameterized into
	// the action's script (i.e. version, sha, etc)
	Argument string `json:"a"`
}

func DecodeMessagePayload(msg []byte) (MessagePayload, error) {
	mp := MessagePayload{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoderBytes(msg, &msgpack)
	err := decoder.Decode(&mp)
	if err != nil {
		return mp, err
	}

	return mp, nil
}

func (m MessagePayload) Encode() ([]byte, error) {
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

type QueryResponse struct {
	Output string //output of query command
	Err    error  //error, if any
}

func DecodeQueryResponse(msg []byte) (QueryResponse, error) {
	qr := QueryResponse{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoderBytes(msg, &msgpack)
	err := decoder.Decode(&qr)
	if err != nil {
		return qr, err
	}

	return qr, nil
}

func (q QueryResponse) Encode() ([]byte, error) {
	b := []byte{}
	msgpack.MapType = reflect.TypeOf(map[string]interface{}(nil))
	encoder := codec.NewEncoderBytes(&b, &msgpack)
	err := encoder.Encode(q)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (q QueryResponse) String() string {
	return fmt.Sprintf("output:%q error:%q", q.Output, q.Err)
}
