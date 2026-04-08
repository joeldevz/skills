---
name: nestjs-patterns
description: NestJS patterns for DDD + CQRS architecture including entities, Value Objects, commands, queries, repositories, controllers, guards, interceptors, pipes, BullMQ jobs, and Prisma integration. Use when implementing backend features in NestJS projects that follow Domain-Driven Design.
---

# NestJS DDD + CQRS Patterns

Reference patterns for building NestJS applications with Domain-Driven Design, CQRS, and clean architecture. These patterns match the conventions defined in `CONVENTIONS.md`.

## When to Use This Skill

- Implementing new bounded contexts or modules
- Creating domain entities with Value Objects
- Writing command/query handlers
- Building controllers that follow CQRS
- Creating Prisma repository implementations
- Adding guards, pipes, or interceptors
- Setting up BullMQ job processors
- Writing tests for handlers and entities

## Architecture Overview

```
src/contexts/<context>/
  domain/           → Entities, Value Objects, Errors, Repository interfaces
  application/      → Commands, Queries, Handlers, DTOs, Mappers, Events
  infrastructure/   → Controllers, Prisma repos, Jobs, External services
```

**Dependency direction**: `infrastructure → application → domain`

---

## 1. Domain Entity

Entities have private constructors, factory methods, and use Value Objects instead of primitives.

```typescript
import { BaseEntity } from '@/shared/domain/entities/base-entity';
import { Uuid } from '@/shared/domain/value-objects/uuid';
import { Money } from '@/shared/domain/value-objects/money';
import { InvoiceStatus } from './invoice-status.enum';
import { InvalidInvoiceError } from '../errors/invoice.errors';

export class Invoice extends BaseEntity {
  private constructor(
    id: Uuid,
    private readonly clientId: Uuid,
    private amount: Money,
    private status: InvoiceStatus,
    private readonly issuedAt: Date,
    private paidAt: Date | null,
  ) {
    super(id);
  }

  // Factory: create new
  static create(props: {
    clientId: Uuid;
    amount: Money;
  }): Invoice {
    return new Invoice(
      Uuid.generate(),
      props.clientId,
      props.amount,
      InvoiceStatus.PENDING,
      new Date(),
      null,
    );
  }

  // Factory: reconstitute from persistence
  static reconstitute(props: {
    id: Uuid;
    clientId: Uuid;
    amount: Money;
    status: InvoiceStatus;
    issuedAt: Date;
    paidAt: Date | null;
  }): Invoice {
    return new Invoice(
      props.id,
      props.clientId,
      props.amount,
      props.status,
      props.issuedAt,
      props.paidAt,
    );
  }

  // Business logic
  markAsPaid(): void {
    if (this.status !== InvoiceStatus.PENDING) {
      throw new InvalidInvoiceError(
        `Cannot mark invoice ${this.id.value} as paid: status is ${this.status}`,
      );
    }
    this.status = InvoiceStatus.PAID;
    this.paidAt = new Date();
  }

  // Getters (expose Value Objects, not primitives)
  get clientIdValue(): Uuid { return this.clientId; }
  get amountValue(): Money { return this.amount; }
  get statusValue(): InvoiceStatus { return this.status; }
  get issuedAtValue(): Date { return this.issuedAt; }
  get paidAtValue(): Date | null { return this.paidAt; }
}
```

**Rules:**
- Constructor is always `private`
- Two factory methods: `create` (new) and `reconstitute` (from DB)
- Business logic lives in methods, not in handlers
- Use Value Objects for ids, money, emails, dates with semantics
- Getters expose Value Objects, not primitives

---

## 2. Domain Errors

```typescript
import { NotFoundError, ConflictError } from '@/shared/domain/errors';

export class InvoiceNotFoundError extends NotFoundError {
  constructor(id: string) {
    super(`Invoice with id ${id} not found`);
  }
}

export class InvoiceAlreadyPaidError extends ConflictError {
  constructor(id: string) {
    super(`Invoice ${id} is already paid`);
  }
}
```

**Rules:**
- Extend base errors from `@/shared/domain/errors`
- `AllExceptionsFilter` maps these to HTTP status codes automatically
- Include the entity ID in the message for debugging

---

## 3. Repository Interface + Token

```typescript
import { Invoice } from '../entities/invoice.entity';
import { Uuid } from '@/shared/domain/value-objects/uuid';

export const INVOICE_REPOSITORY = Symbol('INVOICE_REPOSITORY');

export interface InvoiceRepository {
  findById(id: Uuid): Promise<Invoice | null>;
  save(invoice: Invoice): Promise<void>;
  delete(id: Uuid): Promise<void>;
  findByClientId(clientId: Uuid): Promise<Invoice[]>;
}
```

**Rules:**
- Interface + Symbol in the same file
- Methods receive/return domain objects (not Prisma models)
- The Symbol is used for DI registration

