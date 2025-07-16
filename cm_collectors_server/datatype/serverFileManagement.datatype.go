package datatype

type ServerFileManagement_PathFiles struct {
	Path string `json:"path"`
}
type ServerFileManagement_SearchFiles struct {
	Path        string `json:"path"`
	SearchQuery string `json:"search"`
}
type ServerFileManagement_CreateFile struct {
	Name string `json:"name"`
	Path string `json:"path" binding:"required"`
}

type ServerFileManagement_CreateFolder struct {
	Name        string `json:"name"`
	Path        string `json:"path" binding:"required"`
	Permissions string `json:"permissions"  binding:"required"`
}

type ServerFileManagement_OpenFile struct {
	FilePath   string `json:"filePath" binding:"required"`
	ReturnType string `json:"returnType"`
}
type ServerFileManagement_SaveFile struct {
	FilePath string `json:"filePath" binding:"required"`
	Content  string `json:"content"`
}

type ServerFileManagement_Action struct {
	Path  string   `json:"path" binding:"required"`
	Files []string `json:"files" binding:"required"`
}

type ServerFileManagement_Permissions struct {
	Files       []string `json:"files" binding:"required"`
	Permissions string   `json:"permissions"  binding:"required"`
	SubFiles    bool     `json:"sub_files"`
}

type ServerFileManagement_RenameFile struct {
	Name string `json:"name"`
	Path string `json:"path" binding:"required"`
}

type ServerFileManagement_UnCompressFile struct {
	File string `json:"file"`
}
