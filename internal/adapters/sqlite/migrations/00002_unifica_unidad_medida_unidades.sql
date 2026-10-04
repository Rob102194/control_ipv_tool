-- +goose Up
-- Unifica la unidad de medida "U" en "UNIDADES": coexistían como dos
-- etiquetas distintas para el mismo concepto (88 productos con "UNIDADES"
-- frente a solo 2 con "U"), causados por un valor por defecto inconsistente
-- en versiones anteriores del formulario de productos. Se elige "UNIDADES"
-- como canónica por ser la mayoritaria; "u" deja de ofrecerse en el
-- desplegable del frontend.
UPDATE productos SET unidad_medida = 'UNIDADES' WHERE unidad_medida = 'U';

-- +goose Down
-- No se revierte: no hay forma de distinguir qué filas decían "U" antes de
-- la unificación.
