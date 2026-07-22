package homewizard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// P1Data holds the fields we care about from the HomeWizard P1 /api/v1/data response.
type P1Data struct {
	ActivePowerW   float64 `json:"active_power_w"`
	ActivePowerL1W float64 `json:"active_power_l1_w"`
	ActivePowerL2W float64 `json:"active_power_l2_w"`
	ActivePowerL3W float64 `json:"active_power_l3_w"`
}

// FetchP1Data makes a single GET request to the HomeWizard P1 meter and returns the parsed data.
func FetchP1Data(host string) (*P1Data, error) {
	if host == "" {
		return nil, fmt.Errorf("no host configured")
	}
	url := fmt.Sprintf("http://%s/api/v1/data", host)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting P1 data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from P1 meter", resp.StatusCode)
	}

	var data P1Data
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding P1 response: %w", err)
	}
	return &data, nil
}
