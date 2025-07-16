package serverfilemanagement

func NewServerFileManagement(rootPaths []ServerFileManagement_PathEntry) *ServerFileManagement {
	sfm := &ServerFileManagement{
		RootPath: rootPaths,
	}
	sfm.preprocessRoots()
	return sfm
}
