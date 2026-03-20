---
description: Reporta el estado final del trabajo y guarda lo aprendido en memoria
agent: vibe
---

Cierra la sesión de trabajo actual.

1. Corre `git status` para ver el estado
2. Si hay cambios sin commitear, pregunta si commitear antes de cerrar
3. Guarda en Neurox las decisiones importantes tomadas en esta sesión usando `neurox_save`:
   - Decisiones de arquitectura
   - Patrones descubiertos
   - Bugs encontrados y cómo se resolvieron
   - Configuraciones importantes
4. Reporta un resumen de lo hecho:

```
## Sesión terminada

### Completado
- [item 1]
- [item 2]

### Archivos modificados
- path/to/file — [qué cambió]

### Guardado en memoria
- [decisión/descubrimiento 1]
```

Contexto:
- Working directory: {workdir}
- Project: {project}
