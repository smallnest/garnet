package garnet

type Codec interface {
	Encode(v interface{}) (interface{}, error)
	Decode(v interface{}) (interface{}, error)
}
