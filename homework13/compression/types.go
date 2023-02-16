package compression

type Compression interface {
	Zip([]byte) ([]byte, error)
	UnZip([]byte) ([]byte, error)
	Code() uint8
}
