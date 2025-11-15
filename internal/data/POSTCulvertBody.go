package data

type POSTCulvertBody struct {
	IsNew   bool   `json:"isNew"`
	Week    string `json:"week"`
	Payload []struct {
		CharacterID int64 `json:"character_id"`
		Score       int   `json:"score"`
	} `json:"payload"`
}
