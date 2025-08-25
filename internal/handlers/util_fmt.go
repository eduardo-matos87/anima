package handlers

import "strconv"

// helper para placeholders $n em SQL dinâmico
func fmtInt(i int) string { return strconv.Itoa(i) }
