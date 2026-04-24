package scoring

func ClusterScore(clusterCount int) float64 {
	switch {
	case clusterCount >= 5:
		return 100
	case clusterCount >= 4:
		return 92
	case clusterCount >= 3:
		return 82
	case clusterCount >= 2:
		return 65
	default:
		// Tek kaynaklı haberler artık 0 almıyor.
		// MinClusterCount=1 olan kategoriler (ECONOMY, TECH, GENERAL) için
		// tüm ağırlık keyword'e (0.54-0.60) düşüyordu.
		// 20 puanlık baseline ile scoring dengesi korunuyor.
		return 20
	}
}
