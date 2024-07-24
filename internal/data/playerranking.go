package data

type PlayerRank struct {
	CharacterName   string `json:"characterName"`
	CharacterImgURL string `json:"characterImgURL"`
	JobName         string `json:"jobName"`
	Level           int    `json:"level"`
}

type PlayerRankingResponse struct {
	TotalCount int          `json:"totalCount"`
	Ranks      []PlayerRank `json:"ranks"`
}
