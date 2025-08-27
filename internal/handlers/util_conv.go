package handlers

import (
	"math"
	"strconv"
)

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case uint, uint8, uint16, uint32, uint64:
		// converte via string para evitar overflows
		s := strconv.FormatUint(reflectToUint64(x), 10)
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, false
		}
		// rejeita não-inteiro (ex.: 1.2)
		if math.Trunc(x) != x {
			return 0, false
		}
		return int64(x), true
	case string:
		if x == "" {
			return 0, false
		}
		i, err := strconv.ParseInt(x, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, false
		}
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflectToUint64(x)), true
	case string:
		if x == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// helper interno p/ converter uints sem usar reflect nas chamadas
func reflectToUint64(v any) uint64 {
	switch t := v.(type) {
	case uint:
		return uint64(t)
	case uint8:
		return uint64(t)
	case uint16:
		return uint64(t)
	case uint32:
		return uint64(t)
	case uint64:
		return t
	default:
		return 0
	}
}
