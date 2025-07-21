import React, { useState, useEffect, useRef } from 'react';
import { Table, Button, Container, Alert, Spinner, Badge, Form, Modal } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import recetaApi from '../../api/recetaApi';
import { obtenerHistorial } from '../../api/historialApi';

// Componente para listar, gestionar e importar recetas
const RecetaList = () => {
  // Estados del componente
  const [recetas, setRecetas] = useState([]); // Almacena la lista de recetas
  const [filtro, setFiltro] = useState(''); // Almacena el término de búsqueda
  const [loading, setLoading] = useState(true); // Indica si se están cargando los datos
  const [error, setError] = useState(''); // Almacena mensajes de error
  const [importing, setImporting] = useState(false); // Indica si hay una importación en curso
  const [importResult, setImportResult] = useState(null); // Almacena el resultado de la importación
  const fileInputRef = useRef(null); // Referencia al input de archivo para importación
  const [showHistory, setShowHistory] = useState(false);
  const [history, setHistory] = useState([]);
  const [sortBy, setSortBy] = useState('nombre');
  const [filterBy, setFilterBy] = useState('');

  // Carga las recetas cuando el componente se monta
  useEffect(() => {
    cargarRecetas();
  }, [sortBy, filterBy]);

  // Función para obtener las recetas desde la API
  const cargarRecetas = async () => {
    try {
      setLoading(true);
      const params = { sort_by: sortBy };
      if (filterBy) {
        params.filter_by = filterBy;
      }
      const response = await recetaApi.obtenerTodos(params);
      setRecetas(response.data);
      setError('');
    } catch (err) {
      setError('Error al cargar las recetas');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Maneja la eliminación de una receta
  const handleEliminar = async (id) => {
    if (window.confirm('¿Estás seguro de eliminar esta receta?')) {
      try {
        await recetaApi.eliminar(id);
        cargarRecetas(); // Recarga la lista después de eliminar
      } catch (err) {
        setError(err.response?.data?.error || 'Error al eliminar la receta');
        console.error(err);
      }
    }
  };

  const handleShowHistory = async () => {
    try {
      const response = await obtenerHistorial('Receta');
      setHistory(response.data);
      setShowHistory(true);
    } catch (err) {
      setError('Error al cargar el historial');
      console.error(err);
    }
  };

  const handleCloseHistory = () => {
    setShowHistory(false);
    setHistory([]);
  };

  // Muestra un spinner de carga mientras se obtienen los datos
  if (loading) {
    return (
      <Container className="mt-5 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Cargando...</span>
        </Spinner>
      </Container>
    );
  }

  // Muestra un mensaje de error si falla la carga
  if (error) {
    return (
      <Container className="mt-5">
        <Alert variant="danger">{error}</Alert>
      </Container>
    );
  }

  // Simula un clic en el input de archivo oculto
  const handleImportClick = () => {
    fileInputRef.current.click();
  };

  // Descarga la plantilla de Excel para importar recetas
  const handleExportar = async () => {
    try {
      const response = await recetaApi.exportar();
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'recetas.xlsx');
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      setError('Error al exportar las recetas');
      console.error(error);
    }
  };

  // Filtra las recetas basándose en el término de búsqueda
  const recetasFiltradas = recetas.filter(receta =>
    receta.nombre.toLowerCase().includes(filtro.toLowerCase())
  );

  // Maneja el cambio en el input de archivo y procesa la importación
  const handleFileChange = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    setImporting(true);
    setImportResult(null);
    setError('');

    try {
      const response = await recetaApi.importar(file);
      setImportResult(response.data);
      cargarRecetas(); // Recarga la lista para mostrar las nuevas recetas
    } catch (err) {
      setError(err.response?.data?.error || 'Error al importar el archivo');
      console.error(err);
    } finally {
      setImporting(false);
      event.target.value = null; // Resetea el input de archivo
    }
  };

  // Renderizado del componente
  return (
    <Container className="mt-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Recetas</h1>
        <div>
          <Button variant="success" onClick={handleExportar} className="me-2">
            Exportar
          </Button>
          <Button 
            variant="info" 
            className="me-2"
            onClick={handleImportClick}
            disabled={importing}
          >
            {importing ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" />
                <span className="visually-hidden">Importando...</span>
              </>
            ) : (
              'Importar'
            )}
          </Button>
          <Button variant="secondary" onClick={handleShowHistory} className="me-2">
            Historial
          </Button>
          <input
            type="file"
            ref={fileInputRef}
            style={{ display: 'none' }}
            onChange={handleFileChange}
            accept=".xlsx, .xls"
          />
          <Link to="/recetas/nuevo" className="btn btn-primary ms-2">
            Nueva Receta
          </Link>
        </div>
      </div>

      {/* Muestra el resultado de la importación */}
      {importResult && (
        <Alert variant="info" onClose={() => setImportResult(null)} dismissible>
          Importación completada: {importResult.importadas} recetas importadas, {importResult.omitidas} omitidas.
        </Alert>
      )}

      {/* Campo de búsqueda, ordenamiento y filtro */}
      <Form.Group className="mb-3 d-flex align-items-center">
        <Form.Control
          type="text"
          placeholder="Buscar receta por nombre..."
          value={filtro}
          onChange={(e) => setFiltro(e.target.value)}
          className="me-2"
        />
        <Form.Select value={sortBy} onChange={(e) => setSortBy(e.target.value)} style={{ width: '200px' }} className="me-2">
          <option value="nombre">Ordenar por Nombre</option>
          <option value="modificado">Ordenar por Modificado</option>
        </Form.Select>
        <Form.Check
          type="checkbox"
          label="Sin ingredientes"
          checked={filterBy === 'sin_ingredientes'}
          onChange={(e) => setFilterBy(e.target.checked ? 'sin_ingredientes' : '')}
        />
      </Form.Group>

      {/* Tabla con la lista de recetas */}
      <Table striped bordered hover responsive>
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Estado</th>
            <th>Ingredientes</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {recetasFiltradas.length === 0 ? (
            <tr>
              <td colSpan="4" className="text-center">No se encontraron recetas.</td>
            </tr>
          ) : (
            recetasFiltradas.map((receta) => (
              <tr key={receta.id}>
                <td>{receta.nombre}</td>
                <td>
                  {receta.activa ? (
                    <Badge bg="success">Activa</Badge>
                  ) : (
                    <Badge bg="secondary">Inactiva</Badge>
                  )}
                </td>
                <td>{receta.ingredientes.length} ingredientes</td>
                <td>
                  {/* Botones de acción para editar y eliminar */}
                  <Link 
                    to={`/recetas/editar/${receta.id}`} 
                    className="btn btn-sm btn-warning me-2"
                  >
                    Editar
                  </Link>
                  <Button 
                    variant="danger" 
                    size="sm"
                    onClick={() => handleEliminar(receta.id)}
                  >
                    Eliminar
                  </Button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </Table>

      <Modal show={showHistory} onHide={handleCloseHistory}>
        <Modal.Header closeButton>
          <Modal.Title>Historial de Cambios de Recetas</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Table striped bordered hover responsive>
            <thead>
              <tr>
                <th>Fecha</th>
                <th>Campo Modificado</th>
                <th>Valor Anterior</th>
                <th>Valor Nuevo</th>
              </tr>
            </thead>
            <tbody>
              {history.length === 0 ? (
                <tr>
                  <td colSpan="4" className="text-center">No hay historial de cambios.</td>
                </tr>
              ) : (
                history.map((h) => (
                  <tr key={h.id}>
                    <td>{new Date(h.fecha_cambio).toLocaleString()}</td>
                    <td>{h.campo_modificado}</td>
                    <td>{h.valor_anterior}</td>
                    <td>{h.valor_nuevo}</td>
                  </tr>
                ))
              )}
            </tbody>
          </Table>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={handleCloseHistory}>
            Cerrar
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
};

export default RecetaList;
