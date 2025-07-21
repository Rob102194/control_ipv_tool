import React, { useState, useEffect } from 'react';
import { Form, Button, Container, Alert, Spinner } from 'react-bootstrap';
import { useNavigate, useParams } from 'react-router-dom';
// Importaciones de la API de productos
import { obtenerProductoPorId, actualizarProducto, crearProducto } from '../../api/productoApi';

// Lista de unidades de medida disponibles
const unidadesMedida = [
  'unidades',
  'kg',
  'g',
  'l',
  'ml',
  'paquetes'
];

// Componente de formulario para crear y editar productos
const ProductoForm = () => {
  // Hooks para navegación y parámetros de URL
  const navigate = useNavigate();
  const { id } = useParams();

  // Estado para el producto, carga, guardado y errores
  const [producto, setProducto] = useState({
    nombre: '',
    unidad_medida: 'unidades'
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
      setProducto(response.data);
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
      navigate('/productos'); // Redirige a la lista
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
    <Container className="mt-4">
      <h2 className="mb-4">{id ? 'Editar Producto' : 'Nuevo Producto'}</h2>
      
      {error && <Alert variant="danger" className="mb-4">{error}</Alert>}
      
      <Form onSubmit={handleSubmit}>
        {/* Campo Nombre */}
        <Form.Group className="mb-3">
          <Form.Label>Nombre</Form.Label>
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
          <Form.Label>Unidad de Medida</Form.Label>
          <Form.Select
            name="unidad_medida"
            value={producto.unidad_medida}
            onChange={handleChange}
            required
          >
            {unidadesMedida.map((um) => (
              <option key={um} value={um}>
                {um}
              </option>
            ))}
          </Form.Select>
        </Form.Group>
        
        {/* Botones de acción */}
        <div className="d-flex justify-content-end gap-2">
          <Button 
            variant="secondary" 
            onClick={() => navigate('/productos')}
            disabled={saving}
          >
            Cancelar
          </Button>
          <Button 
            variant="primary" 
            type="submit"
            disabled={saving}
          >
            {saving ? (
              <>
                <Spinner animation="border" size="sm" className="me-2" />
                Guardando...
              </>
            ) : (
              'Guardar'
            )}
          </Button>
        </div>
      </Form>
    </Container>
  );
};

export default ProductoForm;
