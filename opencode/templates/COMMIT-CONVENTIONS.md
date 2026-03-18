# Commit and PR Conventions

## Commits

### Formato: Conventional Commits

```
<type>(<scope>): <description>
```

### Types

| Type | Cuando usarlo |
|---|---|
| `feat` | Nueva funcionalidad |
| `fix` | Correccion de bug |
| `refactor` | Cambio de codigo que no agrega funcionalidad ni corrige bug |
| `test` | Agregar o modificar tests |
| `docs` | Cambios en documentacion |
| `chore` | Tareas de mantenimiento (deps, config, scripts) |
| `style` | Formato, espacios, puntos y comas (sin cambio de logica) |
| `perf` | Mejora de performance |
| `ci` | Cambios en CI/CD |

### Scope

Usar el nombre del contexto o modulo afectado:

```
feat(commercial): add credit transfer between licenses
fix(scheduling): handle timezone offset in session booking
refactor(shared): extract pagination helper to shared kernel
test(commercial): add unit tests for deal activation
```

### Reglas

- Primera linea: maximo 72 caracteres
- Usar imperativo presente en ingles: "add", "fix", "remove" (no "added", "fixes", "removing")
- No terminar con punto
- Si el cambio necesita explicacion, agregar body despues de linea en blanco

### Ejemplos

```
feat(commercial): add workspace owner management

Add commands to add and remove workspace owners.
Includes validation that prevents removing the last owner.
```

```
fix(scheduling): prevent double-booking on same time slot

The previous check only validated date, not time range overlap.
Now uses DateRange.overlaps() for proper collision detection.
```

```
refactor(shared): replace raw string IDs with Uuid value object

Migrates all entity IDs from plain strings to Uuid VO.
No behavior change - all existing tests pass.
```

## Pull Requests

### Titulo

Mismo formato que commits pero puede ser mas descriptivo:

```
feat(commercial): implement credit transfer between licenses with validation and rollback
```

### Body

```markdown
## Summary
- [1-3 bullet points explicando que hace el PR]

## Changes
- [Lista de cambios significativos]

## Testing
- [Que tests se agregaron/modificaron]
- [Como verificar manualmente si aplica]

## Notes
- [Decisiones tomadas, trade-offs, cosas que revisar con atencion]
```

### Reglas de PR

- Un PR por feature/fix/refactor. No mezclar cambios no relacionados
- El PR debe compilar y pasar tests antes de pedir review
- Si el PR es grande, considerar dividirlo en PRs mas pequenos
- Incluir screenshots o logs si el cambio afecta algo visual o genera output
- Linkear issue/ticket si existe

## Pre-commit (ya configurado)

El proyecto usa Husky + lint-staged que en cada commit ejecuta:
1. `eslint --fix` en archivos `.ts/.js` modificados
2. `prettier --write` en archivos modificados
3. `jest --findRelatedTests` para correr tests afectados

Si alguno falla, el commit se rechaza.
