package data

type WeeklySandbaggersStats = struct {
	Name                 string
	Score                int
	RawStats             *CharacterStatistics
	DiffPbPercentage     int
	DiffMedianPercentage int
}
