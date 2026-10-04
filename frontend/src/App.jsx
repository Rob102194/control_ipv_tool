import React, { Suspense, lazy, useContext, useEffect, useState } from 'react';
import { createBrowserRouter, RouterProvider, Link, Outlet, useLocation } from 'react-router-dom';
import { Navbar, Nav, Container, Spinner, Modal, Form, Button } from 'react-bootstrap';
import { ThemeContext } from './contexts/ThemeContext';
import { useToast } from './contexts/ToastContext';
import configuracionApi from './api/configuracionApi';
import { SunIcon, MoonIcon, PencilIcon, PlusIcon, CheckIcon } from './components/icons';

// Carga diferida de componentes de las vistas
const ProductoList = lazy(() => import('./views/productos/ProductoList'));
const ProductoForm = lazy(() => import('./views/productos/ProductoForm'));
const AreaList = lazy(() => import('./views/areas/AreaList'));
const AreaForm = lazy(() => import('./views/areas/AreaForm'));
const RecetaList = lazy(() => import('./views/recetas/RecetaList'));
const RecetaForm = lazy(() => import('./views/recetas/RecetaForm'));
const VentaList = lazy(() => import('./views/ventas/VentaList'));
const IPVControl = lazy(() => import('./views/ipv/IPVControl'));
const RegistroIPV = lazy(() => import('./views/ipv/RegistroIPV'));

// Componente para mostrar mientras se cargan los componentes diferidos
const Loading = () => (
  <div className="d-flex justify-content-center mt-5">
    <Spinner animation="border" role="status">
      <span className="visually-hidden">Cargando...</span>
    </Spinner>
  </div>
);

