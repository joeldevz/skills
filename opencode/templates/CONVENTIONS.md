# Conventions

Este archivo define las convenciones del proyecto. Los agentes de IA y los desarrolladores humanos deben seguir estas reglas al escribir codigo.

## Arquitectura

### Patron principal: DDD + CQRS + Capas

Cada bounded context sigue esta estructura:

```
src/contexts/<context>/
  domain/
    entities/         # Entidades de dominio con logica de negocio
    errors/           # Errores de dominio especificos del contexto
    repositories/     # Interfaces de repositorio + Symbol tokens
  application/
    commands/         # Un directorio por comando (command + handler + spec)
    queries/          # Un directorio por query (query + handler + spec)
    events/           # Event handlers
    dtos/             # DTOs de request/response
    mappers/          # Mappers de dominio a DTO
  infrastructure/
    controllers/      # Controllers REST (solo inyectan CommandBus/QueryBus)
    persistence/      # Implementaciones de repositorios (Prisma)
      prisma/
        mappers/      # Mappers de Prisma a entidad de dominio
    jobs/             # Procesadores de colas (BullMQ)
    services/         # Servicios de infraestructura compartidos
```

### Direccion de dependencias

```
infrastructure -> application -> domain
```

- `domain/` no depende de nada externo
- `application/` depende solo de `domain/`
- `infrastructure/` depende de `application/` y `domain/`
- nunca se importa entre contextos directamente, solo via interfaces compartidas en `shared/`

### Shared kernel

```
src/shared/
  domain/
    enums/            # Enums compartidos (Role, etc.)
    errors/           # DomainError base y subclases
    events/           # Eventos de dominio compartidos
    helpers/          # Utilidades puras (paginacion, busqueda, timezone)
    interfaces/       # Interfaces compartidas entre contextos
    value-objects/    # Value Objects reutilizables
  infrastructure/
    decorators/       # Decoradores compuestos (ApiStringProperty, etc.)
    filters/          # AllExceptionsFilter
    jwt/              # Guards, decoradores de auth, JwtService
    prisma/           # PrismaService, PrismaModule
    redis/            # RedisModule
    throttler/        # Rate limiting
```

## Regla fundamental: objetos de dominio, no primitivos

**Siempre usar Value Objects en lugar de primitivos en entidades de dominio.**

Esto es obligatorio:

| Concepto | Correcto | Incorrecto |
|---|---|---|
| Dinero | `Money.fromNumber(3.33)` | `price: number` |
| Email | `new Email('a@b.com')` | `email: string` |
| Identificador | `new Uuid(id)` | `id: string` |
| Rango de fechas | `DateRange.create(start, end)` | `startDate: Date, endDate: Date` |
| Porcentaje | `Percentage.fromNumber(15)` | `discount: number` |
| Cantidad positiva | `PositiveInteger.create(5)` | `count: number` |

Los Value Objects disponibles estan en `src/shared/domain/value-objects/`.

Si necesitas un concepto nuevo que no existe como VO, crea uno nuevo extendiendo `ValueObject<T>`.

En DTOs si se usan primitivos (son la frontera de serializacion). La conversion se hace en el handler o mapper.

## Naming

### Archivos: kebab-case con sufijo de tipo

```
activate-deal.handler.ts
activate-deal.handler.spec.ts
activate-deal.command.ts
license.entity.ts
license.entity.spec.ts
license.repository.ts
license.errors.ts
prisma-license.repository.ts
license.mapper.ts
money.vo.ts
session-response.dto.ts
all-exceptions.filter.ts
jwt-auth.guard.ts
```

### Clases: PascalCase con sufijo

```
ActivateDealHandler
ActivateDealCommand
LicenseController
PrismaLicenseRepository
LicenseMapper
Money, Email, Uuid          (Value Objects sin sufijo)
LicenseNotFoundError
AllExceptionsFilter
JwtAuthGuard
SessionResponseDto
BookOccasionalLessonRequestDto
```

### Variables: camelCase

```
dealRepository, licenseRepository, commandBus, queryBus
pricePerCredit, currentRemainingCredits
```

### Constantes y tokens: UPPER_SNAKE_CASE

```
LICENSE_REPOSITORY_TOKEN
DEAL_REPOSITORY_TOKEN
ACTIVATE_LICENSES_QUEUE
JWT_COOKIE_NAME
```

