import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
} from 'typeorm';

@Entity('outbox_messages')
export class OutboxMessageOrmEntity {
  @PrimaryGeneratedColumn({ type: 'bigint' })
  id!: string;

  @Column({
    name: 'deduplication_key',
    type: 'varchar',
    length: 120,
  })
  deduplicationKey!: string;

  @Column({
    name: 'event_type',
    type: 'varchar',
    length: 40,
  })
  eventType!: string;

  @Column({
    name: 'aggregate_type',
    type: 'varchar',
    length: 40,
  })
  aggregateType!: string;

  @Column({
    name: 'aggregate_id',
    type: 'varchar',
    length: 40,
  })
  aggregateId!: string;

  @Column({ type: 'json' })
  payload!: Record<string, unknown>;

  @Column({
    type: 'varchar',
    length: 20,
    default: 'PENDING',
  })
  status!: string;

  @CreateDateColumn({
    name: 'created_at',
    type: 'datetime',
    precision: 6,
  })
  createdAt!: Date;

  @Column({
    name: 'published_at',
    type: 'datetime',
    precision: 6,
    nullable: true,
  })
  publishedAt!: Date | null;
}