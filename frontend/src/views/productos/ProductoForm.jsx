import React, { useState, useEffect } from 'react';
import { Form, Button, Container, Alert, Spinner } from 'react-bootstrap';
import { useNavigate, useParams, Link } from 'react-router-dom';
// Importaciones de la API de productos
import { obtenerProductoPorId, actualizarProducto, crearProducto } from '../../api/productoApi';
import { CheckIcon } from '../../components/icons';

// Lista de unidades de medida disponibles. "unidades" es la etiqueta
// canónica para ítems contables (bebidas, platos, etc.); "u" quedó
// unificado en ella (eran dos etiquetas para lo mismo, ver migración
// 00002_unifica_unidad_medida_unidades.sql).
const unidadesMedida = [
  'unidades',
  'kg',
  'l',
  'trago',
  'copa'
];

// "u" ya no se ofrece como opción (unificado en "unidades"), pero se
// traduce por si quedara algún valor así en datos no migrados.
const unidadAliases = { u: 'unidades' };

// Normaliza a minúsculas para que coincida con una opción del <select>: el
// backend siempre devuelve la unidad en MAYÚSCULAS, así que sin esto el
// valor cargado nunca calza con ninguna <option> (comparación sensible a
// mayúsculas) y el desplegable queda mostrando la primera opción en vez del
// valor real guardado.
function normalizarUnidad(valor) {
  const v = (valor || '').trim().toLowerCase();
  return unidadAliases[v] || v;
}

// Componente de formulario para crear y editar productos
const ProductoForm = ({ onProductoCreado }) => {
  // Hooks para navegación y parámetros de URL
  const navigate = useNavigate();
  // useParams() toma los parámetros de la ruta que más cerca coincide en el
  // árbol, aunque este componente esté embebido como modal (p. ej. el botón
  // "Crear Producto" de RecetaForm en /recetas/editar/:id) — en ese caso
  // devolvería el id de la RECETA, no de un producto. Cuando se usa como
  // modal de creación rápida (onProductoCreado presente), ignoramos el
  // parámetro de ruta: este formulario siempre crea, nunca edita.
  const params = useParams();
  const id = onProductoCreado ? undefined : params.id;

  // Estado para el producto, carga, guardado y errores
  const [producto, setProducto] = useState({
    nombre: '',
    unidad_medida: unidadesMedida[0]
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  // Carga los datos del producto si se está editando
  useEffect(() => {
    if (id) {
      cargarProducto();
    }
  }, [id]);

  // Función para cargar un producto por su ID
  const cargarProducto = async () => {
    try {
      setLoading(true);
      const response = await obtenerProductoPorId(id);
      setProducto({
        ...response.data,
        unidad_medida: normalizarUnidad(response.data.unidad_medida),
      });
      setError('');
    } catch (err) {
      setError('Error al cargar el producto');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Maneja los cambios en los campos del formulario
  const handleChange = (e) => {
    const { name, value } = e.target;
    setProducto({
      ...producto,
      [name]: value
    });
  };

  // Maneja el envío del formulario
  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    
    try {
      if (id) {
        // Actualiza si hay ID
        await actualizarProducto(id, producto);
      } else {
        // Crea si no hay ID
        await crearProducto(producto);
      }
      
      if (onProductoCreado) {
        onProductoCreado(); // Llama al callback si existe
      } else {
        // navigate(-1) (no to('/productos')) para volver exactamente a la
        // URL de origen, con su búsqueda/orden tal como se dejaron.
        navigate(-1);
      }
    } catch (err) {
      // Muestra el mensaje de error específico del backend si está disponible
      const errorMessage = err.response?.data?.error || 'Error al guardar el producto';
      setError(errorMessage);
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  // Muestra un spinner mientras se cargan los datos
  if (loading) {
    return (
      <Container className="mt-5 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Cargando...</span>
        </Spinner>
      </Container>
    );
  }

  // Renderiza el formulario
  return (
    <Container className="mt-4" style={{ maxWidth: onProductoCreado ? 'none' : '560px' }}>
      {!onProductoCreado && (
        <>
          <div style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-secondary)', marginBottom: '6px' }}>
            <Link to="/productos" style={{ color: 'inherit' }}>Productos</Link> / {id ? 'Editar' : 'Nuevo'}
          </div>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '24px' }}>{id ? 'Editar Producto' : 'Nuevo Producto'}</h1>
        </>
      )}

      {error && <Alert variant="danger" className="mb-4">{error}</Alert>}

      <div className={onProductoCreado ? '' : 'card'}>
        <Form onSubmit={handleSubmit}>
          {/* Campo Nombre */}
          <Form.Group className="mb-3">
            <Form.Label className="fw-semibold">Nombre</Form.Label>
            <Form.Control
              type="text"
              name="nombre"
              value={producto.nombre}
              onChange={handleChange}
              required
              placeholder="Ej: Queso Gouda"
            />
          </Form.Group>

          {/* Campo Unidad de Medida */}
          <Form.Group className="mb-3">
            <Form.Label className="fw-semibold">Unidad de Medida</Form.Label>
            <Form.Select
              name="unidad_medida"
              value={producto.unidad_medida}
              onChange={handleChange}
              required
            >
              {/* Si el valor cargado no es ninguna opción conocida (otro alias
                  legado no mapeado), se muestra igualmente en vez de
                  reemplazarlo en silencio por la primera opción. */}
              {!unidadesMedida.includes(producto.unidad_medida) && producto.unidad_medida && (
                <option value={producto.unidad_medida}>{producto.unidad_medida} (desconocida)</option>
              )}
              {unidadesMedida.map((um) => (
                <option key={um} value={um}>
                  {um}
                </option>
              ))}
            </Form.Select>
          </Form.Group>

          {/* Botones de acción */}
          <div className="d-flex justify-content-end gap-2 pt-2 mt-2 border-top">
            {!onProductoCreado && (
              <Button
                variant="outline-secondary"
                onClick={() => navigate(-1)}
                disabled={saving}
              >
                Cancelar
              </Button>
            )}
            <Button
              variant="primary"
              type="submit"
              disabled={saving}
              className="d-inline-flex align-items-center gap-2"
            >
              {saving ? (
                <>
                  <Spinner animation="border" size="sm" />
                  Guardando...
                </>
              ) : (
                <>
                  <CheckIcon size={15} /> Guardar
                </>
              )}
            </Button>
          </div>
        </Form>
      </div>
    </Container>
  );
};

export default ProductoForm;