### DTOs

- Request: `*Dto` o `*RequestDto`
- Response: `*ResponseDto`
- Query params: `*QueryDto`

## Imports

### Path aliases (obligatorios para imports entre modulos)

```typescript
import { Uuid } from '@/shared/domain/value-objects';
import { GuardWithJwt, Role } from '@/shared/infrastructure/jwt';
import { Deal } from '@/contexts/commercial/domain/entities/deal.entity';
```

### Imports relativos solo dentro del mismo command/query

```typescript
// Dentro de activate-deal/
import { ActivateDealCommand } from './activate-deal.command';
```

### Barrel exports

Usar `index.ts` para paquetes clave:
- `@/shared/domain/value-objects`
- `@/shared/infrastructure/jwt`

## Controllers

### Reglas

- Los controllers solo inyectan `CommandBus` y `QueryBus`, nunca servicios de dominio
- Cada endpoint usa `@GuardWithJwt([Role.X])` para autenticacion y roles
- Cada endpoint tiene `@ApiOperation`, `@ApiTags`, `@ApiBearerAuth`
- La conversion de primitivos a Value Objects se hace en el controller o handler, no en el DTO

```typescript
@Post('book-occasional')
@HttpCode(HttpStatus.CREATED)
@GuardWithJwt([Role.STUDENT])
@ApiBearerAuth()
@ApiOperation({ summary: 'Agendar clase ocasional' })
async bookOccasionalLesson(
  @Author() author: JwtPayload,
  @Body() dto: BookOccasionalLessonRequestDto,
): Promise<BookOccasionalLessonResponseDto> {
  return await this.commandBus.execute(
    new BookOccasionalLessonCommand(author.sub, dto.durationInMinutes, ...),
  );
}
```

## DTOs y validacion

### Reglas

- Todo DTO debe tener decoradores de Swagger (`@ApiStringProperty`, `@ApiNumberProperty`, etc.)
- Todo DTO debe tener decoradores de class-validator
- Usar los decoradores compuestos de `@/shared/infrastructure/decorators/` que combinan Swagger + validacion
- `ValidationPipe` global con `whitelist: true`, `forbidNonWhitelisted: true`, `transform: true`

### Decoradores compuestos disponibles

```typescript
@ApiStringProperty({ minLength: 3, maxLength: 100 })
@ApiUUIDProperty()
@ApiNumberProperty({ min: 0, isInt: true })
@ApiDateProperty()
@ApiEnumProperty(ServiceType)
@ApiBooleanProperty()
@ApiArrayProperty({ type: String, isString: true })
@ApiObjectProperty({ type: AddressDto })
```

## Repositorios

### Patron: interfaz en dominio + implementacion en infraestructura

```typescript
// domain/repositories/license.repository.ts
export interface LicenseRepository {
  findById(id: string): Promise<License | null>;
  save(license: License): Promise<void>;
}
export const LICENSE_REPOSITORY_TOKEN = Symbol('LicenseRepository');

// infrastructure/persistence/prisma/prisma-license.repository.ts
@Injectable()
export class PrismaLicenseRepository implements LicenseRepository { ... }

// module binding
{ provide: LICENSE_REPOSITORY_TOKEN, useClass: PrismaLicenseRepository }
```

### CQRS: cuando usar repositorio vs Prisma directo

- **Commands (escritura)**: siempre via Repository + Mapper
- **Queries (lectura)**: pueden usar `PrismaService` directamente si es mas eficiente

## Errores

### Jerarquia

```
DomainError (abstract)
  BadRequestError
  NotFoundError
  ConflictError
  ForbiddenError
  UnauthorizedError
  PaymentRequiredError
  UnprocessableError
  InternalError
```

### Errores por contexto

Cada contexto define sus propios errores extendiendo los base:

```typescript
// contexts/commercial/domain/errors/license.errors.ts
export class LicenseNotFoundError extends NotFoundError {
  constructor(id: string) {
    super(`License ${id} not found`);
  }
}
```

El `AllExceptionsFilter` convierte `DomainError` a HTTP status automaticamente y genera un `errorCode` derivado del nombre de la clase.

## Entidades de dominio

### Reglas

