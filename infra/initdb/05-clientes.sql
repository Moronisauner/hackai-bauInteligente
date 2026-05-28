-- Tabela auxiliar da POC: associa cada user_id (anônimo, hash) a um nome
-- legível, gerado aleatoriamente no seed. Usada na UI para apresentar a
-- listagem de clientes (RF-01) de forma humana.
--
-- O nome é apenas decorativo: a massa é anonimizada e o nome real do cliente
-- não existe nos dados. Reseed do volume gera nomes diferentes — é esperado.

CREATE TABLE clientes (
    user_id VARCHAR(64) PRIMARY KEY,
    nome    VARCHAR(255) NOT NULL
);

INSERT INTO clientes (user_id, nome)
SELECT
    u.user_id,
    (ARRAY[
        'Ana', 'Bruno', 'Carla', 'Daniel', 'Eduarda', 'Felipe', 'Gabriela', 'Henrique',
        'Isabela', 'João', 'Karina', 'Lucas', 'Mariana', 'Natália', 'Otávio', 'Patrícia',
        'Rafael', 'Sofia', 'Tiago', 'Vanessa'
    ])[floor(random() * 20 + 1)::int]
    || ' ' ||
    (ARRAY[
        'Almeida', 'Barbosa', 'Cardoso', 'Dias', 'Esteves', 'Ferreira', 'Gomes', 'Henriques',
        'Inácio', 'Jardim', 'Lima', 'Mendes', 'Nogueira', 'Oliveira', 'Pereira', 'Queiroz',
        'Ribeiro', 'Santos', 'Teixeira', 'Vieira'
    ])[floor(random() * 20 + 1)::int] AS nome
FROM (SELECT DISTINCT user_id FROM bank_accounts) u;
