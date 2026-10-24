package service

// MaxSKSByIPK mengembalikan batas SKS per semester berdasarkan IPK terakhir:
//   - IPK >= 3,00        : 24 SKS
//   - IPK 2,50 - 2,99    : 21 SKS
//   - IPK < 2,50         : 18 SKS
func MaxSKSByIPK(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}

func CanTakeSKS(ipk float64, currentSKS, additionalSKS int) bool {
	maxSKS := MaxSKSByIPK(ipk)

	return currentSKS+additionalSKS <= maxSKS
}