-- +goose Up
-- Ajustes del negocio: una sola fila (cada base de datos es de un único
-- negocio). Se pre-inserta la fila id=1 para que Obtener/Guardar no tengan
-- que lidiar con "aún no existe".
CREATE TABLE configuracion (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	nombre_negocio VARCHAR(80) NOT NULL DEFAULT ''
);
INSERT INTO configuracion (id, nombre_negocio) VALUES (1, '');

-- +goose Down
DROP TABLE configuracion;
