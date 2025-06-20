package processors

type App struct {
	FilesBases *[]FilesBases `json:"filesBases"`
}

func (App) InitData() (*App, error) {
	filesBases, err := FilesBases{}.DataList()
	if err != nil {
		return nil, err
	}
	return &App{FilesBases: filesBases}, nil
}
