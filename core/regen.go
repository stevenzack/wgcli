package core

import "errors"

type REGEN_ID string

const (
	HK REGEN_ID = "cn-hongkong"
	SG REGEN_ID = "ap-southeast-1"
	KR REGEN_ID = "ap-northeast-2"
)

var (
	regenId = HK
	zoneMap = map[REGEN_ID]string{
		HK: "cn-hongkong-c",
		SG: "ap-southeast-1c",
		KR: "ap-northeast-2a",
	}
	instanceTypeMap = map[REGEN_ID]string{
		HK: "ecs.e-c4m1.large",
		SG: "ecs.e-c4m1.large",
		KR: "ecs.e-c4m1.large",
	}
	nameToRegionId = map[string]REGEN_ID{
		"hk": HK,
		"sg": SG,
		"kr": KR,
	}
)

func SetRegenName(r string) error {
	if r == "" {
		r = "sg"
	}
	id, ok := nameToRegionId[r]
	if !ok {
		return errors.New("unsupported regen " + string(r))
	}
	regenId = id

	if _, ok := zoneMap[regenId]; !ok {
		panic("zone not set for REGEN " + string(r))
	}
	if _, ok := instanceTypeMap[regenId]; !ok {
		panic("instantTypeMap not set for REGEN " + string(r))
	}
	return nil
}

func GetRegionName() string {
	for k, v := range nameToRegionId {
		if v == regenId {
			return k
		}
	}
	return "sg"
}
