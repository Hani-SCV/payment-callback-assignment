<?php

/**
 * Legacy reference prepared for this assignment.
 * Provider field names follow public callback shapes; all values are test data.
 * This captures the old behavior and is not intended as the target design.
 */
function normalizeLegacyCallback(string $provider, array $input): ?array
{
    if ($provider === 'TOSS') {
        return [
            'provider' => 'TOSS',
            'payment_id' => $input['orderId'] ?? '',
            'transaction_id' => $input['paymentKey'] ?? '',
            'amount' => (float) ($input['amount'] ?? 0),
            'currency' => 'KRW',
        ];
    }

    if ($provider === 'STRIPE') {
        if (($input['type'] ?? '') !== 'checkout.session.completed') {
            return null;
        }
        $session = $input['data']['object'] ?? [];
        return [
            'provider' => 'STRIPE',
            'payment_id' => $session['client_reference_id'] ?? '',
            'transaction_id' => $session['id'] ?? '',
            // Legacy assumed every Stripe currency had two decimal places.
            'amount' => (float) ($session['amount_total'] ?? 0) / 100,
            'currency' => strtoupper($session['currency'] ?? ''),
        ];
    }

    if ($provider === 'ALIPAY') {
        return [
            'provider' => 'ALIPAY',
            'payment_id' => $input['order_id'] ?? '',
            'transaction_id' => $input['trade_no'] ?? '',
            'amount' => (float) ($input['total_amount'] ?? 0),
            'currency' => 'CNY',
        ];
    }

    throw new InvalidArgumentException('Unsupported provider');
}

function completePaymentCallback(PDO $db, string $provider, array $input): array
{
    $callback = normalizeLegacyCallback($provider, $input);
    if ($callback === null) {
        return ['status' => 200, 'body' => ['result' => 'ignored']];
    }

    $query = $db->prepare(
        'SELECT p.*, o.status AS order_status
           FROM payments p
           JOIN orders o ON o.id = p.order_id
          WHERE p.public_id = :payment_id'
    );
    $query->execute(['payment_id' => $callback['payment_id']]);
    $payment = $query->fetch(PDO::FETCH_ASSOC);

    if (!$payment) {
        return ['status' => 404, 'body' => ['error' => 'payment not found']];
    }

    if ($payment['provider'] !== $callback['provider']
        || $payment['status'] !== 'PENDING'
        || $payment['order_status'] !== 'PAYMENT_PENDING') {
        return ['status' => 409, 'body' => ['error' => 'invalid state']];
    }

    if ((float) $payment['amount'] !== $callback['amount']
        || $payment['currency'] !== $callback['currency']) {
        return ['status' => 409, 'body' => ['error' => 'amount mismatch']];
    }

    $db->prepare(
        "UPDATE payments
            SET status = 'COMPLETED',
                external_transaction_id = :transaction_id,
                completed_at = UTC_TIMESTAMP(6),
                updated_at = UTC_TIMESTAMP(6)
          WHERE id = :id"
    )->execute([
        'transaction_id' => $callback['transaction_id'],
        'id' => $payment['id'],
    ]);

    $db->prepare(
        "UPDATE orders SET status = 'PAID', updated_at = UTC_TIMESTAMP(6)
          WHERE id = :id"
    )->execute(['id' => $payment['order_id']]);

    $db->prepare(
        "INSERT INTO payment_events (payment_id, event_type, payload)
         VALUES (:payment_id, 'PAYMENT_COMPLETED', CAST(:payload AS JSON))"
    )->execute([
        'payment_id' => $payment['id'],
        'payload' => json_encode($input),
    ]);

    return [
        'status' => 200,
        'body' => [
            'result' => 'completed',
            'payment_id' => $payment['public_id'],
            'order_status' => 'PAID',
        ],
    ];
}
