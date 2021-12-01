package fsm

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"sync"
)

type MapPutRequest struct {
	MapName string
	Key     string
	Value   interface{}
}

type MapGetRequest struct {
	MapName string
	Key     string
}

type MapRemoveRequest struct {
	MapName string
	Key     string
}

type Map struct {
	sync.RWMutex
	Data map[string]interface{}
}

type InMemoryMapService struct {
	sync.RWMutex
	Maps map[string]*Map
}

func NewInMemoryMapService() FSMService {
	return &InMemoryMapService{Maps: map[string]*Map{}}
}

func (m *InMemoryMapService) Name() string {
	return "in_memory_map"
}

func (m *InMemoryMapService) ApplySnapshot(input interface{}) error {
	var svc InMemoryMapService
	err := mapstructure.Decode(input, &svc)
	if err != nil {
		return err
	}
	m.Maps = svc.Maps
	return nil
}

func (m *InMemoryMapService) NewLog(requestType interface{}, request map[string]interface{}) interface{} {
	switch requestType.(type) {
	case MapPutRequest:
		var req MapPutRequest
		err := mapstructure.Decode(request, &req)
		if err != nil {
			return err
		}
		m.Put(req.MapName, req.Key, req.Value)
		return nil
	case MapGetRequest:
		var req MapGetRequest
		err := mapstructure.Decode(request, &req)
		if err != nil {
			return err
		}
		return m.Get(req.MapName, req.Key)
	case MapRemoveRequest:
		var req MapRemoveRequest
		err := mapstructure.Decode(request, &req)
		if err != nil {
			return err
		}
		m.Remove(req.MapName, req.Key)
		return nil
	default:
		return errors.New("unknown request type")
	}
}

func (m *InMemoryMapService) GetReqDataTypes() []interface{} {
	return []interface{}{MapPutRequest{}, MapGetRequest{}, MapRemoveRequest{}}
}

func (m *InMemoryMapService) Put(mapName string, key string, value interface{}) {
	m.Lock()
	defer m.Unlock()
	fMap, found := m.Maps[mapName]
	if !found {
		m.Maps[mapName] = &Map{Data: map[string]interface{}{}}
		fMap = m.Maps[mapName]
	}
	fMap.Lock()
	defer fMap.Unlock()
	fMap.Data[key] = value
}

func (m *InMemoryMapService) Get(mapName string, key string) interface{} {
	m.RLock()
	defer m.RUnlock()
	fMap, found := m.Maps[mapName]
	if !found {
		m.Maps[mapName] = &Map{Data: map[string]interface{}{}}
		fMap = m.Maps[mapName]
	}
	fMap.RLock()
	defer fMap.RUnlock()
	return fMap.Data[key]
}

func (m *InMemoryMapService) Remove(mapName string, key string) {
	m.Lock()
	defer m.Unlock()
	fMap, found := m.Maps[mapName]
	if !found {
		m.Maps[mapName] = &Map{Data: map[string]interface{}{}}
		fMap = m.Maps[mapName]
	}
	fMap.Lock()
	defer fMap.Unlock()
	delete(fMap.Data, key)
}
