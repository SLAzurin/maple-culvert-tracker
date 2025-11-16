package helpers

const SandbagThreshold float64 = .7

func GetSandbagThreshold(lastKnownGoodScore int64) int64 {
	threshold := int64(float64(lastKnownGoodScore) * SandbagThreshold)
	// if int64(lastKnownGoodScore)-threshold > data.MaxCulvertScoreThreshold {
	// 	threshold = lastKnownGoodScore - data.MaxCulvertScoreThreshold
	// }
	// removing this temporarily to test output, characters nowadays are giga strong

	return threshold
}
