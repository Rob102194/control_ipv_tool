-- Esquema real de la versión Python 0.1.0 (Fase 0).
-- Generado por backend/tests/dump_schema.py a partir de
-- backend/src/infrastructure/db/models.py.
--
-- Estado final tras las migraciones Alembic 7a5988a22b87 y b29f7ae8b140:
-- inventario_diario.comentario EXISTE, modelo_ipv.comentario NO.
--
-- Los DEFAULT de columnas (activa=1, floats=0.0) viven solo en la capa ORM.
-- En Go los fija la capa de dominio/repositorio explícitamente.

CREATE TABLE areas (
	id VARCHAR(36) NOT NULL, 
	nombre VARCHAR(50) NOT NULL, 
	codigo VARCHAR(10), 
	PRIMARY KEY (id), 
	UNIQUE (nombre)
);

CREATE TABLE historial_cambios (
	id VARCHAR(36) NOT NULL, 
	entidad_tipo VARCHAR(50) NOT NULL, 
	entidad_id VARCHAR(36) NOT NULL, 
	campo_modificado VARCHAR(50) NOT NULL, 
	valor_anterior VARCHAR(255), 
	valor_nuevo VARCHAR(255), 
	fecha_cambio DATETIME, 
	PRIMARY KEY (id)
);

CREATE TABLE ingredientes (
	id VARCHAR(36) NOT NULL, 
	receta_id VARCHAR(36) NOT NULL, 
	producto_id VARCHAR(36) NOT NULL, 
	area_id VARCHAR(36) NOT NULL, 
	cantidad FLOAT NOT NULL, 
	PRIMARY KEY (id), 
	FOREIGN KEY(receta_id) REFERENCES recetas (id), 
	FOREIGN KEY(producto_id) REFERENCES productos (id), 
	FOREIGN KEY(area_id) REFERENCES areas (id)
);

CREATE TABLE inventario_diario (
	id VARCHAR(36) NOT NULL, 
	fecha DATE NOT NULL, 
	area_id VARCHAR(36) NOT NULL, 
	producto_id VARCHAR(36) NOT NULL, 
	inicio FLOAT, 
	entradas FLOAT, 
	consumo FLOAT, 
	merma FLOAT, 
	otras_salidas FLOAT, 
	final_fisico FLOAT, 
	final_teorico FLOAT, 
	diferencia FLOAT, 
	comentario TEXT, 
	PRIMARY KEY (id), 
	CONSTRAINT _fecha_area_producto_uc UNIQUE (fecha, area_id, producto_id), 
	FOREIGN KEY(area_id) REFERENCES areas (id), 
	FOREIGN KEY(producto_id) REFERENCES productos (id)
);

CREATE TABLE modelo_ipv (
	id VARCHAR(36) NOT NULL, 
	area_id VARCHAR(36) NOT NULL, 
	producto_id VARCHAR(36) NOT NULL, 
	orden INTEGER NOT NULL, 
	PRIMARY KEY (id), 
	CONSTRAINT _area_producto_uc UNIQUE (area_id, producto_id), 
	FOREIGN KEY(area_id) REFERENCES areas (id), 
	FOREIGN KEY(producto_id) REFERENCES productos (id)
);

CREATE TABLE movimientos (
	id VARCHAR(36) NOT NULL, 
	tipo VARCHAR(7) NOT NULL, 
	producto_id VARCHAR(36) NOT NULL, 
	area_id VARCHAR(36) NOT NULL, 
	cantidad FLOAT NOT NULL, 
	fecha DATETIME, 
	motivo VARCHAR(13), 
	comentarios TEXT, 
	PRIMARY KEY (id), 
	FOREIGN KEY(producto_id) REFERENCES productos (id), 
	FOREIGN KEY(area_id) REFERENCES areas (id)
);

CREATE TABLE productos (
	id VARCHAR(36) NOT NULL, 
	nombre VARCHAR(100) NOT NULL, 
	unidad_medida VARCHAR(10) NOT NULL, 
	PRIMARY KEY (id), 
	UNIQUE (nombre)
);

CREATE TABLE recetas (
	id VARCHAR(36) NOT NULL, 
	nombre VARCHAR(100) NOT NULL, 
	activa BOOLEAN, 
	PRIMARY KEY (id), 
	UNIQUE (nombre)
);

CREATE TABLE ventas (
	id VARCHAR(36) NOT NULL, 
	receta_nombre VARCHAR(100) NOT NULL, 
	cantidad INTEGER NOT NULL, 
	fecha DATE NOT NULL, 
	PRIMARY KEY (id)
);

CREATE INDEX idx_inventario_fecha ON inventario_diario (fecha);

CREATE INDEX idx_venta_fecha ON ventas (fecha);

