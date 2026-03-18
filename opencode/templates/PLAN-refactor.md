# Plan: [Refactor: descripcion del cambio]

## Goal
[Que se esta mejorando y por que. Que problema de mantenimiento, performance, o deuda tecnica resuelve]

## Business Context
[Impacto en funcionalidad existente (deberia ser cero). Riesgo de regresion]

## Technical Context
[Codigo actual que necesita mejora, patrones que se violan, metricas de complejidad]

## Implementation Steps

### Step 1: Tests de cobertura actual
- **What**: Verificar que existen tests suficientes para el codigo que se va a refactorizar. Si faltan, escribirlos primero
- **Why**: Los tests son la red de seguridad del refactor. Sin ellos, no hay forma de saber si algo se rompio
- **Where**: [archivos de test existentes o nuevos]
- **Acceptance**: Tests existentes pasan. Cobertura suficiente para detectar regresiones
- **Status**: [ ] pending

### Step 2: [Primera mejora concreta]
- **What**: [Descripcion especifica del cambio]
- **Why**: [Que mejora en concreto]
- **Where**: [archivos afectados]
- **Acceptance**: Tests siguen pasando. Build compila. El codigo queda mas limpio/simple/correcto
- **Status**: [ ] pending

### Step 3: [Segunda mejora concreta]
- **What**: [Descripcion especifica del cambio]
- **Why**: [Que mejora en concreto]
- **Where**: [archivos afectados]
- **Acceptance**: Tests siguen pasando. Build compila
- **Status**: [ ] pending

### Step N: Verificacion final
- **What**: Correr toda la suite de tests, lint y build. Comparar comportamiento antes y despues
- **Why**: Confirmar que el refactor no cambio el comportamiento observable
- **Where**: proyecto completo
- **Acceptance**: Todos los tests pasan. Lint limpio. Build exitoso
- **Status**: [ ] pending

## Verification
```bash
pnpm test                    # todos los tests pasan
pnpm test:e2e                # e2e tests pasan (si el refactor toca API)
pnpm lint                    # sin errores
npx tsc --noEmit             # sin errores de tipos
```

## Risks / Notes
- El refactor NO debe cambiar comportamiento observable. Si lo hace, es un feature, no un refactor
- Hacer commits pequenos por paso para facilitar revert si algo sale mal
- Si el refactor toca muchos archivos, considerar hacerlo en ramas separadas por area
