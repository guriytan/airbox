package hasher

type Hasher interface {
	Hash(values ...interface{}) string
}
