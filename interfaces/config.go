package interfaces

type Config interface {
	GetString(key string) string
}
