package migrator

type Entity interface {
	// ID 要求返回ID
	ID() int64
	// CompareTo dst必然也是Entity
	CompareTo(dst Entity) bool
}
