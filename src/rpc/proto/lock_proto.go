package proto

const (
	MethodGetDonateInfo = iota
	MethodAddDonateInfo
)

type CenterConnCns struct {
	Addr string
}
type CenterConnCnsResult struct {
	Ret bool
}
