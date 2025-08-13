import { useEffect, useState } from "react";
import { getHealth } from "./services/api";

function App() {
  const [status, setStatus] = useState("…");

  useEffect(() => {
    getHealth()
      .then(setStatus)
      .catch(e => setStatus(`erro: ${e.message}`));
  }, []);

  return (
    <div style={{ fontFamily: "sans-serif", padding: "1rem" }}>
      <h1>Anima Frontend</h1>
      <p>Backend status: <strong>{status}</strong></p>
    </div>
  );
}

export default App;
