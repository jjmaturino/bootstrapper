package platform

type Type string

const VM Type = "virtual_machine"

func (pT *Type) String() string {
	return string(*pT)
}
