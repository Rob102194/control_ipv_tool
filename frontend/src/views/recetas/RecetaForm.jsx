import React, { useState, useEffect } from 'react';
import { Button, Container, Alert, Spinner, Form, Modal } from 'react-bootstrap';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { Typeahead } from 'react-bootstrap-typeahead';
import 'react-bootstrap-typeahead/css/Typeahead.css';
import { Formik } from 'formik';
import * as Yup from 'yup';
// Importaciones de la API
import recetaApi from '../../api/recetaApi';
import * as productoApi from '../../api/productoApi';
import areaApi from '../../api/areaApi';
import ProductoForm from '../productos/ProductoForm';
import { PlusIcon, BoxIcon, TrashIcon, CheckIcon } from '../../components/icons';

// Esquema de validación con Yup para el formulario de recetas
const recetaSchema = Yup.object().shape({
  nombre: Yup.string().required('El nombre es requerido'),
  activa: Yup.boolean().required(),
  ingredientes: Yup.array().of(
    Yup.object().shape({
      producto_id: Yup.string().required('Seleccione un producto'),
      area_id: Yup.string().required('Seleccione un área'),
      cantidad: Yup.number().min(0.01, 'La cantidad debe ser mayor a 0').required('Ingrese la cantidad')
    })
  )
});

