package handlers

import (
<<<<<<< HEAD
=======
	"context"
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

<<<<<<< HEAD
type GenerateReq struct {
	Objetivo string `json:"objetivo"`
	Nivel    string `json:"nivel"`
	Divisao  string `json:"divisao"`
	Dias     int    `json:"dias,omitempty"`
}

type GeneratedExercise struct {
	Nome        string `json:"nome"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
	DescansoSeg int    `json:"descanso_seg"`
}

type GenerateResp struct {
	TreinoID   string              `json:"treino_id"`
	Exercicios []GeneratedExercise `json:"exercicios"`
	CoachNotes string              `json:"coach_notes,omitempty"`
}

=======
// ====== Tipos de request/response

type GenerateReq struct {
	Objetivo string `json:"objetivo"`            // "hipertrofia" | "emagrecimento" | "forca/força" | "resistencia/resistência"
	Nivel    string `json:"nivel"`               // "iniciante" | "intermediario" | "avancado"
	Divisao  string `json:"divisao"`             // "fullbody" | "upper" | "lower" | "upperlower" | "ppl" | "push" | "pull" | "legs"
	Dias     int    `json:"dias,omitempty"`      // default 3 (hint)
	Persist  *bool  `json:"persist,omitempty"`   // default: true (persiste)
	TreinoID string `json:"treino_id,omitempty"` // opcional: fixa chave lógica
}

type GeneratedExercise struct {
	ExercicioID int    `json:"exercicio_id"`
	Nome        string `json:"nome"`
	Series      int    `json:"series"`
	Repeticoes  string `json:"repeticoes"`
	DescansoSeg int    `json:"descanso_seg,omitempty"` // 🆕 descanso entre séries
}

type GenerateResp struct {
	ID         *int                `json:"id,omitempty"` // presente se persistido
	TreinoID   string              `json:"treino_id"`    // key lógica
	Exercicios []GeneratedExercise `json:"exercicios"`   // plano gerado
	CoachNotes string              `json:"coach_notes,omitempty"`
}

// ====== Handler

>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
func GenerateTreino(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req GenerateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json inválido", http.StatusBadRequest)
			return
<<<<<<< HEAD
		}
		if req.Objetivo == "" || req.Nivel == "" || req.Divisao == "" {
			http.Error(w, "campos obrigatórios: objetivo, nivel, divisao", http.StatusBadRequest)
			return
		}
		if req.Dias <= 0 {
			req.Dias = 3
		}

		uid := getUserID(r)

		// Perfil + última métrica
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		// Exercícios mock
		exs := []GeneratedExercise{
			{Nome: "Agachamento Livre", Series: 3, Repeticoes: "8-12", DescansoSeg: 60},
			{Nome: "Supino Reto", Series: 4, Repeticoes: "8-12", DescansoSeg: 90},
			{Nome: "Levantamento Terra", Series: 3, Repeticoes: "8-12", DescansoSeg: 120},
			{Nome: "Remada Curvada", Series: 4, Repeticoes: "8-12", DescansoSeg: 60},
			{Nome: "Desenvolvimento Militar", Series: 3, Repeticoes: "8-12", DescansoSeg: 90},
			{Nome: "Puxada na Frente", Series: 4, Repeticoes: "8-12", DescansoSeg: 120},
		}

		// Coach notes (gera apenas quando use_ai=true)
		coach := ""
		if prof.UseAI != nil && *prof.UseAI {
			coach = buildCoachNotes(req, prof)
		}

		resp := GenerateResp{
			TreinoID:   time.Now().Format("20060102T150405"),
			Exercicios: exs,
			CoachNotes: coach,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// ===== Coach notes =====

func buildCoachNotes(req GenerateReq, p userProfile) string {
	obj := req.Objetivo
	if obj == "" && p.TrainingGoal != nil && *p.TrainingGoal != "" {
		obj = *p.TrainingGoal
	}

	// idade
	var idade *int
	if p.BirthDate != nil {
		y := computeAge(*p.BirthDate, time.Now())
		idade = &y
	}

	// IMC
	var imc *float64
	if p.HeightCM != nil && *p.HeightCM > 0 && p.WeightKG != nil && *p.WeightKG > 0 {
		hm := float64(*p.HeightCM) / 100.0
		v := *p.WeightKG / (hm * hm)
		v = math.Round(v*100) / 100
		imc = &v
	}

	txt := fmt.Sprintf("Plano %s (%s), divisão %s. ",
		valOr("geral", obj), valOr("nível indefinido", req.Nivel), valOr("indefinida", req.Divisao))

	if p.HeightCM != nil {
		txt += fmt.Sprintf("Altura %dcm. ", *p.HeightCM)
	}
	if p.WeightKG != nil {
		txt += fmt.Sprintf("Peso %.1fkg. ", *p.WeightKG)
	}
	if idade != nil {
		txt += fmt.Sprintf("Idade %d. ", *idade)
	}
	if imc != nil {
		txt += fmt.Sprintf("IMC=%.2f. ", *imc)
	}

	switch obj {
	case "hipertrofia":
		txt += "Foque em progressão de carga com técnica sólida; 8–12 reps nos compostos; sono ≥ 7h."
	case "emagrecimento":
		txt += "Dê ênfase à densidade do treino e controle de descanso; mantenha leve déficit calórico."
	case "resistência":
		txt += "Volume moderado e descanso curto; priorize constância e cadência controlada."
	default:
		txt += "Mantenha técnica perfeita, aquecimento e progressão gradual."
=======
		}

		normalizeReq(&req)
		if req.Objetivo == "" || req.Nivel == "" || req.Divisao == "" {
			http.Error(w, "campos obrigatórios: objetivo, nivel, divisao", http.StatusBadRequest)
			return
		}
		persist := true
		if req.Persist != nil {
			persist = *req.Persist
		}

		uid := getUserID(r)

		// Perfil + métrica recente
		prof, _ := loadUserProfile(r.Context(), db, uid)
		if wkg, _ := latestWeight(r.Context(), db, uid); wkg != nil {
			prof.WeightKG = wkg
		}

		// Plano com diversidade por grupo + divisão (v1.1) + descanso
		exs, err := buildPlanV11(r.Context(), db, req)
		if err != nil {
			http.Error(w, "falha ao montar plano: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(exs) == 0 {
			http.Error(w, "não há exercícios cadastrados no catálogo", http.StatusConflict)
			return
		}

		// Coach notes (se use_ai = true)
		coach := ""
		if prof.UseAI == nil || *prof.UseAI {
			coach = buildCoachNotes(req, prof)
		}

		// Persistência opcional
		key := req.TreinoID
		if key == "" {
			key = "gen-" + time.Now().Format("20060102T150405")
		}

		var insertedID *int
		if persist {
			id, err := persistPlan(r.Context(), db, key, req, coach, exs)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					key = "gen-" + time.Now().Format("20060102T150405.000")
					id, err = persistPlan(r.Context(), db, key, req, coach, exs)
				}
			}
			if err != nil {
				http.Error(w, "falha ao salvar treino: "+err.Error(), http.StatusInternalServerError)
				return
			}
			insertedID = &id
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		resp := GenerateResp{
			ID:         insertedID,
			TreinoID:   key,
			Exercicios: exs,
			CoachNotes: coach,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// ====== v1.1 + descanso: diversidade por grupo + divisão

func buildPlanV11(ctx context.Context, db *sql.DB, req GenerateReq) ([]GeneratedExercise, error) {
	target := 6 // nº-alvo por sessão

	// grupos da sessão conforme divisão
	sessionGroups := groupsForDivision(req.Divisao)

	// 1) tenta 1 exercício por grupo-alvo (em ordem)
	var pool []exRow
	for _, g := range sessionGroups {
		row, err := queryFirstByGroup(ctx, db, g, req.Nivel)
		if err != nil {
			return nil, err
		}
		if row != nil {
			pool = append(pool, *row)
		}
	}

	// 2) completa com catálogo geral do nível (sem repetir IDs)
	if len(pool) < target {
		rest, err := queryExercises(ctx, db, req.Nivel, target-len(pool))
		if err != nil {
			return nil, err
		}
		used := make(map[int]struct{}, len(pool))
		for _, r := range pool {
			used[r.id] = struct{}{}
		}
		for _, r := range rest {
			if _, seen := used[r.id]; seen {
				continue
			}
			pool = append(pool, r)
			if len(pool) == target {
				break
			}
		}
	}

	if len(pool) == 0 {
		return nil, nil
	}
	if len(pool) > target {
		pool = pool[:target]
	}

	reps := repsByGoal(req.Objetivo)

	out := make([]GeneratedExercise, 0, len(pool))
	for i, it := range pool {
		series := 3
		if req.Objetivo == "hipertrofia" && i%2 == 1 {
			series = 4
		}
		rest := restForExercise(req.Objetivo, req.Nivel, it) // 🆕 descanso por exercício

		out = append(out, GeneratedExercise{
			ExercicioID: it.id,
			Nome:        it.name,
			Series:      series,
			Repeticoes:  reps,
			DescansoSeg: rest,
		})
	}
	return out, nil
}

func repsByGoal(goal string) string {
	switch strings.ToLower(goal) {
	case "hipertrofia":
		return "8-12"
	case "emagrecimento":
		return "12-15"
	case "forca", "força":
		return "4-6"
	case "resistencia", "resistência":
		return "12-20"
	default:
		return "8-12"
	}
}

// ====== Heurística de descanso

type exRow struct {
	id           int
	name         string
	muscleGroup  string
	difficulty   string
	isBodyweight bool
}

// Define descanso base por objetivo, ajusta por composto/isolado, peso corporal e dificuldade.
// Intervalos alvo (heurística segura):
// - hipertrofia: 60–90s; compostos tendem a ~90s, isolados ~60–75s
// - força: 120–180s; compostos mais altos (~150–180s)
// - emagrecimento/resistência: 30–60s
func restForExercise(goal, nivel string, ex exRow) int {
	goal = strings.ToLower(strings.TrimSpace(goal))
	nivel = strings.ToLower(strings.TrimSpace(nivel))

	// base por objetivo
	base := 75 // default
	switch goal {
	case "hipertrofia":
		base = 75
	case "emagrecimento", "resistencia", "resistência":
		base = 45
	case "forca", "força":
		base = 150
	}

	// composto vs isolado
	if isCompound(ex.name, ex.muscleGroup) {
		base += 30 // compostos pedem mais recuperação
	} else {
		base -= 10 // isolados, um pouco menos
	}

	// peso corporal costuma exigir menos descanso que máximos pesados
	if ex.isBodyweight {
		base -= 10
	}

	// ajuste por dificuldade
	switch nivel {
	case "iniciante":
		// manter faixas moderadas
		if base > 90 && (goal == "hipertrofia" || goal == "emagrecimento" || strings.HasPrefix(goal, "resist")) {
			base = 90
		}
	case "avancado":
		// para força avançada, um pouco mais
		if goal == "forca" || goal == "força" {
			base += 15
		}
	}

	// clamp por objetivo
	switch goal {
	case "hipertrofia":
		return clampRest(base, 60, 105)
	case "forca", "força":
		return clampRest(base, 120, 180)
	case "emagrecimento", "resistencia", "resistência":
		return clampRest(base, 30, 75)
	default:
		return clampRest(base, 45, 105)
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
	}
	if idade != nil && *idade >= 40 {
		txt += " Aqueça bem ombros/quadril; evite picos de carga abruptos."
	}
	return txt
}

<<<<<<< HEAD
func valOr(def, v string) string {
	if v == "" {
		return def
	}
=======
func clampRest(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// heurística simples para detectar composto
func isCompound(name, mg string) bool {
	n := strings.ToLower(name)
	m := strings.ToLower(mg)
	// palavras-chave de compostos
	kw := []string{
		"agachamento", "squat",
		"supino", "bench",
		"terra", "deadlift",
		"remada", "row",
		"desenvolvimento", "overhead", "press",
		"levantamento", "clean", "snatch",
		"barra fixa", "pull-up", "chin-up",
		"paralela", "dip",
		"lunge", "passada",
		"puxada", "lat pulldown",
	}
	for _, k := range kw {
		if strings.Contains(n, k) {
			return true
		}
	}
	// grupos que costumam ser compostos quando o nome é genérico
	if m == "peito" || m == "costas" || m == "pernas" || m == "quadriceps" || m == "posterior" || m == "gluteos" || m == "lombar" {
		if strings.Contains(n, "com barra") || strings.Contains(n, "livre") {
			return true
		}
	}
	return false
}

// ====== Divisão / grupos

func groupsForDivision(div string) []string {
	d := strings.ToLower(strings.TrimSpace(div))
	switch d {
	case "upper", "upperlower":
		return []string{"peito", "costas", "ombros", "biceps", "triceps", "core"}
	case "lower":
		return []string{"quadriceps", "posterior", "gluteos", "panturrilhas", "lombar", "core"}
	case "ppl", "push":
		return []string{"peito", "ombros", "triceps", "core", "quadriceps", "panturrilhas"}
	case "pull":
		return []string{"costas", "lombar", "biceps", "posterior", "core", "trapézio"}
	case "legs":
		return []string{"quadriceps", "posterior", "gluteos", "panturrilhas", "lombar", "core"}
	default: // fullbody
		return []string{"peito", "costas", "pernas", "ombros", "core", "biceps", "triceps"}
	}
}

func normalizeGroupName(g string) []string {
	g = strings.ToLower(strings.TrimSpace(g))
	switch g {
	case "peito", "chest":
		return []string{"peito", "chest"}
	case "costas", "back":
		return []string{"costas", "back"}
	case "ombros", "ombro", "shoulders", "delts", "deltoids":
		return []string{"ombros", "shoulders"}
	case "biceps", "bíceps", "arms", "arm", "bis":
		return []string{"biceps", "arms"}
	case "triceps", "tríceps", "tris":
		return []string{"triceps", "arms"}
	case "core", "abdomen", "abs":
		return []string{"core", "abs"}
	case "pernas", "legs", "lower":
		return []string{"pernas", "legs"}
	case "quadriceps", "quads":
		return []string{"quadriceps", "pernas", "legs"}
	case "posterior", "posterior de coxa", "hamstrings", "hams":
		return []string{"posterior", "pernas", "legs", "hamstrings"}
	case "gluteos", "glúteos", "glutes":
		return []string{"gluteos", "pernas", "legs", "glutes"}
	case "panturrilhas", "calves":
		return []string{"panturrilhas", "pernas", "legs", "calves"}
	case "lombar", "lower back":
		return []string{"lombar", "costas", "back", "lower back"}
	case "trapézio", "trapezio", "traps":
		return []string{"trapézio", "trapezio", "costas", "back", "traps"}
	default:
		if g != "" {
			return []string{g}
		}
		return []string{}
	}
}

// ====== Acesso ao catálogo

// pega 1 exercício do grupo (normalizado), preferindo por nível
func queryFirstByGroup(ctx context.Context, db *sql.DB, group string, nivel string) (*exRow, error) {
	var row exRow

	alts := normalizeGroupName(group)
	if len(alts) == 0 {
		return nil, nil
	}
	inList := "'" + strings.Join(alts, "','") + "'"

	q := `
		SELECT id, name, lower(muscle_group) AS mg, lower(difficulty) AS diff,
		       COALESCE(is_bodyweight, false) AS bw
		FROM exercises
		WHERE lower(muscle_group) IN (` + inList + `)
	`
	args := []any{}
	if nivel != "" {
		q += ` AND lower(difficulty) = $1 `
		args = append(args, strings.ToLower(nivel))
	}
	q += ` ORDER BY id ASC LIMIT 1`

	err := db.QueryRowContext(ctx, q, args...).Scan(&row.id, &row.name, &row.muscleGroup, &row.difficulty, &row.isBodyweight)
	if err == sql.ErrNoRows {
		// sem filtro de nível
		q2 := `
			SELECT id, name, lower(muscle_group) AS mg, lower(difficulty) AS diff,
			       COALESCE(is_bodyweight, false) AS bw
			FROM exercises
			WHERE lower(muscle_group) IN (` + inList + `)
			ORDER BY id ASC LIMIT 1
		`
		err2 := db.QueryRowContext(ctx, q2).Scan(&row.id, &row.name, &row.muscleGroup, &row.difficulty, &row.isBodyweight)
		if err2 == sql.ErrNoRows {
			return nil, nil
		}
		if err2 != nil {
			return nil, err2
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// catálogo geral (por nível, fallback sem nível)
func queryExercises(ctx context.Context, db *sql.DB, nivel string, limit int) ([]exRow, error) {
	args := []any{}
	q := `
		SELECT id, name, lower(muscle_group) AS mg, lower(difficulty) AS diff,
		       COALESCE(is_bodyweight, false) AS bw
		FROM exercises
	`
	if nivel != "" {
		q += ` WHERE lower(difficulty) = $1 `
		args = append(args, strings.ToLower(nivel))
	}
	q += ` ORDER BY id ASC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []exRow
	for rows.Next() {
		var r exRow
		if err := rows.Scan(&r.id, &r.name, &r.muscleGroup, &r.difficulty, &r.isBodyweight); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ====== Persistência

func persistPlan(ctx context.Context, db *sql.DB, key string, req GenerateReq, coach string, plan []GeneratedExercise) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var treinoID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO treinos (objetivo, nivel, dias, divisao, treino_key, coach_notes)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, req.Objetivo, req.Nivel, req.Dias, req.Divisao, key, nullIfEmpty(coach)).Scan(&treinoID)
	if err != nil {
		return 0, err
	}

	for _, ex := range plan {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO treino_exercicios (treino_id, exercicio_id, series, repeticoes)
			VALUES ($1,$2,$3,$4)
		`, treinoID, ex.ExercicioID, ex.Series, ex.Repeticoes)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return treinoID, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// ====== Helpers já existentes no projeto (tipos/funcs usados aqui)

func normalizeReq(req *GenerateReq) {
	req.Objetivo = strings.TrimSpace(strings.ToLower(req.Objetivo))
	req.Nivel = strings.TrimSpace(strings.ToLower(req.Nivel))
	req.Divisao = strings.TrimSpace(strings.ToLower(req.Divisao))
	if req.Dias <= 0 {
		req.Dias = 3
	}
}

func buildCoachNotes(req GenerateReq, p userProfile) string {
	obj := req.Objetivo
	if obj == "" && p.TrainingGoal != nil && *p.TrainingGoal != "" {
		obj = *p.TrainingGoal
	}

	// idade
	var idade *int
	now := time.Now()
	if p.BirthDate != nil {
		y := computeAge(*p.BirthDate, now)
		idade = &y
	}

	// IMC
	var imc *float64
	if p.HeightCM != nil && *p.HeightCM > 0 && p.WeightKG != nil && *p.WeightKG > 0 {
		hm := float64(*p.HeightCM) / 100.0
		v := *p.WeightKG / (hm * hm)
		v = math.Round(v*100) / 100
		imc = &v
	}

	txt := fmt.Sprintf("Plano %s (%s), divisão %s. ",
		valOr("geral", obj), valOr("nível indefinido", req.Nivel), valOr("indefinida", req.Divisao))

	if p.HeightCM != nil {
		txt += fmt.Sprintf("Altura %dcm. ", *p.HeightCM)
	}
	if p.WeightKG != nil {
		txt += fmt.Sprintf("Peso %.1fkg. ", *p.WeightKG)
	}
	if idade != nil {
		txt += fmt.Sprintf("Idade %d. ", *idade)
	}
	if imc != nil {
		txt += fmt.Sprintf("IMC=%.2f. ", *imc)
	}

	switch obj {
	case "hipertrofia":
		txt += "Foque em progressão de carga com técnica sólida; 8–12 reps nos compostos; sono ≥ 7h."
	case "emagrecimento":
		txt += "Aumente densidade do treino (descanso curto) e mantenha leve déficit calórico."
	case "forca", "força":
		txt += "Priorize compostos pesados; séries curtas e descanso maior; monitore a técnica."
	case "resistencia", "resistência":
		txt += "Volume moderado/alto, cadência controlada e constância semanal."
	default:
		txt += "Mantenha técnica perfeita, aquecimento e progressão gradual."
	}
	if idade != nil && *idade >= 40 {
		txt += " Aqueça bem ombros/quadril; evite picos de carga abruptos."
	}
	return txt
}

func valOr(def, v string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
	return v
}
