export class PaymentEvent {
  constructor(
    public readonly id: string,
    public readonly paymentId: string,
    public readonly eventType: string,
    public readonly payload: Record<string, unknown>,
    public readonly createdAt: Date,
  ) {}
}