import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { OrderOrmEntity } from './infrastructure/typeorm/order.orm-entity.js';
import { PaymentOrmEntity } from './infrastructure/typeorm/payment.orm-entity.js';
import { PaymentEventOrmEntity } from './infrastructure/typeorm/payment-event.orm-entity.js';
import { OutboxMessageOrmEntity } from './infrastructure/typeorm/outbox-message.orm-entity.js';

@Module({
  imports: [
    TypeOrmModule.forFeature([
      OrderOrmEntity,
      PaymentOrmEntity,
      PaymentEventOrmEntity,
      OutboxMessageOrmEntity,
    ]),
  ],
})
export class PaymentCallbackModule {}