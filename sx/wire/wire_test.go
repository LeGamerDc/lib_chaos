package wire

import (
	"encoding/json"
	//"fmt"
	"reflect"
	"testing"
)

type Blob struct {
	X int
	Y string
}

func (b *Blob) XXX_Marshal(xxx []byte, _ bool) ([]byte, error) {
	j, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	xxx = append(xxx, j...)
	return xxx, nil
}
func (b *Blob) XXX_Size() int {
	j, _ := json.Marshal(b)
	return len(j)
}

var blob = Blob{
	X: 312,
	Y: "hello world",
}
var blobJson, _ = json.Marshal(&blob)

func TestEncodeCall(t *testing.T) {
	var n = SizeCall(12345, 833, &blob, []byte("hi"))
	var b = make([]byte, n)
	err := EncodeCall(b, 12345, 833, &blob, []byte("hi"))
	if err != nil {
		t.Error(err)
	}
	var msg = new(Msg)
	err = msg.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(msg, &Msg{
		Seq:   12345,
		Type:  CallType_Call,
		Api:   833,
		Data:  blobJson,
		Extra: []byte("hi"),
	}) {
		t.Error("unmarshal not equal")
	}
}

func TestEncodeOneWay(t *testing.T) {
	var n = SizeOneWay(12345, &blob, []byte("hello"))
	var b = make([]byte, n)
	err := EncodeOneWay(b, 12345, &blob, []byte("hello"))
	if err != nil {
		t.Error(err)
	}
	var msg = new(Msg)
	err = msg.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(msg, &Msg{
		Api:   12345,
		Type:  CallType_OneWay,
		Data:  blobJson,
		Extra: []byte("hello"),
	}) {
		t.Error("unmarshal not equal")
	}
}

func TestEncodeException(t *testing.T) {
	var n = SizeException(12345, 213)
	var b = make([]byte, n)
	err := EncodeException(b, 12345, 213)
	if err != nil {
		t.Error(err)
	}
	var msg = new(Msg)
	err = msg.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(msg, &Msg{
		Seq:     12345,
		Type:    CallType_Exception,
		ErrCode: 213,
	}) {
		t.Error("unmarshal not equal")
	}
}

func TestEncodeReply(t *testing.T) {
	var n = SizeReply(1234, &blob)
	var b = make([]byte, n)
	err := EncodeReply(b, 1234, &blob)
	if err != nil {
		t.Error(err)
	}
	var msg = new(Msg)
	err = msg.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(msg, &Msg{
		Seq:  1234,
		Type: CallType_Reply,
		Data: blobJson,
	}) {
		t.Error("unmarshal not equal")
	}
}

func TestEncodeReply2(t *testing.T) {
	var n = SizeReply(1234, nil)
	var b = make([]byte, n)
	err := EncodeReply(b, 1234, nil)
	if err != nil {
		t.Error(err)
	}
	var msg = new(Msg)
	err = msg.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(msg, &Msg{
		Seq:  1234,
		Type: CallType_Reply,
	}) {
		t.Error("unmarshal not equal")
	}
}
