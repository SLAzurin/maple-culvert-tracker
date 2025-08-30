package data

type PlayerRank struct {
	CharacterName   string `json:"characterName"`
	CharacterImgURL string `json:"characterImgURL"`
	Level           int    `json:"level"`
	JobID           int    `json:"jobID"`
	JobDetail       int    `json:"jobDetail"`
}

type PlayerRankingResponse struct {
	TotalCount int          `json:"totalCount"`
	Ranks      []PlayerRank `json:"ranks"`
}
