export const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8081";

export async function getHealth() {
  const r = await fetch(`${API_URL}/health`);
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.text();
}
