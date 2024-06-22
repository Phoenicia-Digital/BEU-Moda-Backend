SELECT 
    order_id,
    ordered_items, 
    total_price, 
    order_time 
FROM 
    pending_orders 
WHERE 
    user_id = $1;