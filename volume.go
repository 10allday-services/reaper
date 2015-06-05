package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"time"
)

type Volumes []*Volume
type Volume struct {
	AWSResource
	SizeGB     int64
	VolumeType string
	SnapShotId string
	LaunchTime time.Time
}

func NewVolume(region string, v *ec2.Volume) *Volume {
	vol := Volume{
		AWSResource: AWSResource{
			Id:     *v.VolumeID,
			Region: region,
			Tags:   make(map[string]string),
		},
		SizeGB:     *v.Size,
		VolumeType: *v.VolumeType,
		SnapShotId: *v.SnapshotID,
		LaunchTime: *v.CreateTime,
	}

	for _, tag := range v.Tags {
		vol.Tags[*tag.Key] = *tag.Value
	}

	// TODO: state
	Log.Info("Volume state: %s", *v.State)

	return &vol
}
