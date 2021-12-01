package serializer

// Serializer interface is to provide serialize and deserialize methods for EasyRaft Node
type Serializer interface {
	// Serialize is used to serialize and data to a []byte
	Serialize(data interface{}) ([]byte, error)

	// Deserialize is used to deserialize []byte to interface{}
	Deserialize(data []byte) (interface{}, error)
}
