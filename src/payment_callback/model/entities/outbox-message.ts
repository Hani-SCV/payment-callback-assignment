import { OutboxStatus } from '../enums/outbox-status.js';

export class OutboxMessage {
  constructor(
    public readonly id: string,
    public readonly deduplicationKey: string,
    public readonly eventType: string,
    public readonly aggregateType: string,
    public readonly aggregateId: string,
    public readonly payload: Record<string, unknown>,
    public status: OutboxStatus,
    public readonly createdAt: Date,
    public publishedAt: Date | null,
  ) {}
}