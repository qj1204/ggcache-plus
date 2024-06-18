package distributekv

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte // 存储真实的缓存值
}

// Len returns the view's length
func (bv ByteView) Len() int {
	return len(bv.b)
}

// ByteSlice b是只读的，这里返回一个拷贝，防止缓存值被外部程序修改
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// String returns the data as a string, making a copy if necessary.
func (bv ByteView) String() string {
	return string(bv.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c

}
