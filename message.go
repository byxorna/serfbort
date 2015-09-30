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
	Status int    //exit status of query command
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
	return fmt.Sprintf("output:%q status:%d", q.Output, q.Status)
}
