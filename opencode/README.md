# OpenCode Setup

Este directorio contiene una configuracion de OpenCode basada en un flujo simple de planificacion y ejecucion paso a paso.

## Objetivo

El flujo esta pensado para trabajar asi:

1. Entender bien la tarea
2. Hacer preguntas de negocio y tecnicas
3. Generar un `PLAN.md` accionable
4. Ejecutar un paso por vez
5. Pedir revision humana entre cada paso
6. Corregir feedback o avanzar al siguiente paso

## Arquitectura Actual

La configuracion gira alrededor de 3 agentes:

### `step-builder-agent`

Rol:
- descubre contexto del proyecto
- hace preguntas en bloques tematicos
- cubre negocio, producto, tecnica y riesgos
- genera `PLAN.md`

No hace:
- no implementa codigo

Uso ideal:
- cuando todavia no existe plan
- cuando el plan actual esta incompleto o quedo viejo

### `execution-orchestrator`

Rol:
- lee `PLAN.md`
- selecciona el siguiente paso
- delega la implementacion a `ts-expert-coder`
- actualiza estados del plan
- obliga a revision humana antes de continuar

No hace:
- no implementa codigo de aplicacion
- no avanza al siguiente paso sin aprobacion humana

Estados usados en `PLAN.md`:
- `[ ] pending`
- `[~] in progress`
- `[!] needs fixes`
- `[x] done`

### `ts-expert-coder`

Rol:
- implementa una unica tarea acotada
- sigue patrones locales del repo
- trabaja principalmente en TypeScript, Node.js y NestJS
- ejecuta verificaciones antes de devolver el resultado

No hace:
- no gestiona el estado global del proyecto
- no define alcance funcional
- no cambia arquitectura porque si

## Commands Disponibles

Los commands viven en `commands/` y se invocan con `/nombre`.

### `/plan <tarea>`

Usa `step-builder-agent` para:
- inspeccionar el codebase
- hacer preguntas necesarias
- confirmar entendimiento
- crear o reemplazar `PLAN.md`

Ejemplo:

```text
/plan agregar autenticacion JWT con refresh token
```

### `/plan-rewrite`

Usa `step-builder-agent` para:
- releer `PLAN.md`
- detectar huecos o ambiguedades
- hacer preguntas faltantes
- reescribir el plan con mejor detalle

Usalo cuando:
- cambian requisitos
- el plan quedo desactualizado
- los pasos son demasiado grandes o vagos

### `/execute`

Usa `execution-orchestrator` para:
- leer `PLAN.md`
- tomar el siguiente paso pendiente o con fixes
- delegarlo a `ts-expert-coder`
- mostrar cambios y pedir review humana

Importante:
- ejecuta un solo paso por vez
- no deberia avanzar automaticamente al siguiente

### `/apply-feedback <cambios>`

Usa `execution-orchestrator` para:
- tomar feedback del humano
- marcar el paso como necesitado de ajustes
- reenviar las correcciones a `ts-expert-coder`
- volver a pedir revision

Ejemplo:

```text
/apply-feedback separar el guard de admin y agregar tests al servicio
```

### `/status`

Usa `execution-orchestrator` para:
- leer `PLAN.md`
- resumir pasos completados
- indicar paso actual o en review
- mostrar lo pendiente
- sugerir la siguiente accion

## Flujo Recomendado

### Flujo minimo

```text
/plan implementar login con JWT
/status
/execute
/apply-feedback mover validacion al DTO y agregar test del guard
/execute
/status
```

### Flujo real esperado

1. Ejecutas `/plan <tarea>`
2. Respondes preguntas del planner
3. Se genera `PLAN.md`
4. Ejecutas `/execute`
5. Revisas los cambios locales
6. Si hay cambios, usas `/apply-feedback ...`
7. Si esta bien, vuelves a `/execute`
8. Repites hasta que todo el plan quede en `[x] done`

## Estructura Esperada de `PLAN.md`

El planner genera un plan con esta idea general:

```markdown
# Plan: Titulo

## Goal
...

## Business Context
...

## Technical Context
...

## Implementation Steps

### Step 1: ...
- **What**: ...
- **Why**: ...
- **Where**: ...
- **Acceptance**: ...
- **Status**: [ ] pending

## Verification
...

## Risks / Notes
...
```

Claves:
- los pasos deben ser pequenos y revisables
- el acceptance debe ser concreto
- el plan debe servir para ejecutar sin volver a redefinir producto a mitad de camino

## Skills Actuales

Solo quedaron 2 skills activas:

### `prd`

Util para:
- documentar requisitos
- estructurar una idea de producto
- generar PRDs o documentos mas orientados a negocio

### `typescript-advanced-types`

Util para:
- tipos complejos en TypeScript
- generics avanzados
- mapped types, conditional types, template literal types

## Skills Eliminadas

Se eliminaron:
- toda la familia `sdd-*`
- `find-skills`
- `_shared` asociado a SDD

La razon es que ya no forman parte del flujo actual basado en planner -> orchestrator -> worker.

## MCPs Configurados

### `engram`

Se usa para memoria persistente:
- guardar contexto de trabajo
- registrar decisiones y cambios
- mantener continuidad entre sesiones

### `context7`

Se usa para consultar documentacion actualizada de librerias.

### `miro`

Quedo configurado pero sin uso dentro del flujo principal actual.

## Archivos Importantes

- `opencode.json` - configuracion principal de agentes y MCPs
- `commands/plan.md` - inicia la planificacion
- `commands/plan-rewrite.md` - reescribe el plan
- `commands/execute.md` - ejecuta el siguiente paso
- `commands/apply-feedback.md` - aplica feedback humano
- `commands/status.md` - muestra progreso del plan

## Como Ajustar el Sistema

Si algo del comportamiento no te convence:

- cambia `opencode.json` si quieres modificar la personalidad o reglas base del agente
- cambia `commands/*.md` si quieres ajustar el comportamiento de un comando concreto

Regla practica:
- problema general del agente -> `opencode.json`
- problema puntual del comando -> `commands/*.md`

## Buenas Practicas

- usa `/plan` antes de empezar una tarea no trivial
- no ejecutes varios pasos a la vez si quieres mantener el control del review
- usa `/apply-feedback` con instrucciones concretas
- usa `/status` cuando retomes una tarea despues de una pausa
- manten `PLAN.md` como fuente visible de verdad del trabajo

## Ejemplo de Sesion

```text
/plan crear modulo de usuarios con CRUD y validaciones

[respondes preguntas del planner]

/execute

[revisas cambios]

/apply-feedback separar DTOs de create y update, y agregar test del service

/execute

/status
```

## Resumen Corto

- `step-builder-agent` piensa y planifica
- `execution-orchestrator` coordina y controla el avance
- `ts-expert-coder` implementa
- `/plan` crea el plan
- `/execute` hace un paso
- `/apply-feedback` corrige
- `/status` muestra avance
