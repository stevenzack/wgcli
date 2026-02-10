package core

import "errors"

type REGEN string

const (
	HK REGEN = "cn-hongkong"
	SG REGEN = "ap-southeast-1"
	KR REGEN = "ap-northeast-2"
)

var (
	regenId = HK
	zoneMap = map[REGEN]string{
		HK: "cn-hongkong-c",
		SG: "ap-southeast-1c",
		KR: "ap-northeast-2a",
	}
	instanceTypeMap = map[REGEN]string{
		HK: "ecs.e-c4m1.large",
		SG: "ecs.e-c4m1.large",
		KR: "ecs.e-c4m1.large",
	}
)

func SetRegen(r string) error {
	switch r {
	case "", "hk":
		regenId = HK
	case "sg":
		regenId = SG
	case "kr":
		regenId = KR
	default:
		return errors.New("unsupported regen " + string(r))
	}

	if _, ok := zoneMap[regenId]; !ok {
		panic("zone not set for REGEN " + string(r))
	}
	if _, ok := instanceTypeMap[regenId]; !ok {
		panic("instantTypeMap not set for REGEN " + string(r))
	}
	return nil
}
