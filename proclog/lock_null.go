package proclog

type NullLocker struct{}

func (that *NullLocker) Lock()   { return }
func (that *NullLocker) Unlock() { return }

func NewNullLocker() *NullLocker {
	return &NullLocker{}
}
