package eks

import (
	"encoding/json"
	cdshttp "github.com/capitalonline/eks-csi-driver/pkg/utils/http"
)

// CreateBlockRequest 创建块存储
type CreateBlockRequest struct {
	*cdshttp.BaseRequest
	DiskInfo CreateBlockRequestDiskInfo `json:"DiskInfo"`
	DiskName string                     `json:"DiskName"`
	AzId     string                     `json:"AzId"`
}

// CreateBlockRequestDiskInfo 块存储磁盘信息
type CreateBlockRequestDiskInfo struct {
	DiskFeature string `json:"DiskFeature"`
	DiskSize    int    `json:"DiskSize"`
}

func (req *CreateBlockRequest) ToJsonString() string {
	b, _ := json.Marshal(req)
	return string(b)
}

func (req *CreateBlockRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &req)
}

type CreateBlockResponse struct {
	*cdshttp.BaseResponse
	Code string                   `json:"Code"`
	Msg  string                   `json:"Msg"`
	Data *CreateBlockResponseData `json:"Data"`
}

func (resp *CreateBlockResponse) ToJsonString() string {
	b, _ := json.Marshal(resp)
	return string(b)
}

func (resp *CreateBlockResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &resp)
}

type CreateBlockResponseData struct {
	BlockId string `json:"BlockId"`
	TaskId  string `json:"TaskId"`
}
