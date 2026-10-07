import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';

@Entity('payments')
export class PaymentOrmEntity {
  @PrimaryGeneratedColumn({ type: 'bigint' })
  id!: string;

  @Column({ name: 'public_id', type: 'varchar', length: 40 })
  publicId!: string;

  @Column({ name: 'order_id', type: 'bigint' })
  orderId!: string;

  @Column({ type: 'varchar', length: 24 })
  provider!: string;

  @Column({ type: 'varchar', length: 24 })
  status!: string;

  @Column({
    type: 'decimal',
    precision: 12,
    scale: 2,
  })
  amount!: string;

  @Column({ type: 'char', length: 3 })
  currency!: string;

  @Column({
    name: 'external_transaction_id',
    type: 'varchar',
    length: 80,
    nullable: true,
  })
  externalTransactionId!: string | null;

  @Column({
    name: 'cancellation_reason',
    type: 'varchar',
    length: 40,
    nullable: true,
  })
  cancellationReason!: string | null;

  @Column({
    name: 'completed_at',
    type: 'datetime',
    precision: 6,
    nullable: true,
  })
  completedAt!: Date | null;

  @CreateDateColumn({
    name: 'created_at',
    type: 'datetime',
    precision: 6,
  })
  createdAt!: Date;

  @UpdateDateColumn({
    name: 'updated_at',
    type: 'datetime',
    precision: 6,
  })
  updatedAt!: Date;
}