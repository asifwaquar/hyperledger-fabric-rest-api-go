package api

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (fsc *FabricSdkClient) Invoke(channelId, chaincodeId, fcn string, args []string) (string, error) {

	// Prepare arguments
	requestArgs := [][]byte{[]byte(fcn)}
	for _, arg := range args {
		requestArgs = append(requestArgs, []byte(arg))
	}


	// Add data that will be visible in the proposal, like a description of the invoke request
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in hello invoke")

	// TODO txStatus event support
	/*_, txStatusEventNotifier, err := eventClient.RegisterTxStatusEvent()
		if err != nil {
			return err
		}

		txStatusEventListener := func() {
			for {
				txEvent := <-txStatusEventNotifier
				notifications.HandleTxStatusEvent(txEvent)
			}
		}
		go txStatusEventListener()*/

	client, err := fsc.channelClient(channelId)
	if err != nil {
		return "", err
	}

	// Create a request (proposal) and send it
	response, err := client.Execute(channel.Request{ChaincodeID: chaincodeId, Fcn: "invoke", Args: requestArgs, TransientMap: transientDataMap})
	if err != nil {
		return "", fmt.Errorf("failed to move funds: %v", err)
	}

	return string(response.TransactionID), nil
}