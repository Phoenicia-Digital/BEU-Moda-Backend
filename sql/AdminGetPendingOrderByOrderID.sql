SELECT 
    user_id,
    ordered_items, 
    total_price, 
    order_time 
FROM 
    pending_orders
WHERE
    order_id = $1;