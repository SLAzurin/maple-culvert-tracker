package data

type ChartMakerPoints = struct {
	Label string `json:"label"`
	Score int    `json:"score"`
}

type ChartMakeMultiplePoints = struct {
	Labels    []string   `json:"labels"`
	DataPlots []DataPlot `json:"dataPlots"`
}

type DataPlot struct {
	CharacterName string `json:"characterName"`
	Scores        []int  `json:"scores"`
	
}
