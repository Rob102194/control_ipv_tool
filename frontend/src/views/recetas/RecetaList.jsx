import React, { useState, useEffect, useRef } from 'react';
import { Table, Button, Container, Alert, Spinner, Form, Modal } from 'react-bootstrap';
import { Link, useSearchParams, useLocation } from 'react-router-dom';
import recetaApi from '../../api/recetaApi';
import { obtenerHistorial } from '../../api/historialApi';
import { DownloadIcon, UploadIcon, ClockIcon, PlusIcon, SearchIcon, PencilIcon, TrashIcon } from '../../components/icons';

// Componente para listar, gestionar e importar recetas
const RecetaList = () => {
  // Búsqueda, orden y filtro viven en la URL (no en useState): al entrar a
  // editar una receta y volver, React Router desmonta y remonta este
  // componente, y un useState perdería el filtro elegido. Con la URL como
  // fuente de verdad, basta con que "Cancelar"/guardar en RecetaForm regrese
  // con navigate(-1) para que se restaure tal cual se dejó.
  const [searchParams, setSearchParams] = useSearchParams();
  const filtro = searchParams.get('q') || '';
  const sortBy = searchParams.get('sort') || 'nombre';
  const filterBy = searchParams.get('filter') || '';

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
  const setFilterBy = (valor) => actualizarParam('filter', valor);

  // Estados del componente
  const [recetas, setRecetas] = useState([]); // Almacena la lista de recetas
  const [loading, setLoading] = useState(true); // Indica si se están cargando los datos
  const [error, setError] = useState(''); // Almacena mensajes de error
  const [importing, setImporting] = useState(false); // Indica si hay una importación en curso
  const [importResult, setImportResult] = useState(null); // Almacena el resultado de la importación
  const fileInputRef = useRef(null); // Referencia al input de archivo para importación
  const [showHistory, setShowHistory] = useState(false);
  const [history, setHistory] = useState([]);

  // Restaura la posición de scroll al volver de editar una receta.
  // `location.key` identifica esta entrada del historial (la misma al volver
  // con "atrás"/navigate(-1), nueva en cada navegación fresca).
  //
  // No basta con leer window.scrollY en la limpieza del efecto al
  // desmontar: para cuando React desmonta esta lista y monta RecetaForm, el
  // documento ya se achicó (de 563 filas a un formulario corto) y el propio
  // navegador ya recortó el scroll a 0 — se captura DEMASIADO TARDE. Por eso
  // se guarda en cada scroll, mientras la lista sigue siendo la página
  // completa.
  const location = useLocation();
  const scrollKey = 'recetas-scroll:' + location.key;
  useEffect(() => {
    let frame = null;
    const onScroll = () => {
      if (frame) return;
      frame = requestAnimationFrame(() => {
        sessionStorage.setItem(scrollKey, String(window.scrollY));
        frame = null;
      });
    };
    window.addEventListener('scroll', onScroll);
    return () => {
      window.removeEventListener('scroll', onScroll);
      if (frame) cancelAnimationFrame(frame);
    };
  }, [scrollKey]);
  useEffect(() => {
    if (loading) return;
    const guardado = sessionStorage.getItem(scrollKey);
    if (guardado) {
      // behavior: 'instant' para evitar un scroll animado: el proyecto fija
      // `scroll-behavior: smooth` a nivel global (Bootstrap), y con una
      // lista de cientos de filas una animación "suave" tarda varios
      // segundos en lugar de restaurar la posición al instante.
      requestAnimationFrame(() => window.scrollTo({ top: parseInt(guardado, 10), behavior: 'instant' }));
    }
  }, [loading, scrollKey]);

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

  // Muestra un mensaje de error a pantalla completa solo si no hay nada que mostrar.
  if (error && recetas.length === 0) {
    return (
      <Container className="mt-5">
        <Alert variant="danger">
          {error}{' '}
          <Button variant="link" className="p-0 align-baseline" onClick={cargarRecetas}>Reintentar</Button>
        </Alert>
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
      <div className="page-header">
        <div>
          <h1>Recetas</h1>
          <p>Cada receta convierte una venta en consumo de ingredientes por área.</p>
        </div>
        <div className="page-actions">
          <Button variant="outline-secondary" onClick={handleExportar}>
            <DownloadIcon size={15} /> Exportar
          </Button>
          <Button variant="outline-secondary" onClick={handleImportClick} disabled={importing}>
            {importing ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" />
                <span className="visually-hidden">Importando...</span>
              </>
            ) : (
              <><UploadIcon size={15} /> Importar</>
            )}
          </Button>
          <Button variant="outline-secondary" onClick={handleShowHistory}>
            <ClockIcon size={15} /> Historial
          </Button>
          <input
            type="file"
            ref={fileInputRef}
            style={{ display: 'none' }}
            onChange={handleFileChange}
            accept=".xlsx, .xls"
          />
          <Link to="/recetas/nuevo" className="btn btn-primary d-inline-flex align-items-center gap-2">
            <PlusIcon size={15} /> Nueva Receta
          </Link>
        </div>
      </div>

      {error && <Alert variant="danger" onClose={() => setError('')} dismissible>{error}</Alert>}

      {/* Muestra el resultado de la importación */}
      {importResult && (
        <Alert variant="success" onClose={() => setImportResult(null)} dismissible>
          {importResult.message}
        </Alert>
      )}

      {/* Campo de búsqueda, ordenamiento y filtro */}
      <Form.Group className="mb-3 d-flex align-items-center gap-2">
        <div className="search-wrap" style={{ flex: 1 }}>
          <span className="search-icon"><SearchIcon size={15} /></span>
          <Form.Control
            type="text"
            placeholder="Buscar receta por nombre..."
            value={filtro}
            onChange={(e) => setFiltro(e.target.value)}
          />
        </div>
        <Form.Select value={sortBy} onChange={(e) => setSortBy(e.target.value)} style={{ width: '200px' }}>
          <option value="nombre">Ordenar por Nombre</option>
          <option value="modificado">Ordenar por Modificado</option>
        </Form.Select>
        <button
          type="button"
          className={`filter-chip${filterBy === 'sin_ingredientes' ? ' active' : ''}`}
          onClick={() => setFilterBy(filterBy === 'sin_ingredientes' ? '' : 'sin_ingredientes')}
        >
          <span className="filter-chip-dot"></span> Sin ingredientes
        </button>
      </Form.Group>

      {/* Tabla con la lista de recetas */}
      <div className="list-card">
        <table className="list-table">
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
                <td colSpan="4" style={{ textAlign: 'center' }}>No se encontraron recetas.</td>
              </tr>
            ) : (
              recetasFiltradas.map((receta) => (
                <tr key={receta.id}>
                  <td data-label="Nombre" style={{ fontWeight: 600 }}>{receta.nombre}</td>
                  <td data-label="Estado">
                    {receta.activa ? (
                      <span className="ipv-badge ipv-badge-pos">Activa</span>
                    ) : (
                      <span className="ipv-badge ipv-badge-zero">Inactiva</span>
                    )}
                  </td>
                  <td data-label="Ingredientes">
                    {receta.ingredientes.length === 0 ? (
                      <span className="ipv-badge ipv-badge-warn">Sin ingredientes</span>
                    ) : (
                      <span className="ipv-badge ipv-badge-zero">
                        {receta.ingredientes.length} {receta.ingredientes.length === 1 ? 'ingrediente' : 'ingredientes'}
                      </span>
                    )}
                  </td>
                  <td data-label="">
                    <Link
                      to={`/recetas/editar/${receta.id}`}
                      className="btn btn-outline-secondary btn-sm me-2 d-inline-flex align-items-center gap-1"
                    >
                      <PencilIcon size={13} /> Editar
                    </Link>
                    <Button
                      variant="outline-danger"
                      size="sm"
                      className="d-inline-flex align-items-center gap-1"
                      onClick={() => handleEliminar(receta.id)}
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
