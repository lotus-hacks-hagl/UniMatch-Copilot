-- Seed Data for Universities
-- UUIDs generated for testing purposes

INSERT INTO universities (id, name, country, qs_rank, group_tag, ielts_min, sat_required, gpa_expectation_normalized, tuition_usd_per_year, scholarship_available, available_majors, acceptance_rate, crawl_status) VALUES
('b1a0300d-5847-4e00-848e-2e5513229864', 'Harvard University', 'USA', 4, 'Ivy League', 7.5, TRUE, 3.9, 57000, TRUE, ARRAY['Computer Science', 'Business', 'Law', 'Medicine'], 0.04, 'never_crawled'),
('e3b56cd6-3cda-46aa-ac9b-0ab751a02b1f', 'Stanford University', 'USA', 3, 'Private Research', 7.0, TRUE, 3.85, 56000, TRUE, ARRAY['Engineering', 'Computer Science', 'Economics'], 0.05, 'never_crawled'),
('7f912f71-2b02-4fdc-9fac-19ceeffcd9ee', 'Massachusetts Institute of Technology (MIT)', 'USA', 1, 'Private Research', 7.0, TRUE, 3.9, 55000, TRUE, ARRAY['Engineering', 'Physics', 'Computer Science'], 0.04, 'never_crawled'),
('c3514a48-89c0-4f51-a2ac-c9fb27568558', 'University of Oxford', 'UK', 2, 'Russell Group', 7.5, FALSE, 3.8, 45000, TRUE, ARRAY['PPE', 'Law', 'Medicine', 'Literature'], 0.17, 'never_crawled'),
('d1445b23-6ec2-4bf3-b541-118544c01ab8', 'University of Cambridge', 'UK', 2, 'Russell Group', 7.5, FALSE, 3.8, 46000, TRUE, ARRAY['Natural Sciences', 'Mathematics', 'Engineering'], 0.21, 'never_crawled'),
('8975de30-681b-417c-a4ec-ded88147d3aa', 'University of Melbourne', 'Australia', 14, 'Group of Eight', 6.5, FALSE, 3.2, 35000, TRUE, ARRAY['Business', 'Engineering', 'Arts', 'Science'], 0.70, 'never_crawled'),
('4e114099-b1d3-49d7-83eb-cfacbed3339e', 'University of Sydney', 'Australia', 19, 'Group of Eight', 6.5, FALSE, 3.2, 34000, TRUE, ARRAY['Medicine', 'Law', 'Business', 'Engineering'], 0.30, 'never_crawled'),
('f82b8423-ed07-4286-9dbd-bb3ef9c0dff8', 'National University of Singapore (NUS)', 'Singapore', 8, 'Public Research', 6.5, FALSE, 3.5, 25000, TRUE, ARRAY['Computer Science', 'Engineering', 'Business'], 0.05, 'never_crawled'),
('29b1580f-9036-4074-be48-3c3b0df3f0de', 'Tsinghua University', 'China', 14, 'C9 League', 6.5, FALSE, 3.5, 10000, TRUE, ARRAY['Engineering', 'Computer Science', 'Materials Science'], 0.02, 'never_crawled'),
('a37e1320-9cbd-428b-b6fb-a79ff1ef97b3', 'University of Toronto', 'Canada', 21, 'U15', 6.5, FALSE, 3.3, 45000, TRUE, ARRAY['Engineering', 'Life Sciences', 'Computer Science'], 0.43, 'never_crawled');
