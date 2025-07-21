import React, { useState, useEffect } from 'react';
import { Button, Container, Alert, Spinner, Form, Row, Col, Table } from 'react-bootstrap';
import { useNavigate, useParams } from 'react-router-dom';
import { Typeahead } from 'react-bootstrap-typeahead';
import 'react-bootstrap-typeahead/css/Typeahead.css';
import { Formik } from 'formik';
import * as Yup from 'yup';
// Importaciones de la API
import recetaApi from '../../api/recetaApi';
import * as productoApi from '../../api/productoApi';
import areaApi from '../../api/areaApi';

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
    <Container className="mt-4">
      <h2 className="mb-4">{id ? 'Editar Receta' : 'Nueva Receta'}</h2>
      
      {error && <Alert variant="danger" className="mb-4">{error}</Alert>}
      
      <Formik
        initialValues={initialValues}
        validationSchema={recetaSchema}
        onSubmit={handleSubmit}
        enableReinitialize // Permite que el formulario se reinicialice si `initialValues` cambia
      >
        {({ values, errors, touched, handleChange, handleSubmit, setFieldValue }) => (
          <Form onSubmit={handleSubmit}>
            {/* Campo Nombre */}
            <Form.Group className="mb-3">
              <Form.Label>Nombre</Form.Label>
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
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                id="activa"
                name="activa"
                label="Activa"
                checked={values.activa}
                onChange={handleChange}
              />
            </Form.Group>
            
            <h4 className="mb-3">Ingredientes</h4>
            
            {/* Botón para agregar un nuevo ingrediente */}
            <Button
              variant="outline-primary"
              className="mb-3"
              onClick={() => setFieldValue('ingredientes', [...values.ingredientes, { producto_id: '', area_id: '', cantidad: 1 }])}
            >
              Agregar Ingrediente
            </Button>
            
            {/* Muestra error general de ingredientes (ej. lista vacía) */}
            {touched.ingredientes && typeof errors.ingredientes === 'string' && (
              <div className="text-danger mb-3">{errors.ingredientes}</div>
            )}
            
            {/* Tabla para gestionar los ingredientes */}
            <Table striped bordered hover>
              <thead>
                <tr>
                  <th>Producto</th>
                  <th>Área</th>
                  <th>Cantidad</th>
                  <th>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {values.ingredientes.map((ing, index) => (
                  <tr key={index}>
                    {/* Columna Producto con autocompletado */}
                    <td>
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
                        <div className="text-danger" style={{ fontSize: '0.875em', marginTop: '0.25rem' }}>
                          {errors.ingredientes[index].producto_id}
                        </div>
                      )}
                    </td>
                    {/* Columna Área */}
                    <td>
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
                        <div className="text-danger">{errors.ingredientes[index].area_id}</div>
                      )}
                    </td>
                    {/* Columna Cantidad */}
                    <td>
                      <Form.Control
                        type="number"
                        step="0.01"
                        min="0.01"
                        name={`ingredientes.${index}.cantidad`}
                        value={ing.cantidad}
                        onChange={handleChange}
                        isInvalid={touched.ingredientes?.[index]?.cantidad && !!errors.ingredientes?.[index]?.cantidad}
                      />
                      {touched.ingredientes?.[index]?.cantidad && errors.ingredientes?.[index]?.cantidad && (
                        <div className="text-danger">{errors.ingredientes[index].cantidad}</div>
                      )}
                    </td>
                    {/* Columna Acciones */}
                    <td>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => {
                          const newIngs = [...values.ingredientes];
                          newIngs.splice(index, 1);
                          setFieldValue('ingredientes', newIngs);
                        }}
                      >
                        Eliminar
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
            
            {/* Botones de acción del formulario */}
            <div className="d-flex justify-content-end gap-2">
              <Button 
                variant="secondary" 
                onClick={() => navigate('/recetas')}
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
        )}
      </Formik>
    </Container>
  );
};

export default RecetaForm;
