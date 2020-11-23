package models

// Model Interface all Zale Models should implement
type APIObject interface {
	JSON() ([]byte, error)
	Valid() error
	Decode(data []byte) (err error)
}
