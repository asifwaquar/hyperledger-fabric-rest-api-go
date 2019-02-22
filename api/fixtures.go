package api

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
)

// Create test channel, install and instantiate test chaincode
func (fsc *FabricSdkClient) InitTestFixturesHandler() error {

	testChannelId := "chainhero"
	testChaincodeName := "heroes-service"
	channelConfigTx := "./test/fixtures/artifacts/chainhero.channel.tx"
	testOrdererId :=  "orderer.hf.chainhero.io"

	ordererEndPoint := resmgmt.WithOrdererEndpoint(testOrdererId)

	response, _ := fsc.admin.QueryChannels(resmgmt.WithTargets(fsc.getFirstPeer()))
	channels := response.GetChannels()

	channelExist := false
	for _, ch := range channels {
		if ch.ChannelId == testChannelId {
			channelExist = true
		}
	}

	if !channelExist {
		req := resmgmt.SaveChannelRequest{ChannelID: testChannelId, ChannelConfigPath: channelConfigTx, SigningIdentities: []msp.SigningIdentity{fsc.adminIdentity}}
		txID, err := FscInstance.admin.SaveChannel(req, ordererEndPoint)
		if err != nil || txID.TransactionID == "" {
			return errors.WithMessage(err, "failed to save channel")
		}
		fmt.Println("Channel created")

		// Make admin user join the previously created channel
		if err = FscInstance.admin.JoinChannel(testChannelId, resmgmt.WithRetry(retry.DefaultResMgmtOpts), ordererEndPoint); err != nil {
			return errors.WithMessage(err, "failed to make admin join channel")
		}
		fmt.Println("Channel joined")
	}

	queryInstalledChaincodesResponse, _ := fsc.admin.QueryInstalledChaincodes(resmgmt.WithTargets(fsc.getFirstPeer()))
	installedChaincodes := queryInstalledChaincodesResponse.GetChaincodes()

	chaincodeExists := false
	for _, cc := range installedChaincodes {
		if string(cc.Name) == testChaincodeName {
			chaincodeExists = true
		}
	}

	if !chaincodeExists {

		// Create the chaincode package that will be sent to the peers
		ccPkg, err := gopackager.NewCCPackage(fsc.ChaincodePath, fsc.GOPATH)
		if err != nil {
			return errors.WithMessage(err, "failed to create chaincode package")
		}
		fmt.Println("ccPkg created")

		// Install example cc to org peers
		installCCReq := resmgmt.InstallCCRequest{Name: testChaincodeName, Path: fsc.ChaincodePath, Version: "0", Package: ccPkg}
		_, err = fsc.admin.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return errors.WithMessage(err, "failed to install chaincode")
		}
		fmt.Println("Chaincode installed")

		// Set up chaincode policy
		ccPolicy := cauthdsl.SignedByAnyMember([]string{"org1.hf.chainhero.io"})

		resp, err := fsc.admin.InstantiateCC(testChannelId, resmgmt.InstantiateCCRequest{Name: testChaincodeName, Path: fsc.GOPATH, Version: "0", Args: [][]byte{[]byte("init")}, Policy: ccPolicy})
		if err != nil || resp.TransactionID == "" {
			return errors.WithMessage(err, "failed to instantiate the chaincode")
		}
		fmt.Println("Chaincode instantiated")
	}

	return nil
}