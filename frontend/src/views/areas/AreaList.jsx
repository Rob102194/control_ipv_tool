import React, { useState, useEffect } from 'react';
import { Button, Container, Alert, Spinner } from 'react-bootstrap';
import { Link } from 'react-router-dom';
// Importación de la API de áreas
import areaApi from '../../api/areaApi';
import { useConfirm } from '../../contexts/ConfirmContext';
import { PlusIcon, PencilIcon, TrashIcon } from '../../components/icons';

// Componente para mostrar la lista de áreas
const AreaList = () => {
  const confirmar = useConfirm();
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
    if (await confirmar('¿Estás seguro de eliminar esta área?')) {
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

  // Muestra un mensaje de error a pantalla completa solo si no hay nada que mostrar.
  if (error && areas.length === 0) {
    return (
      <Container className="mt-5">
        <Alert variant="danger">
          {error}{' '}
          <Button variant="link" className="p-0 align-baseline" onClick={cargarAreas}>Reintentar</Button>
        </Alert>
      </Container>
    );
  }

  // Renderiza la lista de áreas
  return (
    <Container className="mt-4" style={{ maxWidth: '900px' }}>
      <div className="page-header">
        <div>
          <h1>Áreas</h1>
          <p>Las áreas agrupan qué productos aparecen en la hoja de IPV de cada una.</p>
        </div>
        <Link to="/areas/nuevo" className="btn btn-primary d-inline-flex align-items-center gap-2">
          <PlusIcon size={15} /> Nueva Área
        </Link>
      </div>

      {error && <Alert variant="danger" onClose={() => setError('')} dismissible>{error}</Alert>}

      <div className="list-card">
        <table className="list-table">
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
                <td data-label="Nombre" style={{ fontWeight: 600 }}>{area.nombre}</td>
                <td data-label="Código" style={{ color: 'var(--text-secondary)' }}>{area.codigo || '— sin código —'}</td>
                <td data-label="">
                  <Link
                    to={`/areas/editar/${area.id}`}
                    className="btn btn-outline-secondary btn-sm me-2 d-inline-flex align-items-center gap-1"
                  >
                    <PencilIcon size={13} /> Editar
                  </Link>
                  <Button
                    variant="outline-danger"
                    size="sm"
                    className="d-inline-flex align-items-center gap-1"
                    onClick={() => handleEliminar(area.id)}
                  >
                    <TrashIcon size={13} /> Eliminar
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Container>
  );
};

export default AreaList;
