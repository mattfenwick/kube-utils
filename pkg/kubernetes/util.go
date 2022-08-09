package kubernetes

type KeySetComparison struct {
	JustA map[string]bool
	Both  map[string]bool
	JustB map[string]bool
}

func CompareKeySets(a map[string]bool, b map[string]bool) *KeySetComparison {
	justA := map[string]bool{}
	both := map[string]bool{}
	justB := map[string]bool{}

	for key := range a {
		if _, ok := b[key]; ok {
			both[key] = true
		} else {
			justA[key] = true
		}
	}
	for key := range b {
		if _, ok := a[key]; ok {
			// nothing to do: in both
		} else {
			justB[key] = true
		}
	}

	return &KeySetComparison{
		JustA: justA,
		Both:  both,
		JustB: justB,
	}
}
