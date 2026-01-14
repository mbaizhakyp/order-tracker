-- Assign random first names to users who might have generic or placeholder names
-- We update ALL users to ensure everyone has a nice demo name, or we could target specific ones.
-- For this request, "assign currently registered users random first names" implies a bulk update.
UPDATE users 
SET name = (ARRAY['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank', 'Grace', 'Heidi', 'Ivan', 'Judy', 'Kevin', 'Laura', 'Mike', 'Nancy', 'Oscar', 'Peggy', 'Quentin', 'Ruth', 'Steve', 'Trudy'])[floor(random() * 20 + 1)]
WHERE name IS NULL OR name = '' OR name LIKE '%@%' OR length(name) < 2;
