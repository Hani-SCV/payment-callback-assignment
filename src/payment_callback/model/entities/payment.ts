import { PaymentProvider } from '../enums/payment-provider.js';
import { PaymentStatus } from '../enums/payment-status.js';

export class Payment {
  constructor(
    public readonly id: string,
    public readonly orderId: string,
    public readonly provider: PaymentProvider,
    public readonly amount: string,
    public readonly currency: string,
    public status: PaymentStatus,
    public externalTransactionId: string | null,
    public completedAt: Date | null,
  ) {}
}