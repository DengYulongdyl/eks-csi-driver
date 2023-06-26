package eks

import (
	"github.com/capitalonline/eks-csi-driver/pkg/common/consts"
	"github.com/capitalonline/eks-csi-driver/pkg/utils"
	cdshttp "github.com/capitalonline/eks-csi-driver/pkg/utils/http"
	"github.com/capitalonline/eks-csi-driver/pkg/utils/profile"
)

type Client struct {
	utils.Client
}

func NewClient(credential *utils.Credential, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewCreateBlockRequest() (request *CreateBlockRequest) {
	request = &CreateBlockRequest{
		BaseRequest: &cdshttp.BaseRequest{},
	}
	request.SetDomain(consts.ApiHost)
	request.Init().WithApiInfo(consts.ServiceEKS, consts.ApiVersion, consts.ActionCreateBlock)
	return
}

func NewCreateBlockResponse() (response *CreateBlockResponse) {
	response = &CreateBlockResponse{BaseResponse: &cdshttp.BaseResponse{}}
	return
}

func (c *Client) CreateBlock(request *CreateBlockRequest) (response *CreateBlockResponse, err error) {
	if request == nil {
		request = NewCreateBlockRequest()
	}
	response = NewCreateBlockResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteBlockRequest() (request *DeleteBlockRequest) {
	request = &DeleteBlockRequest{
		BaseRequest: &cdshttp.BaseRequest{},
	}
	request.SetDomain(consts.ApiHost)
	request.Init().WithApiInfo(consts.ServiceEKS, consts.ApiVersion, consts.ActionDeleteBlock)
	return
}

func NewDeleteBlockResponse() (response *DeleteBlockResponse) {
	response = &DeleteBlockResponse{BaseResponse: &cdshttp.BaseResponse{}}
	return
}

func (c *Client) DeleteBlock(request *DeleteBlockRequest) (response *DeleteBlockResponse, err error) {
	if request == nil {
		request = NewDeleteBlockRequest()
	}
	response = NewDeleteBlockResponse()
	err = c.Send(request, response)
	return
}
