# Neurox Protocol — Memoria persistente entre sesiones

## Cuándo usar Neurox

- **Al inicio de cada tarea**: `neurox_session_start` + `neurox_context` ANTES de cualquier otra acción
- **Durante la tarea**: `neurox_recall` con queries cortos tipo keyword antes de responder preguntas técnicas o modificar código conocido
- **Al encontrar algo durable**: `neurox_save` inmediatamente (no al final — se puede perder)
- **Al terminar**: `neurox_session_end` con resumen Goal/Discoveries/Accomplished/Next

## Protocolo de inicio (obligatorio)

```
1. neurox_session_start(title, directory, namespace="{project}")
2. neurox_context(namespace="{project}")   ← leer ANTES de explorar el repo
3. Si hay preguntas de identidad/preferencias del usuario:
   neurox_recall(query="nombre preferencia usuario", observation_type="preference")
   → Si no hay resultado: 2-3 búsquedas más con variantes antes de rendirse
```

## Cuándo guardar (triggers obligatorios)

| Evento | observation_type | kind |
|--------|-----------------|------|
| Decisión de arquitectura o diseño | `decision` | `semantic` |
| Bug fix completado (con causa raíz) | `bugfix` | `procedural` |
| Patrón o convención descubierta | `pattern` / `discovery` | `semantic` |
| Usuario corrige el enfoque o da preferencia | `preference` | `procedural` |
| Config de entorno o tool aprendida | `config` | `semantic` |
| Trampa o gotcha encontrada | `gotcha` | `procedural` |

## Formato de contenido al guardar

```
What: [qué se descubrió o decidió]
Why: [por qué importa]
Where: [archivos o módulos relevantes]
Learned: [qué aprender de esto para el futuro]
```

## Reglas críticas

- NUNCA guardar cambios triviales (typos, formato)
- NUNCA guardar info ya en git history
- Usar `topic_key` para temas que evolucionan — mismo `topic_key` = upsert, no duplicado
- Namespace = nombre del directorio del proyecto (ej: `api-core`, `neurox`, `skills`)
- Para memoria cross-project o preferencias personales: omitir namespace o usar `"default"`
- No inferir identidad del usuario desde git history — usar `neurox_recall` con query explícito

## Namespace convention

```
proyecto específico:  namespace="{project-dir-name}"
preferencias globales: namespace="default" o sin namespace
```
