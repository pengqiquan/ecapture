// Copyright © 2022 Hengqi Chen
package user

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type inner struct {
	TimestampNS uint64
	Pid         uint32
	Tid         uint32
	Len         int32
	Comm        [16]byte
	Data        [4096]byte
}

type goSSLEvent struct {
	m IModule
	inner
}

func (e *goSSLEvent) Decode(payload []byte) error {
	r := bytes.NewBuffer(payload)
	return binary.Read(r, binary.LittleEndian, &e.inner)
}

func (e *goSSLEvent) String() string {
	s := fmt.Sprintf("PID: %d, Comm: %s, TID: %d, Payload: %s\n", e.Pid, string(e.Comm[:]), e.Tid, string(e.Data[:e.Len]))
	return s
}

func (e *goSSLEvent) StringHex() string {
	return e.String()
}

func (e *goSSLEvent) Clone() IEventStruct {
	return &goSSLEvent{}
}

func (e *goSSLEvent) Module() IModule {
	return e.m
}

func (e *goSSLEvent) SetModule(m IModule) {
	e.m = m
}

func (e *goSSLEvent) EventType() EVENT_TYPE {
	return EVENT_TYPE_OUTPUT
}
