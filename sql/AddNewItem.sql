INSERT INTO items (name, description, price, quantity, color ) VALUES ($1, $2, $3, $4, $5) RETURNING id;