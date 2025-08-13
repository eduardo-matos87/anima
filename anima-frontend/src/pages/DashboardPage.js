import React, { useMemo, useState } from "react";

const GOALS  = ["hipertrofia","forca","resistencia"];
const LEVELS = ["beginner","intermediate","advanced"];
const EQUIP  = ["halter","barra","maquina","livre"];

export default function DashboardPage() {
  const [goal, setGoal] = useState("hipertrofia");
  const [level, setLevel] = useState("beginner");
  const [days, setDays] = useState(3);
  const [equipment, setEquipment] = useState(new Set(EQUIP)); // todos marcados
  const [restrictions, setRestrictions] = useState("");
  const [loading, setLoading] = useState(false);
  const [plan, setPlan] = useState(null);
  const [error, setError] = useState("");

  const equipList = useMemo(() => Array.from(equipment), [equipment]);

  function toggleEq(eq) {
    setEquipment(prev => {
      const next = new Set(prev);
      if (next.has(eq)) next.delete(eq); else next.add(eq);
      return next;
    });
  }

  async function gerarPlano(e) {
    e?.preventDefault();
    setLoading(true);
    setError("");
    setPlan(null);
    try {
      const body = {
        goal,
        level,
        days_per_week: Number(days),
        equipment: equipList,
        restrictions: restrictions
          .split(",")
          .map(s => s.trim())
          .filter(Boolean)
      };
      const res = await fetch("http://localhost:8081/api/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      setPlan(json.plan);
    } catch (err) {
      setError("Falha ao gerar plano. Verifique a API em :8081.");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  // agrupa por dia
  const itemsByDay = useMemo(() => {
    const map = new Map();
    (plan?.items || []).forEach(it => {
      if (!map.has(it.day_index)) map.set(it.day_index, []);
      map.get(it.day_index).push(it);
    });
    return map;
  }, [plan]);

  return (
    <div style={{maxWidth: 980, margin: "24px auto", padding: 16, fontFamily: "Inter, system-ui, Arial"}}>
      <h1 style={{marginBottom: 8}}>Anima — Gerador de Treinos (IA)</h1>
      <p style={{color:"#555", marginBottom: 24}}>
        Monte um plano baseado no seu objetivo, nível e equipamentos disponíveis.
      </p>

      <form onSubmit={gerarPlano}
            style={{display:"grid", gap:16, gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
                    border:"1px solid #eee", padding:16, borderRadius:12, boxShadow:"0 4px 14px rgba(0,0,0,.05)"}}>
        <div>
          <label style={{display:"block", fontWeight:600, marginBottom:8}}>Objetivo</label>
          <select value={goal} onChange={e=>setGoal(e.target.value)} style={selectStyle}>
            {GOALS.map(g => <option key={g} value={g}>{g}</option>)}
          </select>
        </div>

        <div>
          <label style={{display:"block", fontWeight:600, marginBottom:8}}>Nível</label>
          <select value={level} onChange={e=>setLevel(e.target.value)} style={selectStyle}>
            {LEVELS.map(l => <option key={l} value={l}>{l}</option>)}
          </select>
        </div>

        <div>
          <label style={{display:"block", fontWeight:600, marginBottom:8}}>Dias por semana</label>
          <input type="number" min={2} max={6} value={days} onChange={e=>setDays(e.target.value)} style={inputStyle}/>
        </div>

        <div>
          <label style={{display:"block", fontWeight:600, marginBottom:8}}>Equipamentos</label>
          <div style={{display:"flex", gap:12, flexWrap:"wrap"}}>
            {EQUIP.map(eq => (
              <label key={eq} style={chipStyle(equipment.has(eq))}>
                <input type="checkbox" checked={equipment.has(eq)} onChange={()=>toggleEq(eq)} style={{marginRight:8}}/>
                {eq}
              </label>
            ))}
          </div>
        </div>

        <div style={{gridColumn:"1 / -1"}}>
          <label style={{display:"block", fontWeight:600, marginBottom:8}}>Restrições (ex: ombro, joelho)</label>
          <input value={restrictions} onChange={e=>setRestrictions(e.target.value)}
                 placeholder="separe por vírgula" style={{...inputStyle, width:"100%"}}/>
        </div>

        <div style={{gridColumn:"1 / -1", display:"flex", gap:12}}>
          <button disabled={loading} type="submit" style={btnStyle}>
            {loading ? "Gerando..." : "Gerar plano"}
          </button>
          {error && <span style={{color:"#b00020"}}>{error}</span>}
        </div>
      </form>

      {/* resultado */}
      {plan && (
        <div style={{marginTop:24}}>
          <h2 style={{marginBottom: 6}}>Plano</h2>
          <div style={{color:"#555", marginBottom:16}}>
            <b>Objetivo:</b> {plan.goal} · <b>Nível:</b> {plan.level} · <b>Dias/sem:</b> {plan.days_per_week}
          </div>

          {/* Split */}
          <div style={{display:"grid", gap:16, gridTemplateColumns:"repeat(auto-fit, minmax(280px, 1fr))"}}>
            {plan.split.map((label, i) => {
              const dayIdx = i + 1;
              const list = itemsByDay.get(dayIdx) || [];
              return (
                <div key={dayIdx} style={cardStyle}>
                  <div style={{fontWeight:700, marginBottom:8}}>Dia {dayIdx} — {label}</div>
                  {list.length === 0 && (<div style={{color:"#777"}}>Sem exercícios sugeridos.</div>)}
                  {list.map((ex) => (
                    <div key={ex.exercise_id} style={rowStyle}>
                      <div style={{fontWeight:600}}>{ex.name}</div>
                      <div style={{color:"#555"}}>{ex.sets} x {ex.reps} · descanso {ex.rest_seconds}s</div>
                    </div>
                  ))}
                </div>
              );
            })}
          </div>

          {/* Notes */}
          <div style={{marginTop:16, padding:12, border:"1px dashed #ddd", borderRadius:10, color:"#444"}}>
            <b>Dicas:</b> {plan.notes}
          </div>
        </div>
      )}
    </div>
  );
}

// estilos inline simples
const inputStyle  = { padding:"10px 12px", border:"1px solid #ddd", borderRadius:10, outline:"none" };
const selectStyle = { ...inputStyle };
const btnStyle    = { padding:"10px 16px", background:"#111827", color:"#fff", border:"none", borderRadius:10, cursor:"pointer" };
const cardStyle   = { border:"1px solid #eee", borderRadius:12, padding:12, boxShadow:"0 4px 14px rgba(0,0,0,.05)", background:"#fff" };
const rowStyle    = { display:"grid", gap:4, padding:"8px 0", borderBottom:"1px solid #f3f3f3" };
const chipStyle   = (active) => ({
  display:"inline-flex", alignItems:"center", padding:"6px 10px",
  border:"1px solid " + (active ? "#111827" : "#ddd"),
  borderRadius:999, cursor:"pointer", userSelect:"none"
});
