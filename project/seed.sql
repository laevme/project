-- seed.sql

INSERT INTO users (name, email, phone) VALUES
    ('Аня',   'anya@example.com',   '+7-900-111-22-33'),
    ('Борис', 'boris@example.com',  '+7-900-222-33-44'),
    ('Вера',  'vera@example.com',   '+7-900-333-44-55')
ON CONFLICT (email) DO NOTHING;

INSERT INTO categories (name, slug) VALUES
    ('Учебники',   'books'),
    ('Электроника','electronics'),
    ('Одежда',     'clothes'),
    ('Мебель',     'furniture')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO listings (title, description, price, user_id, category_id) VALUES
    ('Учебник по Go',       'Почти новый, без пометок',       500.00, 1, 1),
    ('Ноутбук Lenovo',      'Б/у, работает отлично',        25000.00, 2, 2),
    ('Зимняя куртка',       'Размер M, тёплая',              3000.00, 3, 3),
    ('Письменный стол',     'Дубовый, 120x60',               4500.00, 1, 4)
ON CONFLICT DO NOTHING;

INSERT INTO messages (body, listing_id, sender_id) VALUES
    ('Ещё продаёте?',          1, 2),
    ('Да, актуально',          1, 1),
    ('Торг возможен?',         2, 3)
ON CONFLICT DO NOTHING;

INSERT INTO favorites (user_id, listing_id) VALUES
    (2, 1),
    (3, 2)
ON CONFLICT DO NOTHING;