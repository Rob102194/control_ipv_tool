import React, { useState, useEffect } from 'react';
import { Form, Button, Container, Alert, Spinner } from 'react-bootstrap';
import { useNavigate, useParams } from 'react-router-dom';
// Importación de la API de áreas
import areaApi from '../../api/areaApi';

// Componente de formulario para crear y editar áreas
const AreaForm = () => {
  // Hooks para navegación y parámetros de URL
  const navigate = useNavigate();
  const { id } = useParams();

  // Estado para el área, carga, guardado y errores
  const [area, setArea] = useState({
    nombre: '',
    codigo: ''
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  // Carga los datos del área si se está editando
  useEffect(() => {
    if (id) {
      cargarArea();
    }
  }, [id]);

  // Función para cargar un área por su ID
  const cargarArea = async () => {
    try {
      setLoading(true);
      const response = await areaApi.obtenerPorId(id);
      setArea(response.data);
      setError('');
    } catch (err) {
      setError('Error al cargar el área');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Maneja los cambios en los campos del formulario
  const handleChange = (e) => {
    const { name, value } = e.target;
    setArea({
      ...area,
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
        await areaApi.actualizar(id, area);
      } else {
        // Crea si no hay ID
        await areaApi.crear(area);
      }
      navigate('/areas'); // Redirige a la lista
    } catch (err) {
      setError('Error al guardar el área');
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
      <h2 className="mb-4">{id ? 'Editar Área' : 'Nueva Área'}</h2>
      
      {error && <Alert variant="danger" className="mb-4">{error}</Alert>}
      
      <Form onSubmit={handleSubmit}>
        {/* Campo Nombre */}
        <Form.Group className="mb-3">
          <Form.Label>Nombre</Form.Label>
          <Form.Control
            type="text"
            name="nombre"
            value={area.nombre}
            onChange={handleChange}
            required
            placeholder="Ej: Pizzeria"
          />
        </Form.Group>
        
        {/* Campo Código */}
        <Form.Group className="mb-3">
          <Form.Label>Código (Opcional)</Form.Label>
          <Form.Control
            type="text"
            name="codigo"
            value={area.codigo}
            onChange={handleChange}
            placeholder="Ej: PZ"
          />
        </Form.Group>
        
        {/* Botones de acción */}
        <div className="d-flex justify-content-end gap-2">
          <Button 
            variant="secondary" 
            onClick={() => navigate('/areas')}
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

export default AreaForm;
