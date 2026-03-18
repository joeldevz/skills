# Plan: [Integracion con <servicio/sistema externo>]

## Goal
[Que sistema se integra, que datos fluyen, en que direccion]

## Business Context
[Por que se necesita la integracion, que proceso de negocio habilita, frecuencia de uso, SLAs esperados]

## Technical Context
[API del servicio externo, autenticacion, formato de datos, rate limits, modulos afectados del proyecto]

## Implementation Steps

### Step 1: Interfaz del servicio externo
- **What**: Definir la interfaz y el Symbol token en el shared kernel o en el dominio del contexto
- **Why**: El dominio no debe depender de implementaciones externas
- **Where**: `src/shared/domain/interfaces/` o `src/contexts/<context>/domain/`
- **Acceptance**: Interfaz con metodos claros, tipos de request/response definidos, token exportado
- **Status**: [ ] pending

### Step 2: DTOs de la integracion
- **What**: Crear DTOs para los payloads que se envian/reciben del servicio externo
- **Why**: Separar la forma de los datos externos de los objetos de dominio
- **Where**: `src/contexts/<context>/application/dtos/`
- **Acceptance**: DTOs con validacion si reciben datos externos (webhooks). Documentados con Swagger si son parte de la API
- **Status**: [ ] pending

### Step 3: Implementacion del adaptador
- **What**: Crear la clase que implementa la interfaz y se comunica con el servicio externo (HTTP, SDK, etc.)
- **Why**: Aislar la logica de comunicacion externa en infraestructura
- **Where**: `src/contexts/<context>/infrastructure/services/` o `src/contexts/<context>/infrastructure/adapters/`
- **Acceptance**: Manejo de errores robusto (timeouts, retries, errores HTTP). Logging de requests/responses
- **Status**: [ ] pending

### Step 4: Command/Query handlers
- **What**: Crear los handlers que usan el adaptador para procesar la integracion
- **Why**: La logica de orquestacion vive en la capa de aplicacion
- **Where**: `src/contexts/<context>/application/commands/` o `queries/`
- **Acceptance**: Handlers inyectan el servicio via token. Manejan errores del servicio externo con errores de dominio
- **Status**: [ ] pending

### Step 5: Controller o event listener
- **What**: Crear el punto de entrada (endpoint para webhooks, controller para API, o event handler)
- **Why**: Conectar el trigger externo con el handler interno
- **Where**: `src/contexts/<context>/infrastructure/controllers/`
- **Acceptance**: Si es webhook: validacion de firma/token. Endpoint documentado con Swagger
- **Status**: [ ] pending

### Step 6: Tests unitarios y de integracion
- **What**: Unit tests para handlers (mockeando el adaptador). Test del adaptador si es critico
- **Why**: Validar la logica sin depender del servicio externo real
- **Where**: `src/contexts/<context>/application/commands/*/handler.spec.ts`
- **Acceptance**: Tests cubren happy path, errores del servicio externo, y edge cases
- **Status**: [ ] pending

### Step 7: Registro en modulo
- **What**: Registrar adaptador, handlers, y controller en el modulo NestJS
- **Why**: DI necesita los providers registrados
- **Where**: `src/contexts/<context>/<context>.module.ts`
- **Acceptance**: Build compila, DI funciona, no hay errores
- **Status**: [ ] pending

## Verification
```bash
pnpm test                    # unit tests pasan
pnpm lint                    # sin errores
npx tsc --noEmit             # sin errores de tipos
```

## Risks / Notes
- Documentar las credenciales/API keys necesarias en `.env.example`
- Considerar circuit breaker o retry policy si el servicio externo puede fallar
- Si el servicio tiene rate limits, documentar y respetar
- Considerar si se necesita cola (BullMQ) para procesamiento asincrono
