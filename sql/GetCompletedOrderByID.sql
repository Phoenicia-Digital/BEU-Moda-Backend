SELECT user_id, ordered_items, total_price, order_time FROM order_history WHERE order_id = $1;