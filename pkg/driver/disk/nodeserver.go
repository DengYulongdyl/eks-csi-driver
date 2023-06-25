package disk

import (
	"context"
	"fmt"
	"github.com/capitalonline/cds-csi-driver/pkg/driver/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	cdsDisk "github.com/capitalonline/cck-sdk-go/pkg/cck/disk"
)

// storing staging disk
var diskStagingMap = map[string]string{}

// storing unstaging disk
var diskUnstagingMap = map[string]string{}

// storing publishing disk
var diskPublishingMap = map[string]string{}

// storing published disk
var diskPublishedMap = map[string]string{}

// storing unpublishing disk
var diskUnpublishingMap = map[string]string{}

// storing formatted disk
var diskFormattedMap = map[string]string{}

func NewNodeServer(d *DiskDriver) *NodeServer {
	return &NodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
	}
}

func (n *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nodeCap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}

	// Disk Metric enable config
	nodeSvcCap := []*csi.NodeServiceCapability{nodeCap}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nodeSvcCap,
	}, nil
}

func (n *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	return nil, nil
}

func (n *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (n *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (n *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (n *NodeServer) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func findDeviceNameByUuid(diskUuid string) (string, error) {
	log.Infof("findDeviceNameByUuid: diskUuid is: %s", diskUuid)

	diskUuidFormat := strings.ReplaceAll(diskUuid, "-", "")
	cmdScan := "ls /dev/sd[a-z]"
	deviceNameStr, err := utils.RunCommand(cmdScan)
	if err != nil {
		log.Errorf("findDeviceNameByUuid: cmdScan: %s failed, err is: %s", cmdScan, err)
		return "", err
	}

	// log.Infof("findDeviceNameByUuid: deviceNameStr : %s", deviceNameStr)
	// get device such as /dev/sda
	deviceNameStr = strings.ReplaceAll(deviceNameStr, "\n", " ")
	// log.Infof("findDeviceNameByUuid: deviceNameStr : %s", deviceNameStr)
	deviceNameUuid := map[string]string{}
	for _, deviceName := range strings.Split(deviceNameStr, " ") {
		if deviceName == "" {
			continue
		}
		cmdGetUid := fmt.Sprintf("/lib/udev/scsi_id -g -u %s", deviceName)
		// log.Infof("findDeviceNameByUuid: cmdGetUid is: %s", cmdGetUid)
		uuidStr, err := utils.RunCommand(cmdGetUid)
		if err != nil {
			log.Errorf("findDeviceNameByUuid: get deviceName: %s uuid failed, err is: %s", deviceName, err)
			return "", err
		}

		deviceNameUuid[strings.TrimSpace(strings.TrimPrefix(uuidStr, "3"))] = deviceName
	}

	log.Debugf("findDeviceNameByUuid: deviceNameUuid is: %+v", deviceNameUuid)

	// compare
	if _, ok := deviceNameUuid[diskUuidFormat]; !ok {
		log.Errorf("findDeviceNameByUuid: diskUuid: %s is not exist", diskUuidFormat)
		return "", fmt.Errorf("findDeviceNameByUuid: diskUuid: %s is not exist", diskUuidFormat)
	}

	deviceName := deviceNameUuid[diskUuidFormat]
	if deviceName == "" {
		log.Errorf("findDeviceNameByUuid: diskUuid: %s, deviceName is empty", diskUuidFormat)
		return "", fmt.Errorf("findDeviceNameByUuid: diskUuid: %s, deviceName is empty", diskUuidFormat)
	}

	log.Infof("findDeviceNameByUuid: successfully, diskUuid: %s, deviceName is: %s", diskUuidFormat, deviceNameUuid[diskUuidFormat])

	return deviceName, nil
}

func bindMountGlobalPathToPodPath(volumeID, stagingTargetPath, podPath string) error {
	log.Infof("bindMountGlobalPathToPodPath: volumeID: %s, stagingTargetPath:%s, targetPath:%s", volumeID, stagingTargetPath, podPath)

	cmd := fmt.Sprintf("mount --bind %s %s", stagingTargetPath, podPath)
	if _, err := utils.RunCommand(cmd); err != nil {
		log.Errorf("bindMountGlobalPathToPodPath: volumeID:%s, bind mount stagingTargetPath: %s to podPath: %s failed, err is: %s", volumeID, stagingTargetPath, podPath, err.Error())
		return fmt.Errorf("bindMountGlobalPathToPodPath: volumeID:%s, bind mount stagingTargetPath: %s to podPath: %s failed, err is: %s", volumeID, stagingTargetPath, podPath, err.Error())
	}

	log.Infof("bindMountGlobalPathToPodPath: Successfully!")

	return nil

}

func unBindMountGlobalPathFromPodPath(targetPath string) error {
	log.Infof("unBindMountGlobalPathFromPodPath: targetPath:%s", targetPath)

	cmdUnBindMount := fmt.Sprintf("umount %s", targetPath)
	if _, err := utils.RunCommand(cmdUnBindMount); err != nil {
		if strings.Contains(err.Error(), "target is busy") || strings.Contains(err.Error(), "device is busy") {
			log.Warnf("unBindMountGlobalPathFromPodPath: unStagingPath is busy(occupied by another process)")

			cmdKillProcess := fmt.Sprintf("fuser -m -k %s", targetPath)
			if _, err = utils.RunCommand(cmdKillProcess); err != nil {
				log.Errorf("unBindMountGlobalPathFromPodPath: targetPath is busy, but kill occupied process failed, err is: %s", err.Error())
				return fmt.Errorf("unBindMountGlobalPathFromPodPath: targetPath is busy, but kill occupied process failed, err is: %s", err.Error())
			}

			log.Warnf("unBindMountGlobalPathFromPodPath: targetPath is busy and kill occupied process succeed!")

			if err = utils.Unmount(targetPath); err != nil {
				log.Errorf("unBindMountGlobalPathFromPodPath: unmount req.TargetPath:%s failed(again), err is: %s", targetPath, err.Error())
				return fmt.Errorf("unBindMountGlobalPathFromPodPath: unmount req.TargetPath:%s failed(again), err is: %s", targetPath, err.Error())
			}

			log.Infof("unBindMountGlobalPathFromPodPath: Successfully!")
			return nil
		}

		log.Errorf("unBindMountGlobalPathFromPodPath: unmount req.TargetPath:%s failed, err is: %s", targetPath, err.Error())
		return fmt.Errorf("unBindMountGlobalPathFromPodPath: unmount req.TargetPath:%s failed, err is: %s", targetPath, err.Error())
	}

	log.Infof("unBindMountGlobalPathFromPodPath: Successfully!")

	return nil

}

func mountDiskDeviceToNodeGlobalPath(deviceName, targetGlobalPath, fstype string) error {
	log.Infof("mountDiskDeviceToNodeGlobalPath: deviceName is: %s, targetGlobalPath is: %s, fstype is: %s", deviceName, targetGlobalPath, fstype)

	cmd := fmt.Sprintf("mount -t %s %s %s", fstype, deviceName, targetGlobalPath)
	if _, err := utils.RunCommand(cmd); err != nil {
		log.Errorf("mountDiskDeviceToNodeGlobalPath: err is: %s", err)
		return err
	}

	log.Infof("mountDiskDeviceToNodeGlobalPath: Successfully!")

	return nil
}

func unMountDiskDeviceFromNodeGlobalPath(volumeID, unStagingPath string) error {
	log.Infof("unMountDeviceFromNodeGlobalPath: volumeID is: %s, unStagingPath: %s", volumeID, unStagingPath)

	cmdUnstagingPath := fmt.Sprintf("umount %s", unStagingPath)
	if _, err := utils.RunCommand(cmdUnstagingPath); err != nil {
		if strings.Contains(err.Error(), "target is busy") || strings.Contains(err.Error(), "device is busy") {
			log.Warnf("unMountDiskDeviceFromNodeGlobalPath: unStagingPath is busy(occupied by another process)")

			cmdKillProcess := fmt.Sprintf("fuser -m -k %s", unStagingPath)
			if _, err = utils.RunCommand(cmdKillProcess); err != nil {
				log.Errorf("unMountDiskDeviceFromNodeGlobalPath: unStagingPath is busy, but kill occupied process failed, err is: %s", err.Error())
				return fmt.Errorf("unMountDiskDeviceFromNodeGlobalPath: unStagingPath is busy, but kill occupied process failed, err is: %s", err.Error())
			}

			log.Warnf("unMountDiskDeviceFromNodeGlobalPath: unStagingPath is busy and kill occupied process succeed!")

			if _, err = utils.RunCommand(cmdUnstagingPath); err != nil {
				log.Errorf("unMountDiskDeviceFromNodeGlobalPath: unmount device from req.StagingTargetPath:%s failed(again), err is: %s", unStagingPath, err.Error())
				return fmt.Errorf("unMountDiskDeviceFromNodeGlobalPath: unmount device from req.StagingTargetPath:%s failed(again), err is: %s", unStagingPath, err.Error())
			}

			log.Infof("unMountDiskDeviceFromNodeGlobalPath: Successfully!")
			return nil
		}

		log.Errorf("unMountDiskDeviceFromNodeGlobalPath: unmount device from req.StagingTargetPath:%s failed, err is: %s", unStagingPath, err.Error())
		return fmt.Errorf("unMountDiskDeviceFromNodeGlobalPath: unmount device from req.StagingTargetPath:%s failed, err is: %s", unStagingPath, err.Error())
	}

	log.Infof("unMountDiskDeviceFromNodeGlobalPath: Successfully!")

	return nil

}

func formatDiskDevice(diskId, deviceName, fsType string) error {
	log.Infof("formatDiskDevice: deviceName is: %s, fsType is: %s", deviceName, fsType)

	// Step 1: check deviceName(disk) is scannable
	scanDeviceCmd := fmt.Sprintf("fdisk -l | grep %s | grep -v grep", deviceName)
	if _, err := utils.RunCommand(scanDeviceCmd); err != nil {
		log.Errorf("formatDiskDevice: scanDeviceCmd: %s failed, err is: %s", scanDeviceCmd, err.Error())
		return err
	}

	// Step 2: format deviceName(disk)
	var formatDeviceCmd string
	if fsType == DefaultFsTypeXfs {
		formatDeviceCmd = fmt.Sprintf("mkfs.xfs %s", deviceName)
	} else if fsType == FsTypeExt3 {
		formatDeviceCmd = fmt.Sprintf("mkfs.ext3 %s", deviceName)
	} else if fsType == FsTypeExt4 {
		formatDeviceCmd = fmt.Sprintf("mkfs.ext4 %s", deviceName)
	} else {
		log.Errorf("formatDiskDevice: fsType not support, should be [ext4/ext3/xfs]")
		return fmt.Errorf("formatDiskDevice: fsType not support, should be [ext4/ext3/xfs]")
	}

	if out, err := utils.RunCommand(formatDeviceCmd); err != nil {
		if strings.Contains(out, "existing filesystem") {
			diskFormattedMap[diskId] = "formatted"
			log.Warnf("formatDiskDevice: deviceName: %s had been formatted, avoid multi formatting, return directly", deviceName)
			// log.Infof("formatDiskDevice: Successfully!")
			return nil
		}
		log.Errorf("formatDiskDevice: formatDeviceCmd: %s failed, err is: %s", formatDeviceCmd, err.Error())
		return err
	}

	// storing formatted disk
	diskFormattedMap[diskId] = "formatted"
	res, err := cdsDisk.UpdateBlockFormatFlag(&cdsDisk.UpdateBlockFormatFlagArgs{
		BlockID:  diskId,
		IsFormat: 1,
	})
	if err != nil || res.Code != "Success" {
		log.Errorf("formatDiskDevice: update  database error, err is: %s", err)
		return err
	}

	log.Infof("formatDiskDevice: Successfully!")

	return nil
}
