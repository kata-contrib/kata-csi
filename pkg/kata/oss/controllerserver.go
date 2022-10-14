package oss

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

// controller server try to create/delete volumes
type controllerServer struct {
	region string
	client kubernetes.Interface
	*csicommon.DefaultControllerServer
	crdClient dynamic.Interface
}

func (cs *controllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewControllerServer(d *csicommon.CSIDriver) csi.ControllerServer {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Create create kube config is failed, err: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Create client set is failed, err: %v", err)
	}

	crdClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Create dynamic client is failed, err: %v", err)
	}

	c := &controllerServer{
		client:                  clientset,
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		crdClient:               crdClient,
	}
	return c
}

func getOssVolumeOptions(req *csi.CreateVolumeRequest) *Options {
	ossVolArgs := &Options{}
	volOptions := req.GetParameters()
	secret := req.GetSecrets()
	ossVolArgs.Path = "/"
	for k, v := range volOptions {
		key := strings.TrimSpace(strings.ToLower(k))
		value := strings.TrimSpace(v)
		if key == "bucket" {
			ossVolArgs.Bucket = value
		} else if key == "url" {
			ossVolArgs.URL = value
		} else if key == "otheropts" {
			ossVolArgs.OtherOpts = value
		} else if key == "path" {
			ossVolArgs.Path = value
		} else if key == "usesharedpath" && value == "true" {
			ossVolArgs.UseSharedPath = true
		} else if key == "authtype" {
			ossVolArgs.AuthType = value
		}
	}
	for k, v := range secret {
		key := strings.TrimSpace(strings.ToLower(k))
		value := strings.TrimSpace(v)
		if key == "akid" {
			ossVolArgs.AkID = value
		} else if key == "aksecret" {
			ossVolArgs.AkSecret = value
		}
	}
	return ossVolArgs
}

// provisioner: create/delete oss volume
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	log.Infof("CreateVolume: Starting oss createvolume, req.Name:%s, req:%v", req.Name, req)
	ossVol := getOssVolumeOptions(req)
	csiTargetVolume := &csi.Volume{}
	volumeContext := req.GetParameters()
	volumeContext["path"] = ossVol.Path
	volSizeBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	csiTargetVolume = &csi.Volume{
		VolumeId:      req.Name,
		CapacityBytes: int64(volSizeBytes),
		VolumeContext: volumeContext,
	}

	log.Infof("Provision oss volume is successfully: %s,pvName: %v", req.Name, csiTargetVolume)
	return &csi.CreateVolumeResponse{Volume: csiTargetVolume}, nil

}

// call nas api to delete oss volume
func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	log.Infof("DeleteVolume: Starting deleting volume %s", req.GetVolumeId())
	_, err := cs.client.CoreV1().PersistentVolumes().Get(context.Background(), req.VolumeId, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("DeleteVolume: Get volume %s is failed, err: %s", req.VolumeId, err.Error())
	}
	log.Infof("Delete volume %s is successfully", req.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Infof("ControllerUnpublishVolume is called, do nothing by now")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Infof("ControllerPublishVolume is called, do nothing by now")
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Infof("CreateSnapshot is called, do nothing now")
	return &csi.CreateSnapshotResponse{}, nil
}

func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Infof("DeleteSnapshot is called, do nothing now")
	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	log.Infof("ControllerExpandVolume is called, do nothing now")
	return &csi.ControllerExpandVolumeResponse{}, nil
}