- Constructor privado
- Factory methods: `static create(...)` para nuevas instancias, `static reconstitute(...)` para rehidratar desde DB
- Logica de negocio como metodos de instancia
- Usar Value Objects para propiedades con semantica (dinero, fechas, ids, emails, etc.)
- Los metodos mutan estado interno y lanzan `DomainError` si violan invariantes

```typescript
export class License {
  pricePerCredit!: Money;
  
  private constructor(data: Partial<License>) {
    Object.assign(this, data);
  }
  
  static create(data: { ... }): License { /* valida + retorna instancia */ }
  static reconstitute(data: ...): License { /* desde DB */ }
  
  activate(): void { /* logica de negocio */ }
  expire(): void { /* logica de negocio */ }
}
```

## Testing

### Framework: Jest 30 con @swc/jest

### Ubicacion

- Unit tests: colocados junto al codigo fuente como `*.spec.ts`
- E2E tests: en `test/*.e2e-spec.ts`

### Comandos

```bash
pnpm test              # unit tests
pnpm test:e2e          # e2e tests
pnpm test -- --watch   # watch mode
```

### Patron de unit tests (handlers)

```typescript
describe('ActivateDealHandler', () => {
  let handler: ActivateDealHandler;
  let dealRepository: jest.Mocked<DealRepository>;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        ActivateDealHandler,
        {
          provide: DEAL_REPOSITORY_TOKEN,
          useValue: { findById: jest.fn(), save: jest.fn() },
        },
      ],
    }).compile();

    handler = module.get(ActivateDealHandler);
    dealRepository = module.get(DEAL_REPOSITORY_TOKEN);
  });

  it('debe activar un deal con licencias PENDING', async () => {
    const deal = Deal.createFromHubspot({ ... });
    dealRepository.findById.mockResolvedValue(deal);

    await handler.execute(new ActivateDealCommand(deal.id, today, endDate));

    expect(dealRepository.save).toHaveBeenCalledTimes(1);
  });
});
```

### Patron de E2E tests

```typescript
describe('Licenses API (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture = await Test.createTestingModule({
      controllers: [LicenseController],
      providers: [
        { provide: CommandBus, useValue: { execute: jest.fn() } },
        { provide: QueryBus, useValue: { execute: jest.fn() } },
      ],
    })
    .overrideGuard(JwtAuthGuard).useValue({ canActivate: () => true })
    .compile();

    app = moduleFixture.createNestApplication();
    app.useGlobalPipes(new ValidationPipe({ whitelist: true, forbidNonWhitelisted: true, transform: true }));
    await app.init();
  });

  it('debe retornar 200', async () => {
    return request(app.getHttpServer()).get('/licenses/group/group-1').expect(200);
  });
});
```

### Reglas de testing

- Mockear repositorios con `jest.fn()` y `jest.Mocked<T>`
- Construir entidades con factory methods (`Deal.createFromHubspot(...)`, `License.create(...)`)
- Usar Value Objects reales en tests (`Money.fromNumber(3.33)`)
- No mockear entidades de dominio, mockear solo infraestructura (repos, buses, servicios externos)
- E2E tests usan `supertest`, mockean `CommandBus`/`QueryBus` y overridean guards

## Stack tecnico

- **Runtime**: Node.js con NestJS 11 sobre Fastify
- **ORM**: Prisma 7 con multi-schema y driver adapter (`@prisma/adapter-pg`)
- **Compilacion**: SWC (build y tests)
- **Auth**: JWT en cookie + guards de roles
- **Colas**: BullMQ con Redis
- **Logging**: nestjs-pino + pino-seq
- **Validacion de env**: Joi
- **Linter**: ESLint 9 flat config con boundaries plugin
- **Formatter**: Prettier (single quotes, trailing commas)
- **Pre-commit**: Husky + lint-staged (eslint + prettier + jest --findRelatedTests)
- **DB**: PostgreSQL 15
- **Cache/Rate limiting**: Redis

## Reglas inquebrantables

1. Siempre usar objetos de dominio (Value Objects) en lugar de primitivos en entidades
2. No importar entre contextos directamente, solo via interfaces compartidas
3. Commands van por Repository + Mapper, queries pueden ir directo a Prisma
4. Todo DTO debe tener decoradores de Swagger y class-validator
5. Todo handler y entidad deben tener tests unitarios
