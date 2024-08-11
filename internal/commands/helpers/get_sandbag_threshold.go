package helpers

import "github.com/slazurin/maple-culvert-tracker/internal/data"

func GetSandbagThreshold(lastKnownGoodScore int64) int64 {
	threshold := int64(float64(lastKnownGoodScore) * .7)
	if int64(lastKnownGoodScore)-threshold > data.MaxCulvertScoreThreshold {
		threshold = lastKnownGoodScore - data.MaxCulvertScoreThreshold
	}

	return threshold
}