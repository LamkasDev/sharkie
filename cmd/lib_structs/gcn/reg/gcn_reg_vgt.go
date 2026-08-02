package reg

type VgtMultiPrimIbResetEn Reg

func (r VgtMultiPrimIbResetEn) Enable() bool { return Reg(r).ExtractBool(0) }