---

## 4. Command + Handler

```typescript
// create-invoice.command.ts
export class CreateInvoiceCommand {
  constructor(
    public readonly clientId: string,
    public readonly amount: number,
    public readonly currency: string,
  ) {}
}

// create-invoice.handler.ts
import { CommandHandler, ICommandHandler } from '@nestjs/cqrs';
import { Inject } from '@nestjs/common';
import { CreateInvoiceCommand } from './create-invoice.command';
import { INVOICE_REPOSITORY, InvoiceRepository } from '../../domain/repositories/invoice.repository';
import { Invoice } from '../../domain/entities/invoice.entity';
import { Uuid } from '@/shared/domain/value-objects/uuid';
import { Money } from '@/shared/domain/value-objects/money';

@CommandHandler(CreateInvoiceCommand)
export class CreateInvoiceHandler implements ICommandHandler<CreateInvoiceCommand> {
  constructor(
    @Inject(INVOICE_REPOSITORY)
    private readonly invoiceRepo: InvoiceRepository,
  ) {}

  async execute(command: CreateInvoiceCommand): Promise<string> {
    const invoice = Invoice.create({
      clientId: Uuid.fromString(command.clientId),
      amount: Money.fromNumber(command.amount, command.currency),
    });

    await this.invoiceRepo.save(invoice);
    return invoice.id.value;
  }
}
```

**Rules:**
- Commands carry primitives (they come from the controller/DTO layer)
- Handlers convert primitives to Value Objects
- Handlers inject repositories via Symbol token
- Handlers call domain methods, not raw DB operations

---

## 5. Query + Handler

```typescript
// get-invoice.query.ts
export class GetInvoiceQuery {
  constructor(public readonly id: string) {}
}

// get-invoice.handler.ts
import { QueryHandler, IQueryHandler } from '@nestjs/cqrs';
import { Inject } from '@nestjs/common';
import { GetInvoiceQuery } from './get-invoice.query';
import { INVOICE_REPOSITORY, InvoiceRepository } from '../../domain/repositories/invoice.repository';
import { InvoiceNotFoundError } from '../../domain/errors/invoice.errors';
import { Uuid } from '@/shared/domain/value-objects/uuid';
import { InvoiceResponseDto } from '../../application/dtos/invoice-response.dto';

@QueryHandler(GetInvoiceQuery)
export class GetInvoiceHandler implements IQueryHandler<GetInvoiceQuery> {
  constructor(
    @Inject(INVOICE_REPOSITORY)
    private readonly invoiceRepo: InvoiceRepository,
  ) {}

  async execute(query: GetInvoiceQuery): Promise<InvoiceResponseDto> {
    const invoice = await this.invoiceRepo.findById(Uuid.fromString(query.id));
    if (!invoice) {
      throw new InvoiceNotFoundError(query.id);
    }
    return InvoiceResponseDto.fromDomain(invoice);
  }
}
```