// Componente de formulario para crear y editar recetas
const RecetaForm = () => {
  // Hooks de React Router para navegación y parámetros de URL
  const navigate = useNavigate();
  const { id } = useParams(); // Obtiene el ID de la receta si se está editando

  // Estados del componente
  const [loading, setLoading] = useState(!!id); // Muestra spinner si se está editando (cargando datos)
  const [saving, setSaving] = useState(false); // Muestra spinner en el botón de guardar
  const [error, setError] = useState(''); // Almacena mensajes de error
  const [productos, setProductos] = useState([]); // Lista de productos para los select
  const [areas, setAreas] = useState([]); // Lista de áreas para los select
  const [showProductoModal, setShowProductoModal] = useState(false);

  // Efecto para cargar datos necesarios (productos, áreas y la receta si se edita)
  useEffect(() => {
    const cargarDependencias = async () => {
      try {
        // Carga productos y áreas en paralelo
        const [productosRes, areasRes] = await Promise.all([
          productoApi.obtenerProductos(),
          areaApi.obtenerTodos()
        ]);
        
        setProductos(productosRes.data);
        setAreas(areasRes.data);
        
        // Si hay un ID, carga los datos de la receta a editar
        if (id) {
          const recetaRes = await recetaApi.obtenerPorId(id);
          setInitialValues({
            ...recetaRes.data,
            ingredientes: recetaRes.data.ingredientes || []
          });
        }
      } catch (err) {
        setError('Error al cargar datos');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    
    cargarDependencias();
  }, [id]); // Se ejecuta cuando cambia el ID

  const handleProductoCreado = async () => {
    setShowProductoModal(false);
    try {
      const productosRes = await productoApi.obtenerProductos();
      setProductos(productosRes.data);
    } catch (error) {
      setError('Error al recargar la lista de productos.');
    }
  };

  // Estado para los valores iniciales del formulario
  const [initialValues, setInitialValues] = useState({
    nombre: '',
    activa: true,
    ingredientes: []
  });

  // Maneja el envío del formulario
  const handleSubmit = async (values) => {
    setSaving(true);
    setError('');
    
    try {
      if (id) {
        // Actualiza la receta si existe un ID
        await recetaApi.actualizar(id, values);
      } else {
        // Crea una nueva receta si no hay ID
        await recetaApi.crear(values);
      }
      navigate('/recetas'); // Redirige a la lista de recetas
    } catch (err) {
      // Muestra el mensaje de error específico del backend si está disponible
      const errorMessage = err.response?.data?.error || 'Error al guardar la receta';
      setError(errorMessage);
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  // Muestra un spinner mientras se cargan los datos iniciales
  if (loading) {
    return (
      <Container className="mt-5 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Cargando...</span>
        </Spinner>
      </Container>
    );
  }

  // Renderizado del formulario con Formik
  return (
    <Container className="mt-4" style={{ maxWidth: '840px' }}>
      <div style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-secondary)', marginBottom: '6px' }}>
        <Link to="/recetas" style={{ color: 'inherit' }}>Recetas</Link> / {id ? 'Editar' : 'Nueva'}
      </div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '24px' }}>{id ? `Editar Receta: ${initialValues.nombre}` : 'Nueva Receta'}</h1>

      {error && <Alert variant="danger" className="mb-4">{error}</Alert>}

      <Formik
        initialValues={initialValues}
        validationSchema={recetaSchema}
        onSubmit={handleSubmit}
        enableReinitialize // Permite que el formulario se reinicialice si `initialValues` cambia
      >
        {({ values, errors, touched, handleChange, handleSubmit, setFieldValue }) => (
          <Form onSubmit={handleSubmit}>
            <div className="card mb-4">
              <div className="d-flex gap-4 flex-wrap align-items-end">
                {/* Campo Nombre */}
                <Form.Group style={{ flex: 1, minWidth: '260px' }}>
                  <Form.Label className="fw-semibold">Nombre</Form.Label>
                  <Form.Control
                    type="text"
                    name="nombre"
                    value={values.nombre}
                    onChange={handleChange}
                    isInvalid={touched.nombre && !!errors.nombre}
                  />
                  <Form.Control.Feedback type="invalid">
                    {errors.nombre}
                  </Form.Control.Feedback>
                </Form.Group>

                {/* Campo Activa */}
                <Form.Check
                  type="switch"
                  id="activa"
                  name="activa"
                  label="Activa"
                  checked={values.activa}
                  onChange={handleChange}
                  className="pb-2"
                />
              </div>
            </div>

            <div className="d-flex align-items-center justify-content-between mb-3 flex-wrap gap-2">
              <h2 style={{ fontSize: '1rem', fontWeight: 700, margin: 0 }}>Ingredientes</h2>
              <div className="d-flex gap-2">
                <Button
                  variant="outline-secondary"
                  size="sm"
                  className="d-inline-flex align-items-center gap-2"
                  onClick={() => setFieldValue('ingredientes', [...values.ingredientes, { id: crypto.randomUUID(), producto_id: '', area_id: '', cantidad: 1 }])}
                >
                  <PlusIcon size={14} /> Agregar Ingrediente
                </Button>
                <Button
                  variant="outline-secondary"
                  size="sm"
                  className="d-inline-flex align-items-center gap-2"
                  onClick={() => setShowProductoModal(true)}
                >
                  <BoxIcon size={14} /> Crear Producto
                </Button>
              </div>
            </div>

            {/* Muestra error general de ingredientes (ej. lista vacía) */}
            {touched.ingredientes && typeof errors.ingredientes === 'string' && (
              <div className="text-danger mb-3">{errors.ingredientes}</div>
            )}

            {/* Constructor de ingredientes */}
            <div className="ing-builder mb-4">
              <div className="ing-head">
                <span>Producto</span><span>Área</span><span>Cantidad</span><span></span>
              </div>
              {values.ingredientes.length === 0 && (
                <div className="p-4 text-center" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                  Todavía no hay ingredientes. Usa "Agregar Ingrediente" para empezar.
                </div>
              )}
              {values.ingredientes.map((ing, index) => (
                <div className="ing-row" key={ing.id ?? index}>
                  {/* Columna Producto con autocompletado */}
                  <div>
                    <Typeahead
                      id={`producto-typeahead-${index}`}
                      options={productos}
                      labelKey={option => `${option.nombre} (${option.unidad_medida})`}
                      selected={productos.filter(p => p.id === ing.producto_id)}
                      onChange={(selected) => {
                        const productoId = selected.length > 0 ? selected[0].id : '';
                        setFieldValue(`ingredientes.${index}.producto_id`, productoId);
                      }}
                      placeholder="Escriba para buscar un producto..."
                      isInvalid={touched.ingredientes?.[index]?.producto_id && !!errors.ingredientes?.[index]?.producto_id}
                    />
                    {touched.ingredientes?.[index]?.producto_id && errors.ingredientes?.[index]?.producto_id && (
                      <div className="text-danger" style={{ fontSize: '0.8125rem', marginTop: '0.25rem' }}>
                        {errors.ingredientes[index].producto_id}
                      </div>
                    )}
                  </div>
                  {/* Columna Área */}
                  <div>
                    <Form.Select
                      name={`ingredientes.${index}.area_id`}
                      value={ing.area_id}
                      onChange={handleChange}
                      isInvalid={touched.ingredientes?.[index]?.area_id && !!errors.ingredientes?.[index]?.area_id}
                    >
                      <option value="">Seleccione un área</option>
                      {areas.map(a => (
                        <option key={a.id} value={a.id}>
                          {a.nombre}
                        </option>
                      ))}
                    </Form.Select>
                    {touched.ingredientes?.[index]?.area_id && errors.ingredientes?.[index]?.area_id && (
                      <div className="text-danger" style={{ fontSize: '0.8125rem', marginTop: '0.25rem' }}>{errors.ingredientes[index].area_id}</div>
                    )}
                  </div>
                  {/* Columna Cantidad */}
                  <div>
                    <Form.Control
                      type="number"
                      step="0.01"
                      min="0.01"
                      name={`ingredientes.${index}.cantidad`}
                      value={ing.cantidad}
                      onChange={handleChange}
                      isInvalid={touched.ingredientes?.[index]?.cantidad && !!errors.ingredientes?.[index]?.cantidad}
                      style={{ textAlign: 'right' }}
                    />
                    {touched.ingredientes?.[index]?.cantidad && errors.ingredientes?.[index]?.cantidad && (
                      <div className="text-danger" style={{ fontSize: '0.8125rem', marginTop: '0.25rem' }}>{errors.ingredientes[index].cantidad}</div>
                    )}
                  </div>
                  {/* Columna Acciones */}
                  <Button
                    variant="outline-danger"
                    className="d-inline-flex align-items-center justify-content-center"
                    style={{ width: '34px', height: '34px', padding: 0 }}
                    aria-label={`Quitar ${ing.producto_id ? 'ingrediente' : 'fila'}`}
                    onClick={() => {
                      const newIngs = [...values.ingredientes];
                      newIngs.splice(index, 1);
                      setFieldValue('ingredientes', newIngs);
                    }}
                  >
                    <TrashIcon size={14} />
                  </Button>
                </div>
              ))}
            </div>

            {/* Botones de acción del formulario */}
            <div className="d-flex justify-content-end gap-2">
              <Button
                variant="outline-secondary"
                onClick={() => navigate('/recetas')}
                disabled={saving}
              >
                Cancelar
              </Button>
              <Button
                variant="primary"
                type="submit"
                disabled={saving}
                className="d-inline-flex align-items-center gap-2"
              >
                {saving ? (
                  <>
                    <Spinner animation="border" size="sm" />
                    Guardando...
                  </>
                ) : (
                  <>
                    <CheckIcon size={15} /> Guardar
                  </>
                )}
              </Button>
            </div>
          </Form>
        )}
      </Formik>

      <Modal show={showProductoModal} onHide={() => setShowProductoModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Crear Nuevo Producto</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <ProductoForm onProductoCreado={handleProductoCreado} />
        </Modal.Body>
      </Modal>
    </Container>
  );
};

export default RecetaForm;
