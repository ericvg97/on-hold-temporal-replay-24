package onhold

import (
	"encoding/json"

	customeseachattributes "github.com/ericvg97/temporal-replay/cmd/customsearchattributes"
	"go.temporal.io/sdk/workflow"
)

func ExecuteWrappedActivity[T any](ctx workflow.Context, c func() (T, error)) T {
	output, err := c()
	if err == nil {
		return output
	}

	return HandleFailure[T](ctx, c)
}

func HandleFailure[T any](ctx workflow.Context, c func() (T, error)) T {
	isRetry, skipObject := WaitForSignals(ctx)
	if isRetry {
		ExecuteWrappedActivity(ctx, c)
	}

	var zeroValue T
	if skipObject == "" {
		return zeroValue
	}

	err := json.Unmarshal([]byte(skipObject), &zeroValue)
	if err != nil {
		return HandleFailure[T](ctx, c)
	}
	return zeroValue
}

func WaitForSignals(ctx workflow.Context) (bool, string) {
	customeseachattributes.SetOnHold(ctx)
	isRetry := false
	skipObject := ""
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(workflow.GetSignalChannel(ctx, "retry"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
		isRetry = true
	})
	selector.AddReceive(workflow.GetSignalChannel(ctx, "skip"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &skipObject)

		isRetry = false
	})
	selector.Select(ctx)
	customeseachattributes.SetRunning(ctx)

	return isRetry, skipObject
}
