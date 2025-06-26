package processors

import "cm_collectors_server/models"

type App struct {
	FilesBases     *[]models.FilesBases     `json:"filesBases"`
	PerformerBases *[]models.PerformerBases `json:"performerBases"`
}

func (App) InitData() (*App, error) {
	filesBases, err := FilesBases{}.DataList()
	if err != nil {
		return nil, err
	}
	performerBases, err := PerformerBases{}.DataList()
	if err != nil {
		return nil, err
	}
	return &App{
		FilesBases:     filesBases,
		PerformerBases: performerBases,
	}, nil
}
