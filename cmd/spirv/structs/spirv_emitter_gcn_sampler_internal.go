package structs

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- spirv_emitter_gcn_sampler_internal_gen.go

type SamplerDescriptor struct {
	ClampX            uint8
	ClampY            uint8
	ClampZ            uint8
	MaxAnisoRatio     uint8
	DepthCompareFunc  uint8
	ForceUnnormalized bool
	AnisoThreshold    uint8
	McCoordTrunc      bool
	ForceDegamma      bool
	AnisoBias         float32
	TruncCoord        bool
	DisableCubeWrap   bool
	FilterMode        uint8

	MinLod  float32
	MaxLod  float32
	PerfMip uint8
	PerfZ   uint8

	LodBias          float32
	LodBiasSec       float32
	XyMagFilter      uint8
	XyMinFilter      uint8
	ZFilter          uint8
	MipFilter        uint8
	MipPointPreclamp bool
	DisableLsbCeil   bool

	BorderColorPtr  uint16
	BorderColorType uint8
}
