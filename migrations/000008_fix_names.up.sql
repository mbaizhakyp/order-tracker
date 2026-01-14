-- Fix specific user names as requested
UPDATE users SET name = 'Margulan' WHERE email = 'mbaizhakyp@gmail.com';

-- Fix generic "Test Courier" to a real name if it exists (assuming this is the user user is complaining about)
UPDATE users SET name = 'Shopper Sam' WHERE name = 'Test Courier';
