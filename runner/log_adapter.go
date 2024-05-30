package runner

import (
	"fmt"
	"github.com/dghubble/sling"
	"strings"
)

type logAdapter struct {
	httpClient *sling.Sling
	RunId      string
	output     [][]byte
}

func newLogAdapter(client *sling.Sling, runId string) *logAdapter {
	adapter := &logAdapter{httpClient: client, RunId: runId}
	return adapter
}

func (adapter *logAdapter) Write(p []byte) (n int, err error) {
	adapter.output = append(adapter.output, p)
	return len(p), nil
	//_, err = adapter.Logs.StreamLogs(context.TODO(), connect.NewRequest(&v1.StreamLogsRequest{Content: string(p)}))
	//if err != nil {
	//	return 0, err
	//}
	//return len(p), nil
}

func (adapter *logAdapter) Flush() error {
	var lines []string
	for _, log := range adapter.output {
		lines = append(lines, string(log))
	}

	_, err := adapter.httpClient.
		Post(fmt.Sprintf("/api/v1/plans/%s/logs", adapter.RunId)).
		Body(strings.NewReader(strings.Join(lines, "\n"))).
		ReceiveSuccess(nil)
	return err
}
