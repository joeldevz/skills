# vibe-coding

Configuración de OpenCode para trabajar rápido sin supervisión constante.

Un solo agente. Commands mínimos. El agente decide, ejecuta y entrega.

---

## Diferencia con el modo supervisado

| | Supervisado | Vibe |
|---|---|---|
| Agentes | planner + manager + coder | 1 solo (`vibe`) |
| PLAN.md | Obligatorio | Opcional |
| Review humano | Después de cada paso | No existe |
| Commands | 10+ | 4 |
| Velocidad | Controlada | Máxima |

---

## Setup

### 1. Copiar archivos

```bash
# Copiar opencode.json a tu config global
cp opencode.json ~/.config/opencode/opencode.json

# Copiar commands
cp commands/*.md ~/.config/opencode/commands/
```

### 2. Configurar API key de Context7 (si la usas)

En `~/.config/opencode/opencode.json`, reemplaza `SET_IN_LOCAL_CONFIG` con tu key real.

O bien deshabilita Context7:
```json
"context7": { "enabled": false }
```

---

## Commands

| Command | Qué hace |
|---|---|
| `/do <tarea>` | Ejecuta una tarea completa de principio a fin |
| `/fix <problema>` | Arregla un bug específico |
| `/commit` | Crea un commit con los cambios actuales |
| `/done` | Cierra la sesión y guarda lo aprendido en memoria |

---

## Cuándo usar este modo

- Exploraciones rápidas
- Features pequeños o medianos bien definidos
- Bugfixes
- Proyectos donde confías en el agente

## Cuándo NO usar este modo

- Features grandes con decisiones de arquitectura importantes
- Cuando necesitas visibilidad paso a paso
- Proyectos con alto riesgo de regresión sin tests

En esos casos usa la configuración supervisada en `../opencode/`.

---

## Flujo típico

```
/do agrega endpoint POST /users con validación de email
# El agente lee el repo, implementa, verifica y reporta

/fix el endpoint de login retorna 500 cuando el user no existe
# El agente encuentra la causa, la arregla, verifica

/commit
# Commit con mensaje convencional

/done
# Guarda decisiones en memoria para la próxima sesión
```
