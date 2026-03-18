# Plan: [Fix: descripcion del bug]

## Goal
[Que esta fallando y cual es el comportamiento esperado]

## Business Context
[Impacto en usuarios, urgencia, quien reporto, escenarios afectados]

## Technical Context
[Donde se origina el error, que modulos estan involucrados, logs o trazas relevantes]

## Implementation Steps

### Step 1: Reproducir el bug
- **What**: Escribir un test que reproduzca el bug y falle (RED)
- **Why**: Confirmar que el bug existe y que el fix se puede verificar automaticamente
- **Where**: `src/contexts/<context>/.../[archivo].spec.ts`
- **Acceptance**: Test escrito que falla con el comportamiento actual
- **Status**: [ ] pending

### Step 2: Aplicar el fix
- **What**: Corregir el codigo que causa el bug
- **Why**: Resolver el problema de raiz, no solo el sintoma
- **Where**: [archivos afectados]
- **Acceptance**: El test del Step 1 ahora pasa (GREEN). No se rompen tests existentes
- **Status**: [ ] pending

### Step 3: Refactor si aplica
- **What**: Limpiar el fix si quedo sucio, mejorar nombres, extraer logica si es necesario
- **Why**: Mantener la calidad del codigo despues del fix
- **Where**: [mismos archivos]
- **Acceptance**: Todos los tests siguen pasando. El codigo queda limpio y legible
- **Status**: [ ] pending

## Verification
```bash
pnpm test                    # todos los tests pasan incluyendo el nuevo
pnpm lint                    # sin errores
npx tsc --noEmit             # sin errores de tipos
```

## Risks / Notes
- Verificar que el fix no introduce regresiones en flujos relacionados
- Si el bug afecta datos existentes, considerar si se necesita migracion o script de correccion
