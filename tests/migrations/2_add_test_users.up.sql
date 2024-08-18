INSERT INTO users(uuid, email,pass_hash,is_admin) --password -> test
VALUES ('22f23689-9b67-4ef9-a693-5ef2d18ee111', 'admin@test.com','$2a$10$thBhIpjEmH22GNr9dxhbbeMwnG16sIATjtNR6vahFUhy7wf0r58NC','true')
ON CONFLICT DO NOTHING;

INSERT INTO users(uuid, email,pass_hash,is_admin) --password -> test
VALUES ('7c2ab9ec-bddf-43ff-96a5-ff1e0785c909', 'user@test.com','$2a$10$thBhIpjEmH22GNr9dxhbbeMwnG16sIATjtNR6vahFUhy7wf0r58NC','false')
ON CONFLICT DO NOTHING;