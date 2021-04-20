package hasher

type Hash interface {
	Hash(values ...interface{}) string
}
