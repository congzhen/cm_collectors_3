package datatype

const V_Search_Not string = "not"
const V_Search_Before2000 string = "before_2000"

type ReqParam_SearchData struct {
	SearchTextSlc []string                 `json:"searchTextSlc"`
	Sort          E_searchSort             `json:"sort"`
	Country       I_searchGroup            `json:"country"`
	Definition    I_searchGroup            `json:"definition"`
	Year          I_searchGroup            `json:"year"`
	Star          I_searchGroup            `json:"star"`
	Performer     I_searchGroup            `json:"performer"`
	Cup           I_searchGroup            `json:"cup"`
	Tag           map[string]I_searchGroup `json:"tag"`
}

type E_searchSort string

const (
	E_searchSort_addTimeDesc     E_searchSort = "addTimeDesc"
	E_searchSort_addTimeAsc      E_searchSort = "addTimeAsc"
	E_searchSort_issueNumberDesc E_searchSort = "issueNumberDesc"
	E_searchSort_issueNumberAsc  E_searchSort = "issueNumberAsc"
	E_searchSort_scoreDesc       E_searchSort = "scoreDesc"
	E_searchSort_scoreAsc        E_searchSort = "scoreAsc"
	E_searchSort_starDesc        E_searchSort = "starDesc"
	E_searchSort_starAsc         E_searchSort = "starAsc"
	E_searchSort_issuingDateDesc E_searchSort = "issuingDateDesc"
	E_searchSort_issuingDateAsc  E_searchSort = "issuingDateAsc"
	E_searchSort_titleDesc       E_searchSort = "titleDesc"
	E_searchSort_titleAsc        E_searchSort = "titleAsc"
	E_searchSort_history         E_searchSort = "history"
	E_searchSort_hot             E_searchSort = "hot"
	E_searchSort_youLike         E_searchSort = "youLike"
)

type I_searchGroup struct {
	Logic   E_searchLogic `json:"logic"`
	Options []string      `json:"options"`
}

type E_searchLogic string

const (
	E_searchLogic_single   E_searchLogic = "single"
	E_searchLogic_multiAnd E_searchLogic = "multiAnd"
	E_searchLogic_multiOr  E_searchLogic = "multiOr"
	E_searchLogic_not      E_searchLogic = "not"
)
