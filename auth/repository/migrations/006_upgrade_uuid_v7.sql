ALTER TABLE users              ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE sessions           ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE verification_codes ALTER COLUMN id SET DEFAULT uuidv7();
