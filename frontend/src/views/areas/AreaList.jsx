import React, { useState, useEffect } from 'react';
import { Table, Button, Container, Alert, Spinner } from 'react-bootstrap';
import { Link } from 'react-router-dom';
// Importación de la API de áreas
import areaApi from '../../api/areaApi';

// Componente para mostrar la lista de áreas
const AreaList = () => {
  // Estados para manejar las áreas, la carga y los errores
  const [areas, setAreas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Carga las áreas cuando el componente se monta
  useEffect(() => {
    cargarAreas();
  }, []);

  // Función asíncrona para cargar las áreas desde la API
  const cargarAreas = async () => {
    try {
      setLoading(true);
      const response = await areaApi.obtenerTodos();
      setAreas(response.data);
      setError('');
    } catch (err) {
      setError('Error al cargar las áreas');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Maneja la eliminación de un área
  const handleEliminar = async (id) => {
    // Pide confirmación al usuario
    if (window.confirm('¿Estás seguro de eliminar esta área?')) {
      try {
        await areaApi.eliminar(id);
        cargarAreas(); // Recarga la lista después de eliminar
      } catch (err) {
        setError('Error al eliminar el área');
        console.error(err);
      }
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

  // Muestra un mensaje de error si ocurre un problema
  if (error) {
    return (
      <Container className="mt-5">
        <Alert variant="danger">{error}</Alert>
      </Container>
    );
  }

  // Renderiza la lista de áreas
  return (
    <Container className="mt-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Áreas</h1>
        <Link to="/areas/nuevo" className="btn btn-primary">
          Nueva Área
        </Link>
      </div>

      <Table striped bordered hover responsive>
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Código</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {areas.map((area) => (
            <tr key={area.id}>
              <td>{area.nombre}</td>
              <td>{area.codigo || '-'}</td>
              <td>
                {/* Enlaces para editar y eliminar */}
                <Link 
                  to={`/areas/editar/${area.id}`} 
                  className="btn btn-sm btn-warning me-2"
                >
                  Editar
                </Link>
                <Button 
                  variant="danger" 
                  size="sm"
                  onClick={() => handleEliminar(area.id)}
                >
                  Eliminar
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
    </Container>
  );
};

export default AreaList;