**Rules:**
- Queries can return DTOs directly (no need to go through domain if it's just reading)
- For complex read models, queries can use PrismaService directly
- Always handle not-found cases with domain errors

---

## 6. DTOs

```typescript
// create-invoice-request.dto.ts
import { ApiStringProperty, ApiNumberProperty } from '@/shared/infrastructure/decorators';

export class CreateInvoiceRequestDto {
  @ApiStringProperty({ description: 'Client UUID' })
  clientId: string;

  @ApiNumberProperty({ description: 'Amount in cents' })
  amount: number;

  @ApiStringProperty({ description: 'Currency code (ISO 4217)' })
  currency: string;
}

// invoice-response.dto.ts
import { Invoice } from '../../domain/entities/invoice.entity';

export class InvoiceResponseDto {
  id: string;
  clientId: string;
  amount: number;
  currency: string;
  status: string;
  issuedAt: string;
  paidAt: string | null;

  static fromDomain(invoice: Invoice): InvoiceResponseDto {
    const dto = new InvoiceResponseDto();
    dto.id = invoice.id.value;
    dto.clientId = invoice.clientIdValue.value;
    dto.amount = invoice.amountValue.cents;
    dto.currency = invoice.amountValue.currency;
    dto.status = invoice.statusValue;
    dto.issuedAt = invoice.issuedAtValue.toISOString();
    dto.paidAt = invoice.paidAtValue?.toISOString() ?? null;
    return dto;
  }
}
```

**Rules:**
- Request DTOs use primitives + validation decorators
- Response DTOs have `static fromDomain()` that converts Value Objects to primitives
- Use the project's custom API decorators (`@ApiStringProperty`, etc.)

---

## 7. Controller

```typescript
import { Controller, Post, Get, Body, Param } from '@nestjs/common';
import { CommandBus, QueryBus } from '@nestjs/cqrs';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { GuardWithJwt } from '@/shared/infrastructure/decorators';
import { CreateInvoiceCommand } from '../../application/commands/create-invoice/create-invoice.command';
import { GetInvoiceQuery } from '../../application/queries/get-invoice/get-invoice.query';
import { CreateInvoiceRequestDto } from '../../application/dtos/create-invoice-request.dto';
import { InvoiceResponseDto } from '../../application/dtos/invoice-response.dto';

@ApiTags('Invoices')
@ApiBearerAuth()
@Controller('invoices')
export class InvoiceController {
  constructor(
    private readonly commandBus: CommandBus,
    private readonly queryBus: QueryBus,
  ) {}

  @Post()
  @GuardWithJwt()
  @ApiOperation({ summary: 'Create a new invoice' })
  async create(@Body() dto: CreateInvoiceRequestDto): Promise<{ id: string }> {
    const id = await this.commandBus.execute(
      new CreateInvoiceCommand(dto.clientId, dto.amount, dto.currency),
    );
    return { id };
  }

  @Get(':id')
  @GuardWithJwt()
  @ApiOperation({ summary: 'Get invoice by ID' })
  async getById(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.queryBus.execute(new GetInvoiceQuery(id));
  }
}
```

**Rules:**
- Controllers ONLY inject `CommandBus` and `QueryBus`
- No business logic in controllers
- Use `@GuardWithJwt()`, `@ApiOperation`, `@ApiBearerAuth`
- Convert DTO to command/query, return the result

---

## 8. Prisma Repository Implementation

```typescript
import { Injectable } from '@nestjs/common';
import { PrismaService } from '@/shared/infrastructure/persistence/prisma.service';
import { InvoiceRepository } from '../../../domain/repositories/invoice.repository';
import { Invoice } from '../../../domain/entities/invoice.entity';
import { Uuid } from '@/shared/domain/value-objects/uuid';
import { InvoicePrismaMapper } from './mappers/invoice-prisma.mapper';

@Injectable()
export class PrismaInvoiceRepository implements InvoiceRepository {
  constructor(private readonly prisma: PrismaService) {}

  async findById(id: Uuid): Promise<Invoice | null> {
    const record = await this.prisma.invoice.findUnique({
      where: { id: id.value },
    });
    return record ? InvoicePrismaMapper.toDomain(record) : null;
  }

  async save(invoice: Invoice): Promise<void> {
    const data = InvoicePrismaMapper.toPrisma(invoice);
    await this.prisma.invoice.upsert({
      where: { id: data.id },
      create: data,
      update: data,
    });
  }

  async delete(id: Uuid): Promise<void> {
    await this.prisma.invoice.delete({
      where: { id: id.value },
    });
  }

  async findByClientId(clientId: Uuid): Promise<Invoice[]> {
    const records = await this.prisma.invoice.findMany({
      where: { clientId: clientId.value },
    });
    return records.map(InvoicePrismaMapper.toDomain);
  }
}
```

---

## 9. Prisma Mapper

```typescript
import { Invoice as PrismaInvoice } from '@prisma/client';
import { Invoice } from '../../../../domain/entities/invoice.entity';
import { Uuid } from '@/shared/domain/value-objects/uuid';
import { Money } from '@/shared/domain/value-objects/money';

export class InvoicePrismaMapper {
  static toDomain(record: PrismaInvoice): Invoice {
    return Invoice.reconstitute({
      id: Uuid.fromString(record.id),
      clientId: Uuid.fromString(record.clientId),
      amount: Money.fromNumber(record.amount, record.currency),
      status: record.status as InvoiceStatus,
      issuedAt: record.issuedAt,
      paidAt: record.paidAt,
    });
  }

  static toPrisma(invoice: Invoice): PrismaInvoice {
    return {
      id: invoice.id.value,
      clientId: invoice.clientIdValue.value,
      amount: invoice.amountValue.cents,
      currency: invoice.amountValue.currency,
      status: invoice.statusValue,
      issuedAt: invoice.issuedAtValue,
      paidAt: invoice.paidAtValue,
    };
  }
}
```

---

## 10. Module Registration

```typescript
import { Module } from '@nestjs/common';
import { CqrsModule } from '@nestjs/cqrs';
import { InvoiceController } from './infrastructure/controllers/invoice.controller';
import { CreateInvoiceHandler } from './application/commands/create-invoice/create-invoice.handler';
import { GetInvoiceHandler } from './application/queries/get-invoice/get-invoice.handler';
import { INVOICE_REPOSITORY } from './domain/repositories/invoice.repository';
import { PrismaInvoiceRepository } from './infrastructure/persistence/prisma/prisma-invoice.repository';

const CommandHandlers = [CreateInvoiceHandler];
const QueryHandlers = [GetInvoiceHandler];

@Module({
  imports: [CqrsModule],
  controllers: [InvoiceController],
  providers: [
    ...CommandHandlers,
    ...QueryHandlers,
    {
      provide: INVOICE_REPOSITORY,
      useClass: PrismaInvoiceRepository,
    },
  ],
})
export class InvoiceModule {}
```

---

## 11. Handler Unit Test

```typescript
import { Test } from '@nestjs/testing';
import { CreateInvoiceHandler } from './create-invoice.handler';
import { CreateInvoiceCommand } from './create-invoice.command';
import { INVOICE_REPOSITORY } from '../../../domain/repositories/invoice.repository';

describe('CreateInvoiceHandler', () => {
  let handler: CreateInvoiceHandler;
  let mockRepo: jest.Mocked<any>;

  beforeEach(async () => {
    mockRepo = {
      save: jest.fn().mockResolvedValue(undefined),
      findById: jest.fn(),
      delete: jest.fn(),
      findByClientId: jest.fn(),
    };

    const module = await Test.createTestingModule({
      providers: [
        CreateInvoiceHandler,
        { provide: INVOICE_REPOSITORY, useValue: mockRepo },
      ],
    }).compile();

    handler = module.get(CreateInvoiceHandler);
  });

  it('should create an invoice and return its id', async () => {
    const command = new CreateInvoiceCommand(
      '550e8400-e29b-41d4-a716-446655440000',
      10000,
      'USD',
    );

    const result = await handler.execute(command);

    expect(result).toBeDefined();
    expect(typeof result).toBe('string');
    expect(mockRepo.save).toHaveBeenCalledTimes(1);

    const savedInvoice = mockRepo.save.mock.calls[0][0];
    expect(savedInvoice.clientIdValue.value).toBe(command.clientId);
    expect(savedInvoice.amountValue.cents).toBe(command.amount);
  });
});
```

---

## 12. BullMQ Job Processor

```typescript
import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Job } from 'bullmq';
import { CommandBus } from '@nestjs/cqrs';

@Processor('invoice-queue')
export class InvoiceJobProcessor extends WorkerHost {
  constructor(private readonly commandBus: CommandBus) {
    super();
  }

  async process(job: Job<{ invoiceId: string; action: string }>): Promise<void> {
    switch (job.data.action) {
      case 'send-reminder':
        await this.commandBus.execute(
          new SendInvoiceReminderCommand(job.data.invoiceId),
        );
        break;
      default:
        throw new Error(`Unknown job action: ${job.data.action}`);
    }
  }
}
```

**Rules:**
- Job processors use `CommandBus` to trigger domain logic
- No direct DB access in processors
- Handle unknown actions explicitly

---

## 13. Guard Pattern

```typescript
import { Injectable, CanActivate, ExecutionContext } from '@nestjs/common';
import { Reflector } from '@nestjs/core';

@Injectable()
export class RolesGuard implements CanActivate {
  constructor(private readonly reflector: Reflector) {}

  canActivate(context: ExecutionContext): boolean {
    const requiredRoles = this.reflector.get<string[]>('roles', context.getHandler());
    if (!requiredRoles) return true;

    const request = context.switchToHttp().getRequest();
    const user = request.user;
    return requiredRoles.some((role) => user.roles?.includes(role));
  }
}
```

---

## Quick Reference

| Layer | What lives here | Can depend on |
|-------|----------------|---------------|
| Domain | Entities, VOs, Errors, Repo interfaces | Nothing external |
| Application | Commands, Queries, Handlers, DTOs, Mappers | Domain |
| Infrastructure | Controllers, Prisma repos, Jobs, Services | Application + Domain |

| Pattern | File location |
|---------|--------------|
| Entity | `domain/entities/<name>.entity.ts` |
| Error | `domain/errors/<name>.errors.ts` |
| Repo interface | `domain/repositories/<name>.repository.ts` |
| Command | `application/commands/<action>/` |
| Query | `application/queries/<action>/` |
| DTO | `application/dtos/` |
| Controller | `infrastructure/controllers/` |
| Prisma repo | `infrastructure/persistence/prisma/` |
| Prisma mapper | `infrastructure/persistence/prisma/mappers/` |
| Job | `infrastructure/jobs/` |
| Test | Co-located as `*.spec.ts` |

## Neurox Memory (obligatorio)

Esta skill DEBE usar Neurox para memoria persistente:
- **Al iniciar**: `neurox_recall(query="nestjs patterns conventions {module}")` — buscar patrones previos del proyecto
- **Cross-namespace**: `neurox_recall(query="nestjs DDD CQRS patterns")` sin namespace — inteligencia de otros proyectos
- **Al descubrir convenciones**: `neurox_save(observation_type="pattern", ...)` inmediatamente
- **Al tomar decisiones de arquitectura**: `neurox_save(observation_type="decision", ...)` con justificación
- Si no tienes acceso a Neurox tools, documenta en tu output qué información guardar.
