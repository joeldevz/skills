# Plan: [CRUD de <Entidad>]

## Goal
[Crear el modulo completo para gestionar <Entidad> con operaciones CRUD]

## Business Context
[Que problema resuelve, quien lo usa, reglas de negocio, permisos, edge cases]

## Technical Context
[Contexto donde vive, entidades relacionadas, dependencias existentes]

## Implementation Steps

### Step 1: Entidad de dominio y Value Objects
- **What**: Crear la entidad de dominio con factory methods (`create`, `reconstitute`), logica de negocio y Value Objects para propiedades con semantica
- **Why**: El dominio es la base de todo. Sin entidad no hay nada mas que implementar
- **Where**: `src/contexts/<context>/domain/entities/<entity>.entity.ts`
- **Acceptance**: La entidad usa Value Objects (no primitivos), tiene constructor privado, factory methods, y metodos de negocio con validaciones
- **Status**: [ ] pending

### Step 2: Errores de dominio
- **What**: Crear errores especificos del contexto extendiendo los base (`NotFoundError`, `ConflictError`, etc.)
- **Why**: Los errores de dominio se mapean automaticamente a HTTP status via `AllExceptionsFilter`
- **Where**: `src/contexts/<context>/domain/errors/<entity>.errors.ts`
- **Acceptance**: Al menos `<Entity>NotFoundError` y errores para invariantes de negocio
- **Status**: [ ] pending

### Step 3: Interfaz de repositorio
- **What**: Definir la interfaz del repositorio y el Symbol token para DI
- **Why**: El dominio define el contrato, la infraestructura lo implementa
- **Where**: `src/contexts/<context>/domain/repositories/<entity>.repository.ts`
- **Acceptance**: Interfaz con metodos necesarios (findById, save, delete, etc.) + token exportado
- **Status**: [ ] pending

### Step 4: Tests de la entidad de dominio
- **What**: Escribir unit tests para la entidad: factory methods, metodos de negocio, validaciones, invariantes
- **Why**: Validar la logica de dominio antes de conectar infraestructura
- **Where**: `src/contexts/<context>/domain/entities/<entity>.entity.spec.ts`
- **Acceptance**: Tests cubren creacion valida, creacion invalida, cada metodo de negocio, y edge cases
- **Status**: [ ] pending

### Step 5: Commands y handlers (Create, Update, Delete)
- **What**: Crear un directorio por command con su command class, handler, y spec
- **Why**: CQRS separa escritura de lectura. Cada operacion de escritura es un command
- **Where**: `src/contexts/<context>/application/commands/create-<entity>/`, `update-<entity>/`, `delete-<entity>/`
- **Acceptance**: Handlers inyectan repositorio via token, usan entidad de dominio, lanzan errores de dominio. Cada handler tiene spec con mocks
- **Status**: [ ] pending

### Step 6: Queries y handlers (List, GetById, etc.)
- **What**: Crear queries para lectura. Pueden usar PrismaService directo o repositorio
- **Why**: Las queries no mutan estado y pueden optimizarse sin pasar por el dominio
- **Where**: `src/contexts/<context>/application/queries/get-<entity>/`, `list-<entities>/`
- **Acceptance**: Handlers retornan DTOs de response. Queries con paginacion usan los helpers de `@/shared`
- **Status**: [ ] pending

### Step 7: DTOs de request y response
- **What**: Crear DTOs con decoradores compuestos (`@ApiStringProperty`, `@ApiUUIDProperty`, etc.) y response DTOs con `fromDomain()` o mappers
- **Why**: Los DTOs son la frontera de serializacion. Aqui si se usan primitivos
- **Where**: `src/contexts/<context>/application/dtos/`
- **Acceptance**: Todo DTO tiene decoradores de Swagger y class-validator. Response DTOs transforman entidad/VO a primitivos
- **Status**: [ ] pending

### Step 8: Implementacion Prisma (repositorio + mapper)
- **What**: Implementar el repositorio contra Prisma y crear mapper de dominio <-> Prisma
- **Why**: Conecta el dominio con la base de datos real
- **Where**: `src/contexts/<context>/infrastructure/persistence/prisma/`
- **Acceptance**: El mapper convierte entre modelo Prisma y entidad de dominio (incluyendo Value Objects). El repositorio implementa la interfaz del dominio
- **Status**: [ ] pending

### Step 9: Controller
- **What**: Crear controller REST. Solo inyecta `CommandBus` y `QueryBus`. Cada endpoint usa `@GuardWithJwt`, `@ApiOperation`, `@ApiBearerAuth`
- **Why**: El controller es solo un adaptador HTTP. No tiene logica de negocio
- **Where**: `src/contexts/<context>/infrastructure/controllers/<entity>.controller.ts`
- **Acceptance**: Endpoints siguen el patron existente. Conversion de primitivos a VOs se hace aqui o en el handler
- **Status**: [ ] pending

### Step 10: Modulo NestJS y registro de providers
- **What**: Registrar controllers, handlers, repositorios (via token + useClass) en el modulo del contexto
- **Why**: NestJS necesita el registro de DI para inyectar todo correctamente
- **Where**: `src/contexts/<context>/<context>.module.ts`
- **Acceptance**: Build compila, todos los providers estan registrados, no hay errores de DI
- **Status**: [ ] pending

### Step 11: E2E tests
- **What**: Crear tests E2E para los endpoints del controller
- **Why**: Validar que la API responde correctamente end-to-end
- **Where**: `test/<entities>.e2e-spec.ts`
- **Acceptance**: Tests cubren happy paths y errores principales. Usan supertest con CommandBus/QueryBus mockeados y guards overrideados
- **Status**: [ ] pending

## Verification
```bash
pnpm test                    # unit tests pasan
pnpm test:e2e                # e2e tests pasan
pnpm lint                    # sin errores de linting
npx tsc --noEmit             # sin errores de tipos
```

## Risks / Notes
- Si la entidad necesita Value Objects nuevos, crearlos en `src/shared/domain/value-objects/` antes del Step 1
- Si el schema de Prisma necesita nuevos modelos, crear la migracion antes del Step 8
- Revisar si hay integracion con otros contextos que requiera interfaces compartidas
