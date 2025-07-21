import React, { useState, useEffect } from 'react';
import { Button, Container, Row, Col, ListGroup, Form, Alert, Spinner, Table } from 'react-bootstrap';
import { Typeahead } from 'react-bootstrap-typeahead';
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';
import 'react-bootstrap-typeahead/css/Typeahead.css';
import ipvApi from '../../api/ipvApi';
import * as productoApi from '../../api/productoApi';
import areaApi from '../../api/areaApi';

const ModeloIPV = () => {
    const [areas, setAreas] = useState([]);
    const [productos, setProductos] = useState([]);
    const [modelos, setModelos] = useState({});
    const [selectedArea, setSelectedArea] = useState(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        const cargarDatos = async () => {
            try {
                setLoading(true);
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
                setError('Error al cargar los datos iniciales.');
            } finally {
                setLoading(false);
            }
        };
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
        try {
            await ipvApi.guardarModelo({
                area_id: selectedArea.id,
                productos: modelos[selectedArea.id] || []
            });
            alert('¡Modelo guardado con éxito!');
        } catch (err) {
            setError('Error al guardar el modelo.');
        } finally {
            setSaving(false);
        }
    };

    if (loading && !areas.length) return <Spinner animation="border" />;
    if (error) return <Alert variant="danger">{error}</Alert>;

    const getProductoNombre = (productoId) => {
        const producto = productos.find(p => p.id === productoId);
        return producto ? `${producto.nombre} (${producto.unidad_medida})` : 'Producto no encontrado';
    };

    return (
        <Container fluid>
            <h3 className="my-4">Configurar Modelos de IPV por Área</h3>
            <Row>
                <Col md={4}>
                    <h4>Áreas</h4>
                    <ListGroup>
                        {areas.map(area => (
                            <ListGroup.Item 
                                key={area.id} 
                                action 
                                active={selectedArea?.id === area.id}
                                onClick={() => handleSelectArea(area)}
                            >
                                {area.nombre}
                            </ListGroup.Item>
                        ))}
                    </ListGroup>
                </Col>
                <Col md={8}>
                    {selectedArea ? (
                        <div>
                            <h4>Productos para {selectedArea.nombre}</h4>
                            <Form.Group>
                                <Typeahead
                                    id="producto-typeahead"
                                    options={productos}
                                    labelKey={option => `${option.nombre} (${option.unidad_medida})`}
                                    onChange={handleAddProducto}
                                    placeholder="Escriba para buscar y agregar un producto..."
                                    selected={[]}
                                />
                            </Form.Group>
                            <Table striped bordered hover className="mt-3">
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
                                                {(modelos[selectedArea.id] || []).sort((a, b) => a.orden - b.orden).map((producto, index) => (
                                                    <Draggable key={`${selectedArea.id}-${producto.id}`} draggableId={`${selectedArea.id}-${producto.id}`} index={index}>
                                                        {(provided) => (
                                                            <tr
                                                                ref={provided.innerRef}
                                                                {...provided.draggableProps}
                                                                {...provided.dragHandleProps}
                                                            >
                                                                <td>{getProductoNombre(producto.id)}</td>
                                                                <td>
                                                                    <Button
                                                                        variant="danger"
                                                                        size="sm"
                                                                        onClick={() => handleRemoveProducto(producto.id)}
                                                                    >
                                                                        Eliminar
                                                                    </Button>
                                                                </td>
                                                            </tr>
                                                        )}
                                                    </Draggable>
                                                ))}
                                                {provided.placeholder}
                                            </tbody>
                                        )}
                                    </Droppable>
                                </DragDropContext>
                            </Table>
                            <Button className="mt-3" onClick={handleSaveChanges} disabled={saving}>
                                {saving ? 'Guardando...' : 'Guardar Cambios'}
                            </Button>
                        </div>
                    ) : (
                        <Alert variant="info">Seleccione un área para configurar su modelo.</Alert>
                    )}
                </Col>
            </Row>
        </Container>
    );
};

export default ModeloIPV;
