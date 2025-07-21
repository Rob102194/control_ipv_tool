import React, { useEffect, useState, useRef } from 'react';
import { Table, Button, Container, Alert, Spinner, Form, Modal } from 'react-bootstrap';
import { Link } from 'react-router-dom';
// Importaciones de la API de productos
import { obtenerProductos, eliminarProducto, exportarProductos, importarProductos } from '../../api/productoApi';
import { obtenerHistorial } from '../../api/historialApi';

// Componente para mostrar la lista de productos
const ProductoList = () => {
  // Estados para manejar los productos, la carga y los errores
  const [productos, setProductos] = useState([]);
  const [filtro, setFiltro] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [sortBy, setSortBy] = useState('nombre');
  const fileInputRef = useRef(null);
  const [showHistory, setShowHistory] = useState(false);
  const [history, setHistory] = useState([]);

  // Carga los productos cuando el componente se monta
  useEffect(() => {
    cargarProductos();
  }, [sortBy]);

  // Función asíncrona para cargar los productos desde la API
  const cargarProductos = async () => {
    try {
      setLoading(true);
      const response = await obtenerProductos({ sort_by: sortBy });
      setProductos(response.data);
      setError('');
    } catch (err) {
      setError('Error al cargar los productos');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Maneja la eliminación de un producto
  const handleEliminar = async (id) => {
    // Pide confirmación al usuario
    if (window.confirm('¿Estás seguro de eliminar este producto?')) {
      try {
        await eliminarProducto(id);
        cargarProductos(); // Recarga la lista después de eliminar
      } catch (err) {
        setError(err.response?.data?.error || 'Error al eliminar el producto');
        console.error(err);
      }
    }
  };

  // Maneja la exportación de productos a Excel
  const handleExportar = async () => {
    try {
      const response = await exportarProductos();
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'productos.xlsx');
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (err) {
      setError('Error al exportar los productos');
      console.error(err);
    }
  };

  // Maneja la importación de productos desde Excel
  const handleImportar = async (event) => {
    const file = event.target.files[0];
    if (file) {
      const formData = new FormData();
      formData.append('file', file);
      try {
        await importarProductos(formData);
        cargarProductos(); // Recarga la lista después de importar
      } catch (err) {
        setError('Error al importar los productos');
        console.error(err);
      }
    }
  };

  const handleImportClick = () => {
    fileInputRef.current.click();
  };

  const handleShowHistory = async () => {
    try {
      const response = await obtenerHistorial('Producto');
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

  // Filtra los productos basándose en el término de búsqueda
  const productosFiltrados = productos.filter(producto =>
    producto.nombre.toLowerCase().includes(filtro.toLowerCase())
  );

  // Renderiza la lista de productos
  return (
    <Container className="mt-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Productos</h1>
        <div>
          <Button variant="success" onClick={handleExportar} className="me-2">
            Exportar
          </Button>
          <Button variant="info" onClick={handleImportClick} className="me-2">
            Importar
          </Button>
          <Button variant="secondary" onClick={handleShowHistory} className="me-2">
            Historial
          </Button>
          <input
            type="file"
            ref={fileInputRef}
            style={{ display: 'none' }}
            onChange={handleImportar}
            accept=".xlsx, .xls"
          />
          <Link to="/productos/nuevo" className="btn btn-primary ms-2">
            Nuevo Producto
          </Link>
        </div>
      </div>

      {/* Campo de búsqueda y ordenamiento */}
      <Form.Group className="mb-3 d-flex">
        <Form.Control
          type="text"
          placeholder="Buscar producto por nombre..."
          value={filtro}
          onChange={(e) => setFiltro(e.target.value)}
          className="me-2"
        />
        <Form.Select value={sortBy} onChange={(e) => setSortBy(e.target.value)} style={{ width: '200px' }}>
          <option value="nombre">Ordenar por Nombre</option>
          <option value="modificado">Ordenar por Modificado</option>
        </Form.Select>
      </Form.Group>

      <Table striped bordered hover responsive>
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Unidad de Medida</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {productosFiltrados.length === 0 ? (
            <tr>
              <td colSpan="3" className="text-center">No se encontraron productos.</td>
            </tr>
          ) : (
            productosFiltrados.map((producto) => (
              <tr key={producto.id}>
                <td>{producto.nombre}</td>
                <td>{producto.unidad_medida}</td>
                <td>
                  {/* Enlaces para editar y eliminar */}
                  <Link 
                    to={`/productos/editar/${producto.id}`} 
                    className="btn btn-sm btn-warning me-2"
                  >
                    Editar
                  </Link>
                  <Button 
                    variant="danger" 
                    size="sm"
                    onClick={() => handleEliminar(producto.id)}
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
          <Modal.Title>Historial de Cambios de Productos</Modal.Title>
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

export default ProductoList;
