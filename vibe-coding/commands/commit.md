---
description: Crea un commit con los cambios actuales
agent: vibe
---

Crea un commit para los cambios actuales.

1. `git status` y `git diff --staged` para ver qué hay
2. Tipo de commit: feat, fix, refactor, test, docs, chore
3. Formato: `<tipo>(<scope>): <descripción en imperativo>`
4. Stage los archivos relevantes si no están staged
5. Crea el commit

Reglas:
- Máximo 72 caracteres en la primera línea
- Imperativo en inglés: "add", "fix", "remove"
- Sin punto al final
- No commitear .env ni archivos con secretos
- No hacer push

Contexto:
- Working directory: {workdir}
- Project: {project}
