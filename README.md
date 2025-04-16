
# 🧘 Anima – SuperApp de Saúde e Bem-Estar

**Anima** é uma API em Go para geração de treinos personalizados.

## Endpoints

### `GET /ping`
> Retorna `pong` para teste de saúde da API.

### `GET /treino`
> Retorna treino fixo (mockado):

```json
{
  "dia": "Segunda",
  "exercicios": ["Supino reto", "Supino inclinado", "Crucifixo", "Tríceps testa"]
}