import { OrderStatus } from '../enums/order-status.js';

export class Order {
  constructor(
    public readonly id: string,
    public readonly publicId: string,
    public readonly customerReference: string,
    public status: OrderStatus,
    public readonly createdAt: Date,
    public updatedAt: Date,
  ) {}
}