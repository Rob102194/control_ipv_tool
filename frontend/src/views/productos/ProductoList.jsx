import React, { useEffect, useState, useRef } from 'react';
import { Table, Button, Container, Alert, Spinner, Form, Modal } from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';
// Importaciones de la API de productos
import { obtenerProductos, eliminarProducto, exportarProductos, importarProductos } from '../../api/productoApi';
import { obtenerHistorial } from '../../api/historialApi';
import { useScrollRestore } from '../../hooks/useScrollRestore';
import { DownloadIcon, UploadIcon, ClockIcon, PlusIcon, SearchIcon, PencilIcon, TrashIcon } from '../../components/icons';

// Componente para mostrar la lista de productos
const ProductoList = () => {
  // Búsqueda y orden viven en la URL (no en useState): al entrar a editar un
  // producto y volver, React Router desmonta y remonta este componente, y un
  // useState perdería el filtro elegido. Con la URL como fuente de verdad,
  // basta con que "Cancelar"/guardar en ProductoForm regrese con
  // navigate(-1) para que se restaure tal cual se dejó.
  const [searchParams, setSearchParams] = useSearchParams();
  const filtro = searchParams.get('q') || '';
  const sortBy = searchParams.get('sort') || 'nombre';

  const actualizarParam = (clave, valor, porDefecto = '') => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev);
      if (valor && valor !== porDefecto) next.set(clave, valor);
      else next.delete(clave);
      return next;
    }, { replace: true });
  };
  const setFiltro = (valor) => actualizarParam('q', valor);
  const setSortBy = (valor) => actualizarParam('sort', valor, 'nombre');

  // Estados para manejar los productos, la carga y los errores
  const [productos, setProductos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const fileInputRef = useRef(null);
  const [showHistory, setShowHistory] = useState(false);
  const [history, setHistory] = useState([]);

  // Restaura la posición de scroll al volver de editar un producto.
  useScrollRestore('productos', !loading);

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
      try {
        await importarProductos(file);
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

  // Muestra un mensaje de error a pantalla completa solo si no hay nada que mostrar.
  if (error && productos.length === 0) {
    return (
      <Container className="mt-5">
        <Alert variant="danger">
          {error}{' '}
          <Button variant="link" className="p-0 align-baseline" onClick={cargarProductos}>Reintentar</Button>
        </Alert>
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
      <div className="page-header">
        <div>
          <h1>Productos</h1>
          <p>Busca, edita y gestiona el catálogo de productos.</p>
        </div>
        <div className="page-actions">
          <Button variant="outline-secondary" onClick={handleExportar}>
            <DownloadIcon size={15} /> Exportar
          </Button>
          <Button variant="outline-secondary" onClick={handleImportClick}>
            <UploadIcon size={15} /> Importar
          </Button>
          <Button variant="outline-secondary" onClick={handleShowHistory}>
            <ClockIcon size={15} /> Historial
          </Button>
          <input
            type="file"
            ref={fileInputRef}
            style={{ display: 'none' }}
            onChange={handleImportar}
            accept=".xlsx, .xls"
          />
          <Link to="/productos/nuevo" className="btn btn-primary d-inline-flex align-items-center gap-2">
            <PlusIcon size={15} /> Nuevo Producto
          </Link>
        </div>
      </div>

      {error && <Alert variant="danger" onClose={() => setError('')} dismissible>{error}</Alert>}

      {/* Campo de búsqueda y ordenamiento */}
      <Form.Group className="mb-3 d-flex">
        <div className="search-wrap me-2" style={{ flex: 1 }}>
          <span className="search-icon"><SearchIcon size={15} /></span>
          <Form.Control
            type="text"
            placeholder="Buscar producto por nombre..."
            value={filtro}
            onChange={(e) => setFiltro(e.target.value)}
          />
        </div>
        <Form.Select value={sortBy} onChange={(e) => setSortBy(e.target.value)} style={{ width: '200px' }}>
          <option value="nombre">Ordenar por Nombre</option>
          <option value="modificado">Ordenar por Modificado</option>
        </Form.Select>
      </Form.Group>

      <div className="list-card">
        <table className="list-table">
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
                <td colSpan="3" style={{ textAlign: 'center' }}>No se encontraron productos.</td>
              </tr>
            ) : (
              productosFiltrados.map((producto) => (
                <tr key={producto.id}>
                  <td data-label="Nombre" style={{ fontWeight: 600 }}>{producto.nombre}</td>
                  <td data-label="Unidad de medida" style={{ color: 'var(--text-secondary)' }}>{producto.unidad_medida}</td>
                  <td data-label="">
                    <Link
                      to={`/productos/editar/${producto.id}`}
                      className="btn btn-outline-secondary btn-sm me-2 d-inline-flex align-items-center gap-1"
                    >
                      <PencilIcon size={13} /> Editar
                    </Link>
                    <Button
                      variant="outline-danger"
                      size="sm"
                      className="d-inline-flex align-items-center gap-1"
                      onClick={() => handleEliminar(producto.id)}
                    >
                      <TrashIcon size={13} /> Eliminar
                    </Button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

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
