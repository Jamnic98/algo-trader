package strategies

type SimpleStrategy struct {
	shortSMA []float64
	longSMA  []float64
}

func (s *SimpleStrategy) Update(close float64) (side string, qty float64) {
	s.shortSMA = append(s.shortSMA, close)
	s.longSMA = append(s.longSMA, close)

	if len(s.shortSMA) > 5 {
		s.shortSMA = s.shortSMA[1:]
	}
	if len(s.longSMA) > 10 {
		s.longSMA = s.longSMA[1:]
	}

	if len(s.shortSMA) < 5 || len(s.longSMA) < 10 {
		return "", 0
	}

	// Compute averages
	var sumShort, sumLong float64
	for _, v := range s.shortSMA {
		sumShort += v
	}
	for _, v := range s.longSMA {
		sumLong += v
	}
	shortAvg := sumShort / float64(len(s.shortSMA))
	longAvg := sumLong / float64(len(s.longSMA))

	// Basic crossover
	if shortAvg > longAvg {
		return "BUY", 0.01 // buy 0.01 BTC
	} else if shortAvg < longAvg {
		return "SELL", 0.01
	}

	return "", 0
}
