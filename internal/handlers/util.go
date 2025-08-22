package handlers

import "strconv"

// place retorna o placeholder posicional do Postgres ($1, $2, ...)
func place(i int) string { return "$" + strconv.Itoa(i) }

// parseInt tenta converter s para int, ou retorna def
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// clampInt limita v no intervalo [lo, hi]
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
