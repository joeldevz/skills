# Plan: [<Feature>]

## Goal
[Que capacidad nueva se agrega al sistema y por que importa]

## Business Context
[Usuarios afectados, comportamiento esperado, reglas de negocio, edge cases, criterios de aceptacion del producto]

## Technical Context
[Modulos existentes que se tocan, dependencias, APIs afectadas, consideraciones de rendimiento o seguridad]

## Implementation Steps

### Step 1: Modelar dominio del feature
- **What**: Crear o extender las entidades, Value Objects, y errores de dominio necesarios para el nuevo comportamiento
- **Why**: El dominio define las reglas de negocio antes de conectar infraestructura
- **Where**: `src/contexts/<context>/domain/`
- **Acceptance**: Entidades con factory methods, Value Objects donde hay semantica, errores de dominio especificos. Sin dependencias de infra
- **Status**: [ ] pending

### Step 2: Definir contratos (interfaces/repositorios)
- **What**: Crear o extender interfaces de repositorios, servicios, o adaptadores que el feature necesita
- **Why**: Los contratos permiten implementar la logica de aplicacion sin depender de implementaciones concretas
- **Where**: `src/contexts/<context>/domain/repositories/` o `src/shared/domain/interfaces/`
- **Acceptance**: Interfaces con metodos claros y tipados. Tokens para DI si son nuevos
- **Status**: [ ] pending

### Step 3: Logica de aplicacion (commands/queries/handlers)
- **What**: Crear los commands, queries, y handlers que implementan el comportamiento del feature
- **Why**: La capa de aplicacion orquesta el dominio y los servicios
- **Where**: `src/contexts/<context>/application/commands/`, `queries/`
- **Acceptance**: Handlers usan inyeccion de dependencias, manejan errores de dominio, respetan CQRS
- **Status**: [ ] pending

### Step 4: DTOs y mappers
- **What**: Crear DTOs de request/response y mappers entre dominio y DTOs
- **Why**: Los DTOs son la frontera de serializacion. Separan la representacion externa del modelo interno
- **Where**: `src/contexts/<context>/application/dtos/`, `mappers/`
- **Acceptance**: DTOs con decoradores de validacion y Swagger. Mappers con `fromDomain()` y `toDomain()` cuando aplique
- **Status**: [ ] pending

### Step 5: Infraestructura (persistencia, servicios, adaptadores)
- **What**: Implementar repositorios, servicios externos, jobs de cola, o cualquier pieza de infraestructura necesaria
- **Why**: Conecta el dominio y la aplicacion con el mundo real (DB, APIs, colas)
- **Where**: `src/contexts/<context>/infrastructure/`
- **Acceptance**: Implementa las interfaces del dominio. Manejo de errores robusto. Mappers Prisma si hay persistencia nueva
- **Status**: [ ] pending

### Step 6: Controller / punto de entrada
- **What**: Crear o extender controllers REST, event listeners, o workers que exponen el feature
- **Why**: El punto de entrada conecta al usuario o al sistema externo con la logica de aplicacion
- **Where**: `src/contexts/<context>/infrastructure/controllers/`
- **Acceptance**: Endpoints con `@ApiOperation`, guards, DTOs. Solo inyecta CommandBus/QueryBus
- **Status**: [ ] pending

### Step 7: Tests unitarios
- **What**: Tests para entidades de dominio, handlers, y logica critica del feature
- **Why**: Validar reglas de negocio y flujos antes de verificar la integracion
- **Where**: Co-locados con el codigo: `*.spec.ts`
- **Acceptance**: Cubren happy path, edge cases, errores. Handlers con mocks del repositorio
- **Status**: [ ] pending

### Step 8: Registro de modulo y verificacion final
- **What**: Registrar todos los providers nuevos en el modulo NestJS. Correr build, lint, y tests completos
- **Why**: Sin registro de DI nada funciona. La verificacion final valida que todo esta integrado
- **Where**: `src/contexts/<context>/<context>.module.ts`
- **Acceptance**: Build compila, tests pasan, lint limpio, no hay errores de DI
- **Status**: [ ] pending

## Verification
```bash
npx tsc --noEmit               # sin errores de tipos
pnpm lint                      # sin errores de linting
pnpm test                      # unit tests pasan
pnpm test:e2e                  # e2e tests si aplican
```

## Risks / Notes
- Si el feature necesita migracion de DB, crear la migracion de Prisma antes del Step 5
- Si cruza bounded contexts, definir interfaces compartidas en `src/shared/` antes de empezar
- Si introduce dependencias nuevas (npm packages), validar compatibilidad y licencia
- Considerar si hay necesidad de feature flag o rollout gradual
