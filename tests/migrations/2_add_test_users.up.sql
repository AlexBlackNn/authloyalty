INSERT INTO users(email,pass_hash,is_admin) --password -> test
VALUES ('admin@test.com','$2a$10$thBhIpjEmH22GNr9dxhbbeMwnG16sIATjtNR6vahFUhy7wf0r58NC','true')
ON CONFLICT DO NOTHING;

INSERT INTO users(email,pass_hash,is_admin) --password -> test
VALUES ('user@test.com','$2a$10$thBhIpjEmH22GNr9dxhbbeMwnG16sIATjtNR6vahFUhy7wf0r58NC','false')
ON CONFLICT DO NOTHING;