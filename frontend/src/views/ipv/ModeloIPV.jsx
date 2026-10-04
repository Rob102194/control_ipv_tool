import React, { useState, useEffect } from 'react';
import { Button, Row, Col, Form, Alert, Spinner } from 'react-bootstrap';
import { Typeahead } from 'react-bootstrap-typeahead';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import 'react-bootstrap-typeahead/css/Typeahead.css';
import ipvApi from '../../api/ipvApi';
import * as productoApi from '../../api/productoApi';
import areaApi from '../../api/areaApi';
import { useToast } from '../../contexts/ToastContext';
import { CheckIcon, TrashIcon, iconoDeArea } from '../../components/icons';

const ModeloIPV = () => {
    const showToast = useToast();
    const [areas, setAreas] = useState([]);
    const [productos, setProductos] = useState([]);
    const [modelos, setModelos] = useState({});
    const [selectedArea, setSelectedArea] = useState(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [loadError, setLoadError] = useState('');
    const [saveError, setSaveError] = useState('');

    const cargarDatos = async () => {
        try {
            setLoading(true);
            setLoadError('');
            const [areasRes, productosRes, modelosRes] = await Promise.all([
                areaApi.obtenerTodos(),
                productoApi.obtenerProductos(),
                ipvApi.getModelos()
            ]);
            setAreas(areasRes.data);
            setProductos(productosRes.data);

            const modelosConOrden = {};
            for (const areaId in modelosRes.data) {
                modelosConOrden[areaId] = modelosRes.data[areaId].map(p => ({
                    id: p.producto_id,
                    orden: p.orden
                }));
            }
            setModelos(modelosConOrden);
        } catch (err) {
            console.error(err);
            setLoadError('Error al cargar los datos iniciales.');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        cargarDatos();
    }, []);

    const handleSelectArea = (area) => {
        setSelectedArea(area);
    };

    const handleAddProducto = (selected) => {
        if (selected.length > 0) {
            const producto = selected[0];
            const areaId = selectedArea.id;
            const currentModel = modelos[areaId] || [];
            if (!currentModel.find(p => p.id === producto.id)) {
                const newModel = [...currentModel, { id: producto.id, orden: currentModel.length }];
                setModelos(prev => ({ ...prev, [areaId]: newModel }));
            }
        }
    };

    const handleRemoveProducto = (productoId) => {
        const areaId = selectedArea.id;
        const currentModel = modelos[areaId] || [];
        let newModel = currentModel.filter(p => p.id !== productoId);
        // Re-assign the 'orden' property based on the new order
        newModel = newModel.map((item, index) => ({ ...item, orden: index }));
        setModelos(prev => ({ ...prev, [areaId]: newModel }));
    };

    const handleOnDragEnd = (result) => {
        if (!result.destination) return;
        const areaId = selectedArea.id;
        const items = Array.from(modelos[areaId]);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        const updatedItems = items.map((item, index) => ({ ...item, orden: index }));
        setModelos(prev => ({ ...prev, [areaId]: updatedItems }));
    };

    const handleSaveChanges = async () => {
        if (!selectedArea) return;
        setSaving(true);
        setSaveError('');
        try {
            await ipvApi.guardarModelo({
                area_id: selectedArea.id,
                productos: modelos[selectedArea.id] || []
            });
            showToast('¡Modelo guardado con éxito!');
        } catch (err) {
            console.error(err);
            setSaveError('Error al guardar el modelo.');
        } finally {
            setSaving(false);
        }
    };

    if (loading && !areas.length) {
        return (
            <div className="text-center py-4">
                <Spinner animation="border" role="status">
                    <span className="visually-hidden">Cargando...</span>
                </Spinner>
            </div>
        );
    }
    if (loadError) {
        return (
            <Alert variant="danger">
                {loadError}{' '}
                <Button variant="link" className="p-0 align-baseline" onClick={cargarDatos}>Reintentar</Button>
            </Alert>
        );
    }

    const getProductoNombre = (productoId) => {
        const producto = productos.find(p => p.id === productoId);
        return producto ? `${producto.nombre} (${producto.unidad_medida})` : 'Producto no encontrado';
    };

    return (
        <Row>
            <Col md={4}>
                <h2 style={{ fontSize: '0.9375rem', fontWeight: 700, marginBottom: '12px' }}>Áreas</h2>
                <div className="list-card">
                    {areas.map(area => {
                        const AreaIcon = iconoDeArea(area.nombre);
                        const isActive = selectedArea?.id === area.id;
                        return (
                            <div
                                key={area.id}
                                className="list-row"
                                role="button"
                                tabIndex={0}
                                onClick={() => handleSelectArea(area)}
                                onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleSelectArea(area); }}
                                style={{
                                    cursor: 'pointer',
                                    gap: '10px',
                                    justifyContent: 'flex-start',
                                    backgroundColor: isActive ? 'var(--color-primary-soft)' : undefined,
                                    color: isActive ? 'var(--color-primary)' : undefined,
                                    fontWeight: isActive ? 600 : undefined,
                                }}
                            >
                                <AreaIcon size={16} />
                                {area.nombre}
                            </div>
                        );
                    })}
                </div>
            </Col>
            <Col md={8}>
                {selectedArea ? (
                    <div>
                        <h2 style={{ fontSize: '0.9375rem', fontWeight: 700, marginBottom: '12px' }}>
                            Productos para {selectedArea.nombre}
                        </h2>
                        {saveError && <Alert variant="danger" onClose={() => setSaveError('')} dismissible>{saveError}</Alert>}
                        <Form.Group>
                            <Typeahead
                                id="producto-typeahead"
                                options={productos}
                                labelKey={option => `${option.nombre} (${option.unidad_medida})`}
                                onChange={handleAddProducto}
                                placeholder="Escriba para buscar y agregar un producto..."
                                selected={[]}
                                positionFixed
                            />
                        </Form.Group>
                        <div className="list-card mt-3">
                            <table className="list-table">
                                <thead>
                                    <tr>
                                        <th>Producto</th>
                                        <th>Acciones</th>
                                    </tr>
                                </thead>
                                <DragDropContext onDragEnd={handleOnDragEnd}>
                                    <Droppable droppableId="productos">
                                        {(provided) => (
                                            <tbody {...provided.droppableProps} ref={provided.innerRef}>
                                                {(modelos[selectedArea.id] || []).length === 0 ? (
                                                    <tr>
                                                        <td colSpan="2" style={{ textAlign: 'center' }}>
                                                            Todavía no hay productos en este modelo.
                                                        </td>
                                                    </tr>
                                                ) : (
                                                    (modelos[selectedArea.id] || []).sort((a, b) => a.orden - b.orden).map((producto, index) => (
                                                        <Draggable key={`${selectedArea.id}-${producto.id}`} draggableId={`${selectedArea.id}-${producto.id}`} index={index}>
                                                            {(provided) => (
                                                                <tr
                                                                    ref={provided.innerRef}
                                                                    {...provided.draggableProps}
                                                                    {...provided.dragHandleProps}
                                                                >
                                                                    <td data-label="Producto" style={{ fontWeight: 600 }}>{getProductoNombre(producto.id)}</td>
                                                                    <td data-label="">
                                                                        <Button
                                                                            variant="outline-danger"
                                                                            size="sm"
                                                                            className="d-inline-flex align-items-center gap-1"
                                                                            onClick={() => handleRemoveProducto(producto.id)}
                                                                        >
                                                                            <TrashIcon size={13} /> Eliminar
                                                                        </Button>
                                                                    </td>
                                                                </tr>
                                                            )}
                                                        </Draggable>
                                                    ))
                                                )}
                                                {provided.placeholder}
                                            </tbody>
                                        )}
                                    </Droppable>
                                </DragDropContext>
                            </table>
                        </div>
                        <Button
                            variant="primary"
                            className="mt-3 d-inline-flex align-items-center gap-2"
                            onClick={handleSaveChanges}
                            disabled={saving}
                        >
                            {saving ? (
                                <>
                                    <Spinner animation="border" size="sm" />
                                    Guardando...
                                </>
                            ) : (
                                <>
                                    <CheckIcon size={15} /> Guardar Cambios
                                </>
                            )}
                        </Button>
                    </div>
                ) : (
                    <Alert variant="info">Seleccione un área para configurar su modelo.</Alert>
                )}
            </Col>
        </Row>
    );
};

export default ModeloIPV;
