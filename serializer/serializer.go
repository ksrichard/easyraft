package serializer

type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte) (interface{}, error)
}
