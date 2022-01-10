package election

type Election struct {
	M int
}

func (e *Election) Add_ID(id int) {
	e.M = id
}

func (e *Election) CompaireElection(eN Election) string {
	if e.M > eN.M {
		return "gt"
	} else if e.M == eN.M {
		return "eq"
	}
	return "lt"
}

func NewElection(id int) *Election {

	return &Election{M: id}
}
