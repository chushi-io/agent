package agent

import (
	"encoding/json"
	"fmt"
	"github.com/chushi-io/agent/types"
	"github.com/dghubble/sling"
)

type ChangeSink struct {
	Sdk   *sling.Sling
	RunId string
}

type ChangeSummary struct {
	Type    string              `json:"type"`
	Changes ChangeSummaryChange `json:"changes"`
}

type ChangeSummaryChange struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Remove    int    `json:"remove"`
	Import    int    `json:"import"`
	Operation string `json:"operation"`
}

func (sink ChangeSink) Write(p []byte) (int, error) {
	var summary ChangeSummary
	if err := json.Unmarshal(p, &summary); err != nil {
		return 0, err
	}

	if summary.Type != "change_summary" {
		return 0, nil
	}

	update := &types.RunStatusUpdate{
		Add:     summary.Changes.Add,
		Change:  summary.Changes.Change,
		Destroy: summary.Changes.Remove,
	}
	_, err := sink.Sdk.Put(fmt.Sprintf("agents/v1/runs/%s", sink.RunId)).BodyJSON(update).ReceiveSuccess(nil)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
