package sdkInit

import (
    "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
    "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
    "fmt"
    "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
    mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
)

const ChaincodeVersion = "1.0"

func SetupSDK (ConfigFile string, initialized bool) (*fabsdk.FabricSDK, error) {
    if initialized {
        return nil, fmt.Errorf("Fabric-SDK已被实例化")
    }

    sdk, err := fabsdk.New(config.FromFile(ConfigFile))
    if err != nil {
        return nil, fmt.Errorf("实例化Fabric-SDK失败: %v", err)
    }

    fmt.Println("Fabric-SDK初始化成功")
    return sdk, nil
}

func CreateChannel (sdk *fabsdk.FabricSDK, info *InitInfo) error {
    clientContext := sdk.Context(fabsdk.WithUser(info.OrgAdmin), fabsdk.WithOrg(info.OrgName)) 
    if clientContext == nil {
        return fmt.Errorf("根据指定的组织名称与管理员创建资源管理客户端Context失败")
    }

    resMgmtClient, err := resmgmt.New(clientContext)
    if err != nil {
        return fmt.Errorf("根据指定的资源管理客户端Context创建通道管理客户端失败: %v", err)
    }

    mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(info.OrgName))
    if err != nil {
        return fmt.Errorf("根据指定的OrgName创建OrgMSP客户端实例失败: %v", err)
    }

    adminIdentity, err := mspClient.GetSigningIdentity(info.OrgAdmin)
    if err != nil {
        return fmt.Errorf("获取指定id的签名标识失败: %v", err)
    }

    channelReq := resmgmt.SaveChannelRequest {
        ChannelID:info.ChannelID,
        ChannelConfigPath:info.ChannelConfig,
        SigningIdentities:[]msp.SigningIdentity{adminIdentity}}

    _, err = resMgmtClient.SaveChannel(channelReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(info.OrdererOrgName)) 
    if err != nil {
        return fmt.Errorf("创建应用通道失败: %v", err)
    }

    fmt.Println("通道已成功创建，")
    info.OrgResMgmt = resMgmtClient

    err = info.OrgResMgmt.JoinChannel(info.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(info.OrdererOrgName))
    if err != nil {
        return fmt.Errorf("peers加入通道失败: %v", err)
    }

    fmt.Println("Peers已成功加入通道.")
    return nil
}