// Nombre del negocio, editable desde la propia app: distingue de un vistazo
// qué negocio estás operando cuando hay varias instancias abiertas a la vez
// (una por negocio, cada una con su propia base de datos — ver
// cmd/desktop/main.go). También actualiza el título de la pestaña en el
// despliegue web, donde no hay ventana nativa que lo muestre.
const NegocioBadge = () => {
  const showToast = useToast();
  const [nombre, setNombre] = useState(null); // null = aún no cargó
  const [showEdit, setShowEdit] = useState(false);
  const [valorEdit, setValorEdit] = useState('');
  const [guardando, setGuardando] = useState(false);

  useEffect(() => {
    configuracionApi.obtener()
      .then(({ data }) => setNombre(data.nombre_negocio || ''))
      .catch(() => setNombre(''));
  }, []);

  useEffect(() => {
    document.title = nombre ? `${nombre} · Control IPV` : 'Control IPV';
  }, [nombre]);

  const abrirEdit = () => {
    setValorEdit(nombre || '');
    setShowEdit(true);
  };

  const guardar = async () => {
    setGuardando(true);
    try {
      const { data } = await configuracionApi.actualizar(valorEdit);
      setNombre(data.nombre_negocio);
      setShowEdit(false);
      showToast('Nombre del negocio actualizado.');
    } catch (err) {
      showToast(err.response?.data?.error || 'Error al guardar el nombre del negocio.', 'danger');
    } finally {
      setGuardando(false);
    }
  };

  if (nombre === null) return null; // evita un parpadeo mientras carga

  return (
    <>
      <button
        type="button"
        onClick={abrirEdit}
        className="d-inline-flex align-items-center gap-2"
        style={{
          background: nombre ? 'var(--color-primary-soft)' : 'transparent',
          color: nombre ? 'var(--color-primary)' : 'var(--text-secondary)',
          border: nombre ? '1px solid transparent' : '1px dashed var(--border-color)',
          borderRadius: '999px',
          padding: '4px 12px',
          fontSize: '0.8125rem',
          fontWeight: 600,
          marginLeft: '12px',
        }}
      >
        {nombre ? (
          <>{nombre} <PencilIcon size={12} /></>
        ) : (
          <><PlusIcon size={12} /> Nombre del negocio</>
        )}
      </button>

      <Modal show={showEdit} onHide={() => setShowEdit(false)} centered>
        <Modal.Header closeButton>
          <Modal.Title>Nombre del negocio</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form.Group>
            <Form.Label className="fw-semibold">Nombre</Form.Label>
            <Form.Control
              type="text"
              value={valorEdit}
              onChange={(e) => setValorEdit(e.target.value)}
              placeholder="Ej: Pizzería Don Mario"
              maxLength={80}
              autoFocus
            />
            <Form.Text style={{ color: 'var(--text-secondary)' }}>
              Útil si operas más de un negocio con instalaciones separadas de la app: ayuda a distinguir
              cada ventana. El título de la ventana se actualiza al reabrir la app.
            </Form.Text>
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="outline-secondary" onClick={() => setShowEdit(false)} disabled={guardando}>
            Cancelar
          </Button>
          <Button variant="primary" onClick={guardar} disabled={guardando} className="d-inline-flex align-items-center gap-2">
            {guardando ? <Spinner animation="border" size="sm" /> : <CheckIcon size={15} />}
            Guardar
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
};

const AppLayout = () => {
  const { theme, toggleTheme } = useContext(ThemeContext);
  const location = useLocation();
  const isActive = (base) => {
    if (base === '/') return location.pathname === '/' || location.pathname.startsWith('/ipv/');
    return location.pathname.startsWith(base);
  };

  return (
    <>
      <Navbar bg={theme} variant={theme} expand="lg" className="border-bottom">
        <Container>
          <Navbar.Brand as={Link} to="/">Gestión de Inventario</Navbar.Brand>
          <NegocioBadge />
          <Navbar.Toggle aria-controls="basic-navbar-nav" className="ms-auto" />
          <Navbar.Collapse id="basic-navbar-nav">
            <Nav className="me-auto">
              <Nav.Link as={Link} to="/" active={isActive('/')}>Control IPV</Nav.Link>
              <Nav.Link as={Link} to="/productos" active={isActive('/productos')}>Productos</Nav.Link>
              <Nav.Link as={Link} to="/areas" active={isActive('/areas')}>Áreas</Nav.Link>
              <Nav.Link as={Link} to="/recetas" active={isActive('/recetas')}>Recetas</Nav.Link>
              <Nav.Link as={Link} to="/ventas" active={isActive('/ventas')}>Ventas</Nav.Link>
            </Nav>
            <button
              type="button"
              className="ipv-icon-btn"
              onClick={toggleTheme}
              aria-label={theme === 'light' ? 'Cambiar a modo oscuro' : 'Cambiar a modo claro'}
              title={theme === 'light' ? 'Modo oscuro' : 'Modo claro'}
            >
              {theme === 'light' ? <MoonIcon size={17} /> : <SunIcon size={17} />}
            </button>
          </Navbar.Collapse>
        </Container>
      </Navbar>
      <Container fluid className="mt-4 px-0">
        <Suspense fallback={<Loading />}>
          <Outlet />
        </Suspense>
      </Container>
    </>
  );
};

const router = createBrowserRouter([
  {
    element: <AppLayout />,
    children: [
      { path: "/", element: <IPVControl /> },
      { path: "/ipv/registro/:fecha", element: <RegistroIPV /> },
      { path: "/productos", element: <ProductoList /> },
      { path: "/productos/nuevo", element: <ProductoForm /> },
      { path: "/productos/editar/:id", element: <ProductoForm /> },
      { path: "/areas", element: <AreaList /> },
      { path: "/areas/nuevo", element: <AreaForm /> },
      { path: "/areas/editar/:id", element: <AreaForm /> },
      { path: "/recetas", element: <RecetaList /> },
      { path: "/recetas/nuevo", element: <RecetaForm /> },
      { path: "/recetas/editar/:id", element: <RecetaForm /> },
      { path: "/ventas", element: <VentaList /> },
    ],
  },
], {
  future: {
    v7_startTransition: true,
  },
});

function App() {
  return <RouterProvider router={router} />;
}

export default App;
