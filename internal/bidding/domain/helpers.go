package domain

import "math"

func SafeLess(a, b, eps float64) bool {
	return (b - a) > math.Abs(eps)
}
