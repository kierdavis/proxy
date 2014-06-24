package proxy

import (
	"log"
	"reflect"
)

type handlerManager struct {
	types map[PacketID]reflect.Type
	handlers map[PacketID][]reflect.Value
}

func newHandlerManager() (hm *handlerManager) {
	return &handlerManager{
		types: make(map[PacketID]reflect.Type),
		handlers: make(map[PacketID][]reflect.Value),
	}
}

func (hm *handlerManager) Add(handler interface{}) {
	// error checking!!!
	
	v := reflect.ValueOf(handler)
	packetType := v.Type().In(1).Elem()
	id := reflect.New(packetType).Interface().(Packet).ID()
	
	hm.types[id] = packetType
	hm.handlers[id] = append(hm.handlers[id], v)
	
	log.Printf("Registered handler for %s", id.String())
}

func (hm *handlerManager) Lookup(id PacketID) (packet Packet) {
	t, ok := hm.types[id]
	if ok {
		return reflect.New(t).Interface().(Packet)
	}
	return nil
}

func (hm *handlerManager) Process(session *Session, packet Packet) (accept bool) {
	defer func() {
		if x := recover(); x != nil {
			log.Printf("Panic caught when handling %s packet: %v", packet.ID().String(), x)
		}
	}()
	
	accept = true
	
	s := reflect.ValueOf(session)
	v := reflect.ValueOf(packet)
	handlers := hm.handlers[packet.ID()]
	
	for _, handler := range handlers {
		ins := []reflect.Value{s, v}
		outs := handler.Call(ins)
		a := outs[0].Bool()
		
		accept = accept && a
	}
	
	return accept
}
