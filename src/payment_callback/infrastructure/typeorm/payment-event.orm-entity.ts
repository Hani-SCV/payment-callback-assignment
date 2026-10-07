import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
} from 'typeorm';

@Entity('payment_events')
export class PaymentEventOrmEntity {
  @PrimaryGeneratedColumn({ type: 'bigint' })
  id!: string;

  @Column({ name: 'payment_id', type: 'bigint' })
  paymentId!: string;

  @Column({ name: 'event_type', type: 'varchar', length: 40 })
  eventType!: string;

  @Column({ type: 'json' })
  payload!: Record<string, unknown>;

  @CreateDateColumn({
    name: 'created_at',
    type: 'datetime',
    precision: 6,
  })
  createdAt!: Date;
}