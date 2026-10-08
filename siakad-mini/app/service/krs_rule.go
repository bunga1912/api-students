package service

func MaxSKSByIPK(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 20
	default:
		return 18
	}
}

func CanTakeSKS(ipk float64, currentSKS, additionalSKS int) bool {
	maxSKS := MaxSKSByIPK(ipk)

	return currentSKS+additionalSKS <= maxSKS
}
