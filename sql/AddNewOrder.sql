INSERT INTO pending_orders (user_id, non_user_billing_address, non_user_full_name, non_user_phone_number, ordered_items, total_price, order_time) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING order_id;