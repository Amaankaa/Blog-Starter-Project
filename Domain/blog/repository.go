package blogpkg

// BlogRepository interface defines the methods
type IBlogRepository interface {
	CreateBlog(blog *Blog) (*Blog, error)
}
