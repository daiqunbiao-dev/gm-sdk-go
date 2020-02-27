package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"os"

	//"strconv"
)

var (
	cc          = "lotuscc"
	user        = "admin"
	secret      = "adminpw"
	channelName = "gmchannel"
	lvl         = logging.INFO
)

func queryInstalledCC(sdk *fabsdk.FabricSDK) {
	userContext := sdk.Context(fabsdk.WithUser(user))

	resClient, err := resmgmt.New(userContext)
	if err != nil {
		fmt.Println("Failed to create resmgmt: ", err)
	}

	resp2, err := resClient.QueryInstalledChaincodes()
	if err != nil {
		fmt.Println("Failed to query installed cc: ", err)
	}
	fmt.Println("Installed cc: ", resp2.GetChaincodes())
}

func queryCC(client *channel.Client, name []byte) string {
	var queryArgs = [][]byte{name}
	response, err := client.Query(channel.Request{
		ChaincodeID: cc,
		Fcn:         "QueryUserinfo",
		Args:        queryArgs,
	})

	if err != nil {
		fmt.Println("Failed to query: ", err)
	}

	ret := string(response.Payload)
	fmt.Println("Chaincode status: ", response.ChaincodeStatus)
	fmt.Println("Payload: ", ret)
	return ret
}

func invokeCC(client *channel.Client, newValue string) {
	fmt.Println("Invoke cc with new value:", newValue)
	invokeArgs := [][]byte{[]byte("6"), []byte(newValue), []byte("13611266023"),[]byte("2020-02-20 15:04:05"), []byte(newValue), []byte("13611266023")}

	_, err := client.Execute(channel.Request{
		ChaincodeID: cc,
		Fcn:         "RegisterUser",
		Args:        invokeArgs,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}

/**
申请证书
 */
func enrollUser(sdk *fabsdk.FabricSDK) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}
	err = mspClient.Enroll(user, msp.WithSecret(secret))
	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		fmt.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			fmt.Printf("Failed to enroll user: %s\n", err)
		} else {
			fmt.Printf("Success enroll user: %s\n", user)
		}

	} else if err != nil {
		fmt.Printf("Failed to get user: %s\n", err)
	} else {
		fmt.Printf("User %s already enrolled, skip enrollment.\n", user)
	}
}

/**
注册用户
 */
func registerUser(sdk *fabsdk.FabricSDK) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}
	user := "daiqunbiao"
	secret := "123456"
	caName := "ca-LotusOrg1"
	org :="LotusOrg1"

	//注册用户
	regReq := &mspclient.RegistrationRequest{
		Name:           user,
		Secret:         secret,
		Type:           "client",
		MaxEnrollments: -1,
		Attributes:     []mspclient.Attribute{{Name: "role", Value: "client", ECert: true}},
		CAName:         caName,
		Affiliation:    org,
	}

	//用户注册
	_, err2 := mspClient.Register(regReq)

	if err2 !=nil {
		fmt.Print(err2)
	}

}

func queryChannelConfig(ledgerClient *ledger.Client) {
	resp1, err := ledgerClient.QueryConfig()
	if err != nil {
		fmt.Printf("Failed to queryConfig: %s", err)
	}
	fmt.Println("ChannelID: ", resp1.ID())
	fmt.Println("Channel Orderers: ", resp1.Orderers())
	fmt.Println("Channel Versions: ", resp1.Versions())
}

func queryChannelInfo(ledgerClient *ledger.Client) {
	resp, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Printf("Failed to queryInfo: %s", err)
	}
	fmt.Println("BlockChainInfo:", resp.BCI)
	fmt.Println("Endorser:", resp.Endorser)
	fmt.Println("Status:", resp.Status)
}

func setupLogLevel() {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}

func readInput() {
	if len(os.Args) != 5 {
		fmt.Printf("Usage: main.go <user-name> <user-secret> <channel> <chaincode-name>\n")
		os.Exit(1)
	}
	user = os.Args[1]
	secret = os.Args[2]
	channelName = os.Args[3]
	cc = os.Args[4]
}

func main() {
	/**
	readInput()
	*
	*/

	fmt.Println("Reading connection profile..")
	c := config.FromFile("./config.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	setupLogLevel()

	enrollUser(sdk)
	//registerUser(sdk)

	//创建通道的链接客户端
	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	fmt.Printf("\n===== Channel: %s ===== \n", channelName)

	queryChannelInfo(ledgerClient)
	queryChannelConfig(ledgerClient)

	fmt.Println("\n====== Chaincode =========")

	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:", channelName, err)
	}

	//invokeCC(client, "daiqunbiao")
	//old := queryCC(client, []byte("john"))

	//oldInt, _ := strconv.Atoi(old)
	//invokeCC(client, strconv.Itoa(oldInt+1))

	queryCC(client, []byte("6"))

	fmt.Println("===============")
	fmt.Println("Done.")

}
