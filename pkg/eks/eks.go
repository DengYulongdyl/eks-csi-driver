package eks

import (
	eksapi "github.com/capitalonline/eks-csi-driver/pkg/api/eks"
	"github.com/capitalonline/eks-csi-driver/pkg/common/consts"
	"github.com/capitalonline/eks-csi-driver/pkg/utils"
	"github.com/capitalonline/eks-csi-driver/pkg/utils/profile"
	"net/http"
)

func CreateBlock(clusterId, nodeId string) (*eksapi.CreateBlockResponse, error) {
	credential := utils.NewCredential(consts.AccessKeyID, consts.AccessKeySecret)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = http.MethodPost
	client, _ := eksapi.NewClient(credential, consts.Region, cpf)
	request := eksapi.NewCreateBlockRequest()
	response, err := client.CreateBlock(request)
	if err != nil {
		return nil, err
	}
	return response, err
}
