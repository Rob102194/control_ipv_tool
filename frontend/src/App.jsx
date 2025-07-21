import React, { Suspense, lazy, useContext } from 'react';
import { createBrowserRouter, RouterProvider, Link, Outlet } from 'react-router-dom';
import { Navbar, Nav, Container, Spinner, Button } from 'react-bootstrap';
import { ThemeContext } from './contexts/ThemeContext';

// Carga diferida de componentes de las vistas
const ProductoList = lazy(() => import('./views/productos/ProductoList'));
const ProductoForm = lazy(() => import('./views/productos/ProductoForm'));
const AreaList = lazy(() => import('./views/areas/AreaList'));
const AreaForm = lazy(() => import('./views/areas/AreaForm'));
const RecetaList = lazy(() => import('./views/recetas/RecetaList'));
const RecetaForm = lazy(() => import('./views/recetas/RecetaForm'));
const VentaList = lazy(() => import('./views/ventas/VentaList'));
const IPVControl = lazy(() => import('./views/ipv/IPVControl'));

// Componente para mostrar mientras se cargan los componentes diferidos
const Loading = () => (
  <div className="d-flex justify-content-center mt-5">
    <Spinner animation="border" role="status">
      <span className="visually-hidden">Cargando...</span>
    </Spinner>
  </div>
);

const AppLayout = () => {
  const { theme, toggleTheme } = useContext(ThemeContext);

  return (
    <>
      <Navbar bg={theme} variant={theme} expand="lg" className="border-bottom">
        <Container>
          <Navbar.Brand as={Link} to="/">Gestión de Inventario</Navbar.Brand>
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse id="basic-navbar-nav">
            <Nav className="me-auto">
              <Nav.Link as={Link} to="/">Control IPV</Nav.Link>
              <Nav.Link as={Link} to="/productos">Productos</Nav.Link>
              <Nav.Link as={Link} to="/areas">Áreas</Nav.Link>
              <Nav.Link as={Link} to="/recetas">Recetas</Nav.Link>
              <Nav.Link as={Link} to="/ventas">Ventas</Nav.Link>
            </Nav>
            <Button variant="outline-primary" onClick={toggleTheme}>
              {theme === 'light' ? 'Modo Oscuro' : 'Modo Claro'}
            </Button>
          </Navbar.Collapse>
        </Container>
      </Navbar>
      <Container className="mt-4">
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
