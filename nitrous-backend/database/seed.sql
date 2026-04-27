-- Seed data for the Nitrous PostgreSQL database.
-- These rows mirror the prototype content currently shown in the app.
-- The file is idempotent so it can be executed safely during initialization.

INSERT INTO categories (id, name, slug, icon, live_count, description, color)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'MOTORSPORT', 'motorsport', 'R', 24, 'NASCAR - F1 - Dirt - Rally', 'cyan'),
    ('22222222-2222-2222-2222-222222222222', 'WATER', 'water', 'W', 8, 'Speed Boats - Jet Ski - Surf', 'blue'),
    ('33333333-3333-3333-3333-333333333333', 'AIR & SKY', 'air', 'A', 5, 'Skydive - Air Race - Wing', 'purple'),
    ('44444444-4444-4444-4444-444444444444', 'OFF-ROAD', 'offroad', 'O', 12, 'Dakar - Baja - Enduro', 'orange')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO events (id, title, location, date, time, is_live, category, thumbnail_url)
VALUES
    ('55555555-5555-5555-5555-555555555551', 'Speed Boat Cup - Finals', 'Lake Como - Italy', NOW() + INTERVAL '14 days', '14:00 UTC', FALSE, 'water', NULL),
    ('55555555-5555-5555-5555-555555555552', 'Red Bull Skydive Series - Rd. 3', 'Interlaken Drop Zone - Switzerland', NOW() + INTERVAL '20 days', '11:30 UTC', FALSE, 'air', NULL),
    ('55555555-5555-5555-5555-555555555553', 'Crop Duster Air Racing', 'Bakersfield Airfield - California', NOW() + INTERVAL '26 days', '16:00 UTC', FALSE, 'air', NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO journeys (id, title, category, description, badge, slots_left, date, price, thumbnail_url)
VALUES
    ('66666666-6666-6666-6666-666666666661', 'DAYTONA PIT CREW EXPERIENCE', 'MOTORSPORT - BEHIND THE SCENES', 'Go behind the wall at Daytona 500. Watch pit stops up close, meet the crew chiefs, and ride the pace car on track.', 'EXCLUSIVE', 12, NOW() + INTERVAL '10 days', 2400, NULL),
    ('66666666-6666-6666-6666-666666666662', 'DAKAR DESERT CONVOY', 'RALLY - DESERT EXPEDITION', 'Ride a support vehicle through the Dakar stages. Sleep under the stars, eat with the team, and feel the dust.', 'MEMBERS ONLY', 6, NOW() + INTERVAL '345 days', 5800, NULL),
    ('66666666-6666-6666-6666-666666666663', 'RED BULL TANDEM SKYDIVE', 'AIR - EXTREME SPORT', 'Jump with a Red Bull certified instructor at 15,000ft. Camera-equipped, full debrief, and a story you''ll never forget.', 'LIMITED', 3, NOW() + INTERVAL '20 days', 1200, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO merch_items (id, name, icon, price, category)
VALUES
    ('merch-team-hoodie', 'Team Hoodie', 'H', 89, 'apparel'),
    ('merch-nitrous-cap', 'NITROUS Cap', 'C', 42, 'apparel'),
    ('merch-racing-jacket', 'Racing Jacket', 'J', 189, 'apparel'),
    ('merch-pit-watch', 'Pit Watch', 'W', 249, 'accessories'),
    ('merch-gear-backpack', 'Gear Backpack', 'B', 120, 'accessories'),
    ('merch-drop-keychain', 'Drop Keychain', 'K', 28, 'collectibles')
ON CONFLICT (id) DO NOTHING;

INSERT INTO passes (id, tier, event_name, location, event_date, category, price, perks, spots_left, total_spots, badge, tier_color)
VALUES
    ('pass-daytona-grandstand', 'GRANDSTAND', 'Daytona 500', 'Daytona Beach, FL', NOW() + INTERVAL '30 days', 'motorsport', 299, '["Track access", "Pit lane tour"]'::jsonb, 4, 20, 'LIMITED', '#ff4d4d'),
    ('pass-pit-experience', 'PIT ACCESS', 'F1 Grand Prix', 'Austin, TX', NOW() + INTERVAL '45 days', 'motorsport', 599, '["Pit walk", "Garage access"]'::jsonb, 12, 50, NULL, '#60a5fa')
ON CONFLICT (id) DO NOTHING;
