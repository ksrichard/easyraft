package serializer

import (
	"github.com/vmihailenco/msgpack/v5"
)

func NewMsgPackSerializer() Serializer {
	return &ConverterService{}
}

type ConverterService struct {
}

func (s *ConverterService) Serialize(data interface{}) ([]byte, error) {
	return msgpack.Marshal(data)
}

func (s *ConverterService) Deserialize(data []byte) (interface{}, error) {
	var result interface{}
	err := msgpack.Unmarshal(data, &result)
	return result, err
}
