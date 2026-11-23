package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func FetchCharacterData(name string, region string) (*data.PlayerRank, error) {
	if region != "na" && region != "eu" {
		region = "na"
	}
	req, err := http.NewRequest(http.MethodGet, "https://www.nexon.com/api/maplestory/ranking/v1/no-auth/"+region+"/overall", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("id", "weekly")
	q.Add("reboot_index", "0")
	q.Add("page_index", "1")
	q.Add("character_name", name)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("official rankings returned HTTP %d", resp.StatusCode)
	}
	body := data.PlayerRankingResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	if body.TotalCount != 1 || len(body.Ranks) != 1 || body.Ranks[0].CharacterName == "" {
		return nil, errors.New("character not found on official rankings")
	}

	return &body.Ranks[0], nil
}
