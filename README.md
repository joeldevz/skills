# Skills & OpenCode Config

Repositorio con la configuracion de OpenCode y recursos compartidos para el equipo.

## Estructura

```
skills/
  prd/                          # Skill para generar PRDs

opencode/
  opencode.json                 # Configuracion principal de agentes y MCPs
  tui.json                      # Configuracion de la interfaz TUI
  README.md                     # Guia de uso del flujo de trabajo
  commands/
    plan.md                     # /plan - crear plan de trabajo
    plan-rewrite.md             # /plan-rewrite - reescribir plan existente
    execute.md                  # /execute - ejecutar siguiente paso
    apply-feedback.md           # /apply-feedback - aplicar feedback humano
    status.md                   # /status - ver progreso del plan
    onboard.md                  # /onboard - explorar proyecto antes de trabajar
    commit.md                   # /commit - crear commit convencional
    pr.md                       # /pr - crear pull request
  plugins/
    engram.ts                   # Plugin de memoria persistente
  skills/
    prd/                        # Skill para PRDs
    typescript-advanced-types/  # Skill para tipos avanzados de TS
  templates/
    CONVENTIONS.md              # Template de convenciones para copiar a cada proyecto
    COMMIT-CONVENTIONS.md       # Reglas de commits y PRs
    PLAN-crud.md                # Template de plan para CRUDs
    PLAN-bugfix.md              # Template de plan para bugfixes
    PLAN-integration.md         # Template de plan para integraciones
    PLAN-refactor.md            # Template de plan para refactors
```

## Como usar

### Setup inicial

1. Clonar este repositorio
2. Copiar el contenido de `opencode/` a `~/.config/opencode/`
3. Ejecutar `bun install` en `~/.config/opencode/` para instalar dependencias del plugin

### Para cada proyecto nuevo

1. Copiar `opencode/templates/CONVENTIONS.md` a la raiz del proyecto y adaptarlo
2. Ejecutar `/onboard` para que el agente explore el proyecto
3. Usar `/plan <tarea>` para comenzar a trabajar

### Flujo de trabajo diario

```text
/plan implementar feature X        # planificar
/execute                            # ejecutar paso a paso
/apply-feedback cambiar Y           # corregir si hace falta
/status                             # ver progreso
/commit                             # crear commit convencional
/pr                                 # crear pull request
```

### Templates de plan

Los templates en `opencode/templates/PLAN-*.md` sirven como referencia para el planner. No es necesario copiarlos, el `step-builder-agent` ya conoce la estructura general. Son utiles para:
- entender que pasos esperar para cada tipo de tarea
- como referencia manual si quieres armar un plan a mano

## Agentes

| Agente | Rol |
|---|---|
| `step-builder-agent` | Planifica. Hace preguntas de negocio y tecnicas. Genera `PLAN.md` |
| `execution-orchestrator` | Coordina. Ejecuta un paso por vez. Pide review humano |
| `ts-expert-coder` | Implementa. Worker TS/Node/NestJS. Se auto-verifica |

## Convenciones clave

- Siempre usar objetos de dominio (Value Objects) en lugar de primitivos en entidades
- DDD + CQRS: commands por repositorio, queries pueden ir directo a Prisma
- Todo DTO con decoradores de Swagger y class-validator
- Commits en formato Conventional Commits
- Review humano obligatorio entre cada paso de ejecucion
